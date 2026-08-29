package controller

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/loadbalancer"
	"github.com/vikukumar/tarak/internal/network"
	"github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/statestore"
)

// Manager is the native Tarak controller manager.
// It watches and reconciles Deployments, Pods, Services, MetalLB, and Ingress routes.
type Manager struct {
	store       statestore.Store
	runtime     runtime.Runtime
	lbCtrl      *loadbalancer.Controller
	netDriver   *network.Driver
	log         *zap.Logger
	mu          sync.Mutex
	backoffs    map[string]time.Time
	ingressFunc func(ctx context.Context)
}

// NewManager constructs a new controller manager.
func NewManager(store statestore.Store, rt runtime.Runtime, lbCtrl *loadbalancer.Controller, netDriver *network.Driver, log *zap.Logger) *Manager {
	if log == nil {
		log = zap.NewNop()
	}
	namedLog := log.Named("controller")
	if rt == nil {
		rt = runtime.NewEngine("", namedLog)
	}
	return &Manager{
		store:     store,
		runtime:   rt,
		lbCtrl:    lbCtrl,
		netDriver: netDriver,
		log:       namedLog,
		backoffs:  make(map[string]time.Time),
	}
}

// SetIngressReconciler attaches an external Ingress reconciler callback.
func (m *Manager) SetIngressReconciler(fn func(ctx context.Context)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ingressFunc = fn
}

// Start launches the background reconciliation loop.
func (m *Manager) Start(ctx context.Context) {
	m.log.Info("starting Tarak native control loops (Deployments, Pods, Services, Ingress, MetalLB)")
	go m.runLoop(ctx)
}

func (m *Manager) runLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 10s is sufficient; 1s causes noisy hot loops
	defer ticker.Stop()

	// Run initial reconciliation immediately
	m.reconcileAll(ctx)

	for {
		select {
		case <-ctx.Done():
			m.log.Info("stopping Tarak controller manager")
			return
		case <-ticker.C:
			m.reconcileAll(ctx)
		}
	}
}

func (m *Manager) reconcileAll(ctx context.Context) {
	m.reconcileDeployments(ctx, "apps", "v1")
	m.reconcileDeployments(ctx, "apps.tarak.io", "v1")
	m.reconcilePods(ctx)
	m.reconcileServices(ctx, "", "v1") // services are stored under core group ""

	m.mu.Lock()
	fn := m.ingressFunc
	m.mu.Unlock()
	if fn != nil {
		fn(ctx)
	}
}

// ─── Deployment Controller ───────────────────────────────────────────────────

