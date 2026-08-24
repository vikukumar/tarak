// Package controller implements the native Tarak micro-Kubelet node manager.
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/hardware"
	"github.com/vikukumar/tarak/internal/network"
	tarakruntime "github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/runtime/tcr"
	"github.com/vikukumar/tarak/internal/statestore"
)

// MicroKubelet is the self-contained, autonomous node agent running on every Tarak node.
type MicroKubelet struct {
	mu         sync.RWMutex
	nodeName   string
	dataDir    string
	store      statestore.Store
	runtime    tarakruntime.Runtime
	cri        *tcr.CRIEngine
	cni        *network.InbuiltCNI
	dns        *network.MicroCoreDNS
	log        *zap.Logger
	podStatus  map[string]string // PodKey -> Phase
	probeFails map[string]int    // PodKey -> Failure count
	closed     bool
}

// NewMicroKubelet constructs a new micro-Kubelet node manager.
func NewMicroKubelet(nodeName, dataDir string, store statestore.Store, rt tarakruntime.Runtime, cni *network.InbuiltCNI, dns *network.MicroCoreDNS, log *zap.Logger) *MicroKubelet {
	if log == nil {
		log = zap.NewNop()
	}
	if nodeName == "" {
		nodeName = "tarak-node-01"
	}
	if dataDir == "" {
		dataDir = "./data"
	}

	return &MicroKubelet{
		nodeName:   nodeName,
		dataDir:    dataDir,
		store:      store,
		runtime:    rt,
		cri:        tcr.NewCRIEngine(dataDir, log),
		cni:        cni,
		dns:        dns,
		log:        log.Named("micro-kubelet"),
		podStatus:  make(map[string]string),
		probeFails: make(map[string]int),
	}
}

// Start launches the node registration, pod sync loop, and probe checker.
func (k *MicroKubelet) Start(ctx context.Context) {
	k.log.Info("starting Tarak micro-Kubelet node agent",
		zap.String("node", k.nodeName),
		zap.String("os", runtime.GOOS),
		zap.String("arch", runtime.GOARCH),
	)

	// 1. Register Node object in statestore
	k.registerNode(ctx)

	// 2. Heartbeat loop (every 10s)
	go k.heartbeatLoop(ctx)

	// 3. Pod reconciliation and Probe loop (every 2s)
	go k.podSyncLoop(ctx)
}

func (k *MicroKubelet) registerNode(ctx context.Context) {
	specs := hardware.DetectHost()
	nowStr := time.Now().UTC().Format(time.RFC3339)

	nodeObj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
		"metadata": map[string]interface{}{
			"name": k.nodeName,
			"labels": map[string]interface{}{
				"kubernetes.io/hostname":           k.nodeName,
				"kubernetes.io/os":                 runtime.GOOS,
				"kubernetes.io/arch":               runtime.GOARCH,
				"node.kubernetes.io/instance-type": "tarak.bare-metal",
			},
			"creationTimestamp": nowStr,
		},
		"status": map[string]interface{}{
			"capacity": map[string]interface{}{
				"cpu":    fmt.Sprintf("%d", specs.CPUCores),
				"memory": fmt.Sprintf("%dKi", specs.TotalMemoryBytes/1024),
				"pods":   "110",
			},
			"allocatable": map[string]interface{}{
				"cpu":    fmt.Sprintf("%d", specs.CPUCores),
				"memory": fmt.Sprintf("%dKi", (specs.TotalMemoryBytes*9/10)/1024),
				"pods":   "110",
			},
			"conditions": []map[string]interface{}{
				{
					"type":               "Ready",
					"status":             "True",
					"lastHeartbeatTime":  nowStr,
					"lastTransitionTime": nowStr,
					"reason":             "TarakKubeletReady",
					"message":            "tarak micro-kubelet is posting ready status",
				},
				{
					"type":               "MemoryPressure",
					"status":             "False",
					"lastHeartbeatTime":  nowStr,
					"lastTransitionTime": nowStr,
					"reason":             "TarakKubeletHasSufficientMemory",
					"message":            "tarak micro-kubelet has sufficient memory available",
				},
				{
					"type":               "DiskPressure",
					"status":             "False",
					"lastHeartbeatTime":  nowStr,
					"lastTransitionTime": nowStr,
					"reason":             "TarakKubeletHasNoDiskPressure",
					"message":            "tarak micro-kubelet has sufficient disk space available",
				},
			},
			"nodeInfo": map[string]interface{}{
				"machineID":               fmt.Sprintf("tarak-%s-%s", runtime.GOOS, k.nodeName),
				"systemUUID":              fmt.Sprintf("tarak-%s-%s", runtime.GOOS, k.nodeName),
				"bootID":                  fmt.Sprintf("tarak-%s-%s", runtime.GOOS, k.nodeName),
				"kernelVersion":           runtime.GOOS + "-kernel",
				"osImage":                 fmt.Sprintf("Tarak Native Node (%s)", runtime.GOOS),
				"containerRuntimeVersion": "tcr://v1.0.6",
				"kubeletVersion":          "v1.31.0-tarak",
				"architecture":            runtime.GOARCH,
				"operatingSystem":         runtime.GOOS,
			},
		},
	}

	nodeRaw, err := json.Marshal(nodeObj)
	if err == nil {
		key := statestore.ResourceKey{Group: "", Version: "v1", Resource: "nodes", Name: k.nodeName}
		if _, getErr := k.store.Get(ctx, key); getErr != nil {
			_, _ = k.store.Create(ctx, key, nodeRaw)
			k.log.Info("registered node in cluster", zap.String("node", k.nodeName))
		} else {
			_, _ = k.store.Update(ctx, key, nodeRaw, 0)
		}
	}
}

