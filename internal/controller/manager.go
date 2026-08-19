package controller

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/statestore"
)

// Manager is the native Tarak controller manager.
// It watches and reconciles Deployments, Pods, Services, and MetalLB-style LoadBalancers.
type Manager struct {
	store   statestore.Store
	runtime runtime.Runtime
	log      *zap.Logger
	mu       sync.Mutex
	backoffs map[string]time.Time
}

// NewManager constructs a new controller manager.
func NewManager(store statestore.Store, rt runtime.Runtime, log *zap.Logger) *Manager {
	if log == nil {
		log = zap.NewNop()
	}
	namedLog := log.Named("controller")
	if rt == nil {
		rt = runtime.NewEngine("", namedLog)
	}
	return &Manager{
		store:    store,
		runtime:  rt,
		log:      namedLog,
		backoffs: make(map[string]time.Time),
	}
}

// Start launches the background reconciliation loop.
func (m *Manager) Start(ctx context.Context) {
	m.log.Info("starting Tarak native control loops (Deployments, Pods, Services, MetalLB)")
	go m.runLoop(ctx)
}

func (m *Manager) runLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
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

		// Create missing pods
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
						m.log.Info("deployment controller created pod", zap.String("deployment", name), zap.String("pod", podName))
					}
				}
			}
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

		// Update Deployment Status
		nowStr := time.Now().UTC().Format(time.RFC3339)
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

			rawContainers, _ := spec["containers"].([]interface{})
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

				// Real / Emulated Image Pull
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
			updatedStatusRaw, _ := json.Marshal(pod["status"])
			existingStatusRaw, _ := json.Marshal(existingStatus)
			if string(updatedStatusRaw) != string(existingStatusRaw) {
				needsUpdate = true
			}
			if needsUpdate {
				m.log.Info("tarak node runtime reconciled pod", zap.String("pod", name), zap.String("phase", podPhase))
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
				m.log.Info("pod controller: updated pod status -> Running",
					zap.String("pod", name),
					zap.String("ns", ns),
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
			spec["type"] = svcType
		}

		needsUpdate := false

		// 1. Assign ClusterIP if missing
		clusterIP, _ := spec["clusterIP"].(string)
		if clusterIP == "" && clusterIP != "None" {
			spec["clusterIP"] = fmt.Sprintf("10.96.0.%d", (idx%240)+10)
			needsUpdate = true
		}

		// 2. Assign nodePorts for NodePort & LoadBalancer services
		ports, _ := spec["ports"].([]interface{})
		for pIdx, p := range ports {
			pMap, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			if pMap["protocol"] == nil || pMap["protocol"] == "" {
				pMap["protocol"] = "TCP"
				needsUpdate = true
			}
			if svcType == "NodePort" || svcType == "LoadBalancer" {
				if np, _ := pMap["nodePort"].(float64); np == 0 {
					pMap["nodePort"] = 30000 + (idx*10) + pIdx + 120
					needsUpdate = true
				}
			}
		}

		// 3. MetalLB / Public IP LoadBalancer allocation
		if svcType == "LoadBalancer" {
			targetIP := "192.168.1.240"
			if lbIP, _ := spec["loadBalancerIP"].(string); lbIP != "" {
				targetIP = lbIP
			}

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
				m.log.Info("tarak metallb controller assigned public IP to service", zap.String("service", name), zap.String("externalIP", targetIP))
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