func (m *Manager) reconcileDeployments(ctx context.Context, group, version string) {
	envs, _, err := m.store.List(ctx, statestore.ListQuery{
		Key: statestore.ResourceKey{
			Group:    group,
			Version:  version,
			Resource: "deployments",
		},
	})
	if err != nil || len(envs) == 0 {
		return
	}

	for _, env := range envs {
		var deploy map[string]interface{}
		if err := json.Unmarshal(env.Object, &deploy); err != nil {
			continue
		}

		meta, _ := deploy["metadata"].(map[string]interface{})
		spec, _ := deploy["spec"].(map[string]interface{})
		if meta == nil || spec == nil {
			continue
		}

		name, _ := meta["name"].(string)
		ns, _ := meta["namespace"].(string)
		if ns == "" {
			ns = "default"
		}
		uid, _ := meta["uid"].(string)

		replicas := int32(1)
		if r, ok := spec["replicas"].(float64); ok && r >= 0 {
			replicas = int32(r)
		}

		template, _ := spec["template"].(map[string]interface{})
		if template == nil {
			continue
		}
		tmplMeta, _ := template["metadata"].(map[string]interface{})
		tmplLabels, _ := tmplMeta["labels"].(map[string]interface{})
		tmplSpec, _ := template["spec"].(map[string]interface{})
		if tmplSpec == nil {
			continue
		}

		// Calculate pod template hash
		h := sha256.New()
		h.Write([]byte(name))
		hash := hex.EncodeToString(h.Sum(nil))[:9]

		// List existing pods in this namespace
		podEnvs, _, err := m.store.List(ctx, statestore.ListQuery{
			Key: statestore.ResourceKey{
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
				Namespace: ns,
			},
		})
		if err != nil {
			continue
		}

		var matchingPods []*statestore.Envelope
		for _, pe := range podEnvs {
			var podObj map[string]interface{}
			if err := json.Unmarshal(pe.Object, &podObj); err != nil {
				continue
			}
			pMeta, _ := podObj["metadata"].(map[string]interface{})
			if pMeta == nil {
				continue
			}
			pName, _ := pMeta["name"].(string)
			if strings.HasPrefix(pName, fmt.Sprintf("%s-%s-", name, hash)) || strings.HasPrefix(pName, fmt.Sprintf("%s-", name)) {
				matchingPods = append(matchingPods, pe)
			}
		}

		// Create missing pods (Scale Up)
		currentCount := int32(len(matchingPods))
		if currentCount < replicas {
			for i := currentCount; i < replicas; i++ {
				podName := fmt.Sprintf("%s-%s-%s", name, hash, randomSuffix(5))
				mergedLabels := map[string]interface{}{
					"app.kubernetes.io/managed-by": "tarak-deployment-controller",
					"pod-template-hash":            hash,
				}
				for k, v := range tmplLabels {
					mergedLabels[k] = v
				}

				podObj := map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      podName,
						"namespace": ns,
						"labels":    mergedLabels,
						"ownerReferences": []map[string]interface{}{
							{
								"apiVersion": group + "/" + version,
								"kind":       "Deployment",
								"name":       name,
								"uid":        uid,
								"controller": true,
							},
						},
					},
					"spec": tmplSpec,
				}

				podRaw, err := json.Marshal(podObj)
				if err == nil {
					key := statestore.ResourceKey{
						Group:     "",
						Version:   "v1",
						Resource:  "pods",
						Namespace: ns,
						Name:      podName,
					}
					if createdEnv, err := m.store.Create(ctx, key, podRaw); err == nil {
						matchingPods = append(matchingPods, createdEnv)
						m.log.Info("deployment controller created pod (scaled up)", zap.String("deployment", name), zap.String("pod", podName))
					}
				}
			}
		} else if currentCount > replicas {
			// Descale / Delete excess pods (Scale Down)
			excess := currentCount - replicas
			for i := int32(0); i < excess && i < int32(len(matchingPods)); i++ {
				idx := len(matchingPods) - 1 - int(i)
				pe := matchingPods[idx]
				var pObj map[string]interface{}
				_ = json.Unmarshal(pe.Object, &pObj)
				pMeta, _ := pObj["metadata"].(map[string]interface{})
				pName, _ := pMeta["name"].(string)

				pKey := statestore.ResourceKey{
					Group:     "",
					Version:   "v1",
					Resource:  "pods",
					Namespace: ns,
					Name:      pName,
				}
				_, _ = m.store.Delete(ctx, pKey, 0)
				if m.runtime != nil {
					_ = m.runtime.StopPodContainers(ctx, ns, pName)
				}
				m.log.Info("deployment controller deleted excess pod (scaled down)", zap.String("deployment", name), zap.String("pod", pName))
			}
			matchingPods = matchingPods[:replicas]
		}

		// Count ready pods
		readyCount := int32(0)
		for _, pe := range matchingPods {
			var podObj map[string]interface{}
			if err := json.Unmarshal(pe.Object, &podObj); err != nil {
				continue
			}
			st, _ := podObj["status"].(map[string]interface{})
			if st != nil && st["phase"] == "Running" {
				readyCount++
			}
		}

		// Create / Sync ReplicaSet for this deployment revision
		nowStr := time.Now().UTC().Format(time.RFC3339)
		rsName := fmt.Sprintf("%s-%s", name, hash)
		rsLabels := map[string]interface{}{
			"pod-template-hash": hash,
		}
		for k, v := range tmplLabels {
			rsLabels[k] = v
		}

		rsObj := map[string]interface{}{
			"apiVersion": group + "/" + version,
			"kind":       "ReplicaSet",
			"metadata": map[string]interface{}{
				"name":      rsName,
				"namespace": ns,
				"labels":    rsLabels,
				"ownerReferences": []map[string]interface{}{
					{
						"apiVersion": group + "/" + version,
						"kind":       "Deployment",
						"name":       name,
						"uid":        uid,
						"controller": true,
					},
				},
				"creationTimestamp": nowStr,
			},
			"spec": map[string]interface{}{
				"replicas": replicas,
				"selector": map[string]interface{}{
					"matchLabels": tmplLabels,
				},
				"template": template,
			},
			"status": map[string]interface{}{
				"replicas":          currentCount,
				"readyReplicas":     readyCount,
				"availableReplicas": readyCount,
				"observedGeneration": 1,
			},
		}

		if rsBytes, mErr := json.Marshal(rsObj); mErr == nil {
			rsKey := statestore.ResourceKey{
				Group:     group,
				Version:   version,
				Resource:  "replicasets",
				Namespace: ns,
				Name:      rsName,
			}
			if _, getErr := m.store.Get(ctx, rsKey); getErr != nil {
				_, _ = m.store.Create(ctx, rsKey, rsBytes)
			} else {
				_, _ = m.store.Update(ctx, rsKey, rsBytes, 0)
			}
		}

		// Update Deployment Status
		nowStr = time.Now().UTC().Format(time.RFC3339)
		deployStatus := map[string]interface{}{
			"replicas":            replicas,
			"readyReplicas":       readyCount,
			"availableReplicas":   readyCount,
			"updatedReplicas":     replicas,
			"observedGeneration":  1,
			"conditions": []map[string]interface{}{
				{
					"type":               "Available",
					"status":             "True",
					"lastUpdateTime":     nowStr,
					"lastTransitionTime": nowStr,
					"reason":             "MinimumReplicasAvailable",
					"message":            "Deployment has minimum availability.",
				},
				{
					"type":               "Progressing",
					"status":             "True",
					"lastUpdateTime":     nowStr,
					"lastTransitionTime": nowStr,
					"reason":             "NewReplicaSetAvailable",
					"message":            fmt.Sprintf("ReplicaSet %q has successfully progressed.", name+"-"+hash),
				},
			},
		}

		deploy["status"] = deployStatus
		updatedDeployRaw, err := json.Marshal(deploy)
		if err == nil {
			key := statestore.ResourceKey{
				Group:     group,
				Version:   version,
				Resource:  "deployments",
				Namespace: ns,
				Name:      name,
			}
			if _, err := m.store.Update(ctx, key, updatedDeployRaw, 0); err != nil {
				m.log.Warn("deployment controller: failed to update deployment status", zap.String("name", name), zap.Error(err))
			}
		}
	}
}