func (k *MicroKubelet) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.registerNode(ctx)
		}
	}
}

func (k *MicroKubelet) podSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.reconcileAssignedPods(ctx)
		}
	}
}

func (k *MicroKubelet) reconcileAssignedPods(ctx context.Context) {
	podEnvs, _, err := k.store.List(ctx, statestore.ListQuery{
		Key: statestore.ResourceKey{Group: "", Version: "v1", Resource: "pods"},
	})
	if err != nil || len(podEnvs) == 0 {
		return
	}

	for _, pe := range podEnvs {
		var pod map[string]interface{}
		if err := json.Unmarshal(pe.Object, &pod); err != nil {
			continue
		}

		meta, _ := pod["metadata"].(map[string]interface{})
		spec, _ := pod["spec"].(map[string]interface{})
		status, _ := pod["status"].(map[string]interface{})
		if meta == nil || spec == nil {
			continue
		}

		pName, _ := meta["name"].(string)
		pNamespace, _ := meta["namespace"].(string)
		if pNamespace == "" {
			pNamespace = "default"
		}
		podKey := fmt.Sprintf("%s/%s", pNamespace, pName)

		nodeName, _ := spec["nodeName"].(string)
		if nodeName != "" && nodeName != k.nodeName {
			continue // Assigned to another node
		}

		containers, _ := spec["containers"].([]interface{})
		if len(containers) == 0 {
			continue
		}

		// Check Probes for running pods
		phase, _ := status["phase"].(string)
		if phase == "Running" {
			k.checkPodProbes(ctx, podKey, pNamespace, pName, containers, pod)
		}
	}
}

func (k *MicroKubelet) checkPodProbes(ctx context.Context, podKey, namespace, podName string, containers []interface{}, pod map[string]interface{}) {
	for _, c := range containers {
		cMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		// 1. Liveness Probe
		if liveness, ok := cMap["livenessProbe"].(map[string]interface{}); ok {
			healthy := k.executeProbe(namespace, podName, liveness)
			if !healthy {
				k.mu.Lock()
				k.probeFails[podKey]++
				fails := k.probeFails[podKey]
				k.mu.Unlock()

				if fails >= 3 {
					k.log.Warn("pod liveness probe failed 3 times, restarting container", zap.String("pod", podKey))
					// Restart pod
					if k.runtime != nil {
						_ = k.runtime.StopPodContainers(ctx, namespace, podName)
					}
					k.mu.Lock()
					k.probeFails[podKey] = 0
					k.mu.Unlock()
				}
			} else {
				k.mu.Lock()
				k.probeFails[podKey] = 0
				k.mu.Unlock()
			}
		}
	}
}

func (k *MicroKubelet) executeProbe(namespace, podName string, probe map[string]interface{}) bool {
	// HTTPGet Action
	if httpGet, ok := probe["httpGet"].(map[string]interface{}); ok {
		path, _ := httpGet["path"].(string)
		if path == "" {
			path = "/"
		}
		port := 80
		if pVal, ok := httpGet["port"].(float64); ok && pVal > 0 {
			port = int(pVal)
		}
		if k.runtime != nil {
			port = k.runtime.GetHostPort(namespace, podName, port)
		}

		url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
		client := http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
			_ = resp.Body.Close()
			return true
		}
		return false
	}

	// TCPSocket Action
	if tcpSock, ok := probe["tcpSocket"].(map[string]interface{}); ok {
		port := 80
		if pVal, ok := tcpSock["port"].(float64); ok && pVal > 0 {
			port = int(pVal)
		}
		if k.runtime != nil {
			port = k.runtime.GetHostPort(namespace, podName, port)
		}
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 1*time.Second)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}

	return true
}

// MountPodVolumes constructs ConfigMap, Secret, and EmptyDir mounts for a pod.
func (k *MicroKubelet) MountPodVolumes(ctx context.Context, namespace, podName string, volumes []interface{}) (map[string]string, error) {
	podVolDir := filepath.Join(k.dataDir, "volumes", namespace, podName)
	_ = os.MkdirAll(podVolDir, 0755)

	mountMap := make(map[string]string)

	for _, v := range volumes {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		vName, _ := vMap["name"].(string)

		// 1. EmptyDir Volume
		if _, ok := vMap["emptyDir"]; ok {
			dirPath := filepath.Join(podVolDir, vName)
			_ = os.MkdirAll(dirPath, 0755)
			mountMap[vName] = dirPath
		}

		// 2. ConfigMap Volume
		if cmRef, ok := vMap["configMap"].(map[string]interface{}); ok {
			cmName, _ := cmRef["name"].(string)
			dirPath := filepath.Join(podVolDir, vName)
			_ = os.MkdirAll(dirPath, 0755)

			cmKey := statestore.ResourceKey{Group: "", Version: "v1", Resource: "configmaps", Namespace: namespace, Name: cmName}
			if cmEnv, err := k.store.Get(ctx, cmKey); err == nil {
				var cmObj map[string]interface{}
				if json.Unmarshal(cmEnv.Object, &cmObj) == nil {
					data, _ := cmObj["data"].(map[string]interface{})
					for filename, content := range data {
						if contentStr, ok := content.(string); ok {
							_ = os.WriteFile(filepath.Join(dirPath, filename), []byte(contentStr), 0644)
						}
					}
				}
			}
			mountMap[vName] = dirPath
		}

		// 3. Secret Volume
		if secRef, ok := vMap["secret"].(map[string]interface{}); ok {
			secName, _ := secRef["secretName"].(string)
			dirPath := filepath.Join(podVolDir, vName)
			_ = os.MkdirAll(dirPath, 0755)

			secKey := statestore.ResourceKey{Group: "", Version: "v1", Resource: "secrets", Namespace: namespace, Name: secName}
			if secEnv, err := k.store.Get(ctx, secKey); err == nil {
				var secObj map[string]interface{}
				if json.Unmarshal(secEnv.Object, &secObj) == nil {
					data, _ := secObj["data"].(map[string]interface{})
					for filename, content := range data {
						if contentStr, ok := content.(string); ok {
							_ = os.WriteFile(filepath.Join(dirPath, filename), []byte(contentStr), 0644)
						}
					}
				}
			}
			mountMap[vName] = dirPath
		}

		// 4. HostPath Volume
		if hpRef, ok := vMap["hostPath"].(map[string]interface{}); ok {
			path, _ := hpRef["path"].(string)
			if path != "" {
				mountMap[vName] = path
			}
		}
	}

	return mountMap, nil
}