// ─── Pod & Scheduler Controller ──────────────────────────────────────────────

func (m *Manager) reconcilePods(ctx context.Context) {
	// Find available node
	nodeName := "tarak-control-plane"
	nodeEnvs, _, err := m.store.List(ctx, statestore.ListQuery{
		Key: statestore.ResourceKey{Group: "", Version: "v1", Resource: "nodes"},
	})
	if err == nil && len(nodeEnvs) > 0 {
		var nObj map[string]interface{}
		if err := json.Unmarshal(nodeEnvs[0].Object, &nObj); err == nil {
			if nMeta, _ := nObj["metadata"].(map[string]interface{}); nMeta != nil {
				if nm, _ := nMeta["name"].(string); nm != "" {
					nodeName = nm
				}
			}
		}
	}

	podEnvs, _, listErr := m.store.List(ctx, statestore.ListQuery{
		Key: statestore.ResourceKey{Group: "", Version: "v1", Resource: "pods"},
	})
	if listErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[controller] FATAL: pod list failed: %v\n", listErr)
		return
	}

	activePods := make(map[string]bool)
	for _, env := range podEnvs {
		var pod map[string]interface{}
		if err := json.Unmarshal(env.Object, &pod); err == nil {
			if meta, _ := pod["metadata"].(map[string]interface{}); meta != nil {
				nm, _ := meta["name"].(string)
				ns, _ := meta["namespace"].(string)
				if ns == "" {
					ns = "default"
				}
				if nm != "" {
					activePods[ns+"/"+nm] = true
				}
			}
		}
	}

	// Always cleanup containers for pods that were deleted from statestore
	if m.runtime != nil {
		m.runtime.CleanupDeletedPods(ctx, activePods)
	}

	if len(podEnvs) == 0 {
		return
	}

	for idx, env := range podEnvs {
		var pod map[string]interface{}
		if err := json.Unmarshal(env.Object, &pod); err != nil {
			continue
		}

		meta, _ := pod["metadata"].(map[string]interface{})
		spec, _ := pod["spec"].(map[string]interface{})
		if meta == nil || spec == nil {
			continue
		}

		name, _ := meta["name"].(string)
		ns, _ := meta["namespace"].(string)
		if ns == "" {
			ns = "default"
		}

		podKey := ns + "/" + name
		m.mu.Lock()
		backoffUntil, hasBackoff := m.backoffs[podKey]
		m.mu.Unlock()
		if hasBackoff && time.Now().Before(backoffUntil) {
			continue
		}

		needsUpdate := false

		// Scheduler: bind to node
		currNode, _ := spec["nodeName"].(string)
		if currNode == "" {
			spec["nodeName"] = nodeName
			needsUpdate = true
			m.log.Info("tarak scheduler assigned pod to node", zap.String("pod", name), zap.String("node", nodeName))
			m.recordEvent(ctx, ns, "Pod", name, "Scheduled", fmt.Sprintf("Successfully assigned %s/%s to %s", ns, name, nodeName), "tarak-scheduler", "Normal")
		}

		// Pod runtime lifecycle — always reconcile so crashes are detected
		existingStatus, _ := pod["status"].(map[string]interface{})
		existingPhase, _ := existingStatus["phase"].(string)
		podIP := fmt.Sprintf("10.244.0.%d", (idx%240)+2)
		if existingStatus != nil {
			if ip, ok := existingStatus["podIP"].(string); ok && ip != "" {
				podIP = ip
			}
		}
		now := time.Now().UTC()
		nowStr := now.Format(time.RFC3339)
		_ = existingPhase // used below

		{

			// ──────────────────────────────────────────────────────────────────
			// Skip image pull and container start when the pod is already Running
			// and all containers are confirmed healthy in the runtime.
			// This prevents the hot reconcile loop that spams PullImage every second.
			// ──────────────────────────────────────────────────────────────────
			rawContainers, _ := spec["containers"].([]interface{})
			podAlreadyRunning := existingPhase == "Running"
			if podAlreadyRunning && m.runtime != nil {
				// Verify every container is still alive in the runtime
				allHealthy := true
				for _, c := range rawContainers {
					cMap, _ := c.(map[string]interface{})
					if cMap == nil {
						continue
					}
					cName, _ := cMap["name"].(string)
					info, err := m.runtime.GetContainerInfo(ctx, ns, name, cName)
					if err != nil || info == nil || info.State != runtime.StateRunning {
						allHealthy = false
						break
					}
				}
				if allHealthy {
					// Pod is healthy — nothing to do this cycle
					continue
				}
				// Container(s) crashed — fall through to restart logic
			}

			var cSpecs []runtime.ContainerSpec
			anyPullFailed := false
			for _, c := range rawContainers {

				cMap, _ := c.(map[string]interface{})
				if cMap == nil {
					continue
				}
				cName, _ := cMap["name"].(string)
				cImg, _ := cMap["image"].(string)
				if cImg == "" {
					cImg = "tarak-runtime/app:latest"
				}

				var ports []runtime.ContainerPort
				if pList, ok := cMap["ports"].([]interface{}); ok {
					for _, p := range pList {
						if pMap, ok := p.(map[string]interface{}); ok {
							cp, _ := pMap["containerPort"].(float64)
							hp, _ := pMap["hostPort"].(float64)
							pName, _ := pMap["name"].(string)
							proto, _ := pMap["protocol"].(string)
							ports = append(ports, runtime.ContainerPort{
								Name:          pName,
								ContainerPort: int(cp),
								HostPort:      int(hp),
								Protocol:      proto,
							})
						}
					}
				}

				envMap := make(map[string]string)
				if eList, ok := cMap["env"].([]interface{}); ok {
					for _, e := range eList {
						if eMap, ok := e.(map[string]interface{}); ok {
							k, _ := eMap["name"].(string)
							v, _ := eMap["value"].(string)
							if k != "" {
								envMap[k] = v
							}
						}
					}
				}

				cSpecs = append(cSpecs, runtime.ContainerSpec{
					Name:  cName,
					Image: cImg,
					Ports: ports,
					Env:   envMap,
				})

				// Pull image only when the pod is not already running
				m.recordEvent(ctx, ns, "Pod", name, "Pulling", fmt.Sprintf("Pulling image %q", cImg), "tarak-runtime", "Normal")
				pullRes, pullErr := m.runtime.PullImage(ctx, cImg)
				if pullErr != nil {
					m.recordEvent(ctx, ns, "Pod", name, "Failed", fmt.Sprintf("Failed to pull image %q: %v", cImg, pullErr), "tarak-runtime", "Warning")
					anyPullFailed = true
				} else {
					dur := 0.8
					if pullRes != nil && pullRes.Duration > 0 {
						dur = pullRes.Duration.Seconds()
					}
					m.recordEvent(ctx, ns, "Pod", name, "Pulled", fmt.Sprintf("Successfully pulled image %q in %.2fs", cImg, dur), "tarak-runtime", "Normal")
				}
			}

			// Run containers only if pulls succeeded
			runtimeSpec := runtime.PodRuntimeSpec{
				Namespace:  ns,
				PodName:    name,
				Containers: cSpecs,
			}
			
			var cInfos map[string]*runtime.ContainerInfo
			if !anyPullFailed {
				cInfos, _ = m.runtime.RunPodContainers(ctx, runtimeSpec)
			} else {
				cInfos = make(map[string]*runtime.ContainerInfo)
			}

			var cStatuses []map[string]interface{}
			podPhase := "Running"
			podReady := "True"
			
			for _, cs := range cSpecs {
				cID := fmt.Sprintf("tarak://%s", randomSuffix(16))
				imgID := fmt.Sprintf("docker-pullable://%s@sha256:%s", cs.Image, randomSuffix(16))
				isReady := true
				stateObj := map[string]interface{}{
					"running": map[string]interface{}{
						"startedAt": nowStr,
					},
				}

				if info, ok := cInfos[cs.Name]; ok && info != nil {
					if info.ContainerID != "" {
						cID = info.ContainerID
					}
					if info.ImageID != "" {
						imgID = info.ImageID
					}
					switch info.State {
					case runtime.StateRunning:
						if existingPhase != "Running" {
							m.recordEvent(ctx, ns, "Pod", name, "Started", fmt.Sprintf("Started container %s", cs.Name), "tarak-runtime", "Normal")
						}
					case runtime.StateError, runtime.StateTerminated:
						isReady = false
						podPhase = "Failed"
						podReady = "False"
						stateObj = map[string]interface{}{
							"terminated": map[string]interface{}{
								"exitCode": 1,
								"reason":   "Error",
								"message":  "Container failed to start",
							},
						}
						if existingPhase != "Failed" {
							m.recordEvent(ctx, ns, "Pod", name, "Failed", fmt.Sprintf("Container %s failed to start", cs.Name), "tarak-runtime", "Warning")
						}
					case runtime.StatePending, runtime.StateContainerCreating:
						isReady = false
						podPhase = "Pending"
						podReady = "False"
						stateObj = map[string]interface{}{
							"waiting": map[string]interface{}{
								"reason": "ContainerCreating",
							},
						}
					}
				} else {
					isReady = false
					podPhase = "Pending"
					podReady = "False"
					stateObj = map[string]interface{}{
						"waiting": map[string]interface{}{
							"reason": "ImagePullBackOff",
						},
					}
					
					m.mu.Lock()
					m.backoffs[podKey] = time.Now().Add(10 * time.Second)
					m.mu.Unlock()
				}

				cStatuses = append(cStatuses, map[string]interface{}{
					"name":         cs.Name,
					"image":        cs.Image,
					"ready":        isReady,
					"restartCount": 0,
					"imageID":      imgID,
					"containerID":  cID,
					"state":        stateObj,
				})
			}

			pod["status"] = map[string]interface{}{
				"phase":             podPhase,
				"hostIP":            "127.0.0.1",
				"podIP":             podIP,
				"podIPs":            []map[string]interface{}{{"ip": podIP}},
				"startTime":         nowStr,
				"containerStatuses": cStatuses,
				"conditions": []map[string]interface{}{
					{"type": "Initialized", "status": "True", "lastTransitionTime": nowStr},
					{"type": "Ready", "status": podReady, "lastTransitionTime": nowStr},
					{"type": "ContainersReady", "status": podReady, "lastTransitionTime": nowStr},
					{"type": "PodScheduled", "status": "True", "lastTransitionTime": nowStr},
				},
			}
			// Compare stable status fields only — exclude dynamic timestamps to avoid
			// infinite update loops where nowStr differs every reconcile cycle.
			updatedPhaseStr, _ := pod["status"].(map[string]interface{})["phase"].(string)
			existingPhaseStr, _ := existingStatus["phase"].(string)
			updatedReadyStatus := ""
			if conds, ok := pod["status"].(map[string]interface{})["conditions"].([]map[string]interface{}); ok {
				for _, c := range conds {
					if c["type"] == "Ready" {
						updatedReadyStatus, _ = c["status"].(string)
						break
					}
				}
			}
			phaseChanged := updatedPhaseStr != existingPhaseStr
			if phaseChanged {
				needsUpdate = true
			}
			_ = updatedReadyStatus
			if needsUpdate {
				m.log.Info("tarak node runtime reconciled pod", zap.String("pod", name), zap.String("phase", updatedPhaseStr))
			}
		}

		if needsUpdate {
			updatedRaw, err := json.Marshal(pod)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[controller] marshal pod %s error: %v\n", name, err)
				m.log.Error("pod controller: failed to marshal pod", zap.String("pod", name), zap.Error(err))
				continue
			}
			key := statestore.ResourceKey{
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
				Namespace: ns,
				Name:      name,
			}
		// Extract phase for logging (podPhase is scoped to inner block above)
		updatedPhase := ""
		if updSt, ok := pod["status"].(map[string]interface{}); ok {
			updatedPhase, _ = updSt["phase"].(string)
		}
		_, _ = fmt.Fprintf(os.Stderr, "[controller] updating pod %s/%s status -> %s\n", ns, name, updatedPhase)
			if _, updateErr := m.store.Update(ctx, key, updatedRaw, 0); updateErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[controller] UPDATE FAILED pod %s/%s: %v\n", ns, name, updateErr)
				m.log.Error("pod controller: failed to update pod status",
					zap.String("pod", name),
					zap.String("ns", ns),
					zap.Error(updateErr),
				)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "[controller] UPDATE OK pod %s/%s\n", ns, name)
				updPhase := ""
				if updSt, ok2 := pod["status"].(map[string]interface{}); ok2 {
					updPhase, _ = updSt["phase"].(string)
				}
				m.log.Info("pod controller: updated pod status",
					zap.String("pod", name),
					zap.String("ns", ns),
					zap.String("phase", updPhase),
				)
			}
		}
	}
}

func (m *Manager) recordEvent(ctx context.Context, ns, kind, name, reason, message, component, evType string) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	evName := fmt.Sprintf("%s.%x", name, now.UnixNano())

	evObj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Event",
		"metadata": map[string]interface{}{
			"name":              evName,
			"namespace":         ns,
			"creationTimestamp": nowStr,
		},
		"involvedObject": map[string]interface{}{
			"apiVersion": "v1",
			"kind":       kind,
			"name":       name,
			"namespace":  ns,
		},
		"reason":         reason,
		"message":        message,
		"source": map[string]interface{}{
			"component": component,
		},
		"firstTimestamp": nowStr,
		"lastTimestamp":  nowStr,
		"count":          1,
		"type":           evType,
	}

	raw, err := json.Marshal(evObj)
	if err != nil {
		return
	}
	key := statestore.ResourceKey{
		Group:     "",
		Version:   "v1",
		Resource:  "events",
		Namespace: ns,
		Name:      evName,
	}
	_, _ = m.store.Create(ctx, key, raw)
}

// ─── Service & MetalLB Controller ─────────────────────────────────────────────

func (m *Manager) reconcileServices(ctx context.Context, group, version string) {
	desiredRoutes := make(map[string][]loadbalancer.Endpoint)
	defer func() {
		if m.lbCtrl != nil {
			m.lbCtrl.SyncAllServiceForwarding(ctx, desiredRoutes)
		}
	}()

	svcEnvs, _, err := m.store.List(ctx, statestore.ListQuery{
		Key: statestore.ResourceKey{Group: group, Version: version, Resource: "services"},
	})
	if err != nil || len(svcEnvs) == 0 {
		return
	}

	for idx, env := range svcEnvs {
		var svc map[string]interface{}
		if err := json.Unmarshal(env.Object, &svc); err != nil {
			continue
		}

		meta, _ := svc["metadata"].(map[string]interface{})
		spec, _ := svc["spec"].(map[string]interface{})
		if meta == nil || spec == nil {
			continue
		}

		name, _ := meta["name"].(string)
		ns, _ := meta["namespace"].(string)
		if ns == "" {
			ns = "default"
		}

		svcType, _ := spec["type"].(string)
		if svcType == "" {
			svcType = "ClusterIP"
		}

		needsUpdate := false

		// 1. ClusterIP Assignment
		currClusterIP, _ := spec["clusterIP"].(string)
		if currClusterIP == "" {
			spec["clusterIP"] = fmt.Sprintf("10.96.0.%d", (idx%240)+10)
			needsUpdate = true
			m.log.Info("tarak service controller assigned clusterIP", zap.String("service", name), zap.String("clusterIP", spec["clusterIP"].(string)))
		}

		// 2. NodePort Assignment
		ports, _ := spec["ports"].([]interface{})
		for pIdx, p := range ports {
			pMap, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			currNP, _ := pMap["nodePort"].(float64)
			if currNP == 0 && (svcType == "NodePort" || svcType == "LoadBalancer") {
				pMap["nodePort"] = 30000 + float64(pIdx*10) + float64((idx%50)*100) + 1
				ports[pIdx] = pMap
				needsUpdate = true
			}
		}
		if needsUpdate {
			spec["ports"] = ports
		}

		// 3. Endpoint Resolution & Dynamic IP Allocation
		var endpoints []loadbalancer.Endpoint
		selector, _ := spec["selector"].(map[string]interface{})
		if selector != nil {
			podEnvs, _, _ := m.store.List(ctx, statestore.ListQuery{
				Key: statestore.ResourceKey{
					Group:     "",
					Version:   "v1",
					Resource:  "pods",
					Namespace: ns,
				},
			})
			for _, pEnv := range podEnvs {
				var pObj map[string]interface{}
				if err := json.Unmarshal(pEnv.Object, &pObj); err != nil {
					continue
				}
				pMeta, _ := pObj["metadata"].(map[string]interface{})
				pLabels, _ := pMeta["labels"].(map[string]interface{})
				match := true
				for k, v := range selector {
					if pLabels[k] != v {
						match = false
						break
					}
				}
				if match {
					pStatus, _ := pObj["status"].(map[string]interface{})
					podPhase, _ := pStatus["phase"].(string)
					if podPhase != "" && podPhase != "Running" && podPhase != "Pending" {
						continue
					}
					podIP, _ := pStatus["podIP"].(string)
					if podIP == "" {
						podIP = "127.0.0.1"
					}
					pName, _ := pMeta["name"].(string)

					// Find container target port
					pSpec, _ := pObj["spec"].(map[string]interface{})
					pContainers, _ := pSpec["containers"].([]interface{})
					tPort := 80
					for _, c := range pContainers {
						cMap, _ := c.(map[string]interface{})
						cPorts, _ := cMap["ports"].([]interface{})
						for _, cp := range cPorts {
							cpMap, _ := cp.(map[string]interface{})
							if portVal, ok := cpMap["containerPort"].(float64); ok && portVal > 0 {
								tPort = int(portVal)
								break
							}
						}
					}

					// Resolve real active host port for this pod container
					realPort := tPort
					if m.runtime != nil {
						realPort = m.runtime.GetHostPort(ns, pName, tPort)
					}

					endpoints = append(endpoints, loadbalancer.Endpoint{
						Address: "127.0.0.1",
						Port:    realPort,
						Healthy: true,
					})
					if m.netDriver != nil && podIP != "127.0.0.1" {
						m.netDriver.RegisterPodRoute(podIP, fmt.Sprintf("127.0.0.1:%d", realPort))
					}
				}
			}
		}

		if svcType == "LoadBalancer" {
			targetIP := m.resolveServiceExternalIP(ctx, spec, meta)

			status, _ := svc["status"].(map[string]interface{})
			if status == nil {
				status = map[string]interface{}{}
			}
			lbStatus, _ := status["loadBalancer"].(map[string]interface{})
			var ingressLen int
			if lbStatus != nil {
				if ingList, ok := lbStatus["ingress"].([]interface{}); ok {
					ingressLen = len(ingList)
				}
			}
			if lbStatus == nil || ingressLen == 0 {
				status["loadBalancer"] = map[string]interface{}{
					"ingress": []map[string]interface{}{
						{
							"ip":       targetIP,
							"ipMode":   "VIP",
							"hostname": "lb.tarak.local",
						},
					},
				}
				svc["status"] = status
				needsUpdate = true
				m.log.Info("tarak metallb controller dynamically assigned IP to service", zap.String("service", name), zap.String("externalIP", targetIP))
			}
		}

		// 4. Live TCP Proxy Forwarding to Host
		if len(endpoints) > 0 {
			for _, p := range ports {
				pMap, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				svcPort, _ := pMap["port"].(float64)
				np, _ := pMap["nodePort"].(float64)

				if np > 0 {
					listenAddr := fmt.Sprintf("0.0.0.0:%d", int(np))
					desiredRoutes[listenAddr] = endpoints
				}
				if svcType == "LoadBalancer" && svcPort > 0 {
					listenAddr := fmt.Sprintf("0.0.0.0:%d", int(svcPort))
					desiredRoutes[listenAddr] = endpoints
				}
			}
		}

		if needsUpdate {
			updatedRaw, err := json.Marshal(svc)
			if err == nil {
				key := statestore.ResourceKey{
					Group:     group,
					Version:   version,
					Resource:  "services",
					Namespace: ns,
					Name:      name,
				}
				_, _ = m.store.Update(ctx, key, updatedRaw, 0)
			}
		}
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func randomSuffix(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, length)
	for i, v := range b {
		out[i] = chars[int(v)%len(chars)]
	}
	return string(out)
}

// resolveServiceExternalIP dynamically detects the IP for LoadBalancer services from spec, annotations, network driver, or host interfaces.
func (m *Manager) resolveServiceExternalIP(ctx context.Context, spec, meta map[string]interface{}) string {
	// 1. Explicit user request in spec.loadBalancerIP
	if lbIP, _ := spec["loadBalancerIP"].(string); lbIP != "" {
		return lbIP
	}

	// 2. Explicit user request in annotations
	if annotations, ok := meta["annotations"].(map[string]interface{}); ok {
		if annIP, ok := annotations["metallb.universe.tf/loadBalancerIPs"].(string); ok && annIP != "" {
			return annIP
		}
		if annIP, ok := annotations["tarak.io/external-ip"].(string); ok && annIP != "" {
			return annIP
		}
	}

	// 3. Dynamic query from live network bridge driver (Primary LAN IP / Public IP)
	if m.netDriver != nil {
		netInfo := m.netDriver.GetHostNetworkInfo()
		if netInfo.PrimaryLANIP != "" && netInfo.PrimaryLANIP != "127.0.0.1" {
			return netInfo.PrimaryLANIP
		}
		if netInfo.PublicIP != "" && netInfo.PublicIP != "127.0.0.1" {
			return netInfo.PublicIP
		}
	}

	// 4. Dynamic query from loadbalancer controller detector / pool
	if m.lbCtrl != nil {
		if pub := m.lbCtrl.PublicIP(); pub != "" && pub != "127.0.0.1" {
			return pub
		}
	}

	// 5. Dynamic fallback: probe host network interfaces for first active IPv4
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
					return ipNet.IP.String()
				}
			}
		}
	}

	return "127.0.0.1"
}
