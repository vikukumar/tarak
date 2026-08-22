package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
)

// Controller reconciles Ingress resources and syncs route tables with the Router.
type Controller struct {
	store  statestore.Store
	router *Router
	log    *zap.Logger
}

// NewController creates a new Ingress controller.
func NewController(store statestore.Store, router *Router, log *zap.Logger) *Controller {
	if log == nil {
		log = zap.NewNop()
	}
	return &Controller{
		store:  store,
		router: router,
		log:    log.Named("ingress-controller"),
	}
}

// Reconcile extracts all Ingress rules across namespaces and updates the router.
func (c *Controller) Reconcile(ctx context.Context, publicHost string) error {
	var allRoutes []Route

	groups := []string{"networking.k8s.io", "networking.tarak.io"}
	for _, group := range groups {
		envs, _, err := c.store.List(ctx, statestore.ListQuery{
			Key: statestore.ResourceKey{
				Group:    group,
				Version:  "v1",
				Resource: "ingresses",
			},
		})
		if err != nil || len(envs) == 0 {
			continue
		}

		for _, env := range envs {
			var ing map[string]interface{}
			if err := json.Unmarshal(env.Object, &ing); err != nil {
				continue
			}

			meta, _ := ing["metadata"].(map[string]interface{})
			spec, _ := ing["spec"].(map[string]interface{})
			if meta == nil || spec == nil {
				continue
			}

			ns, _ := meta["namespace"].(string)
			if ns == "" {
				ns = "default"
			}
			name, _ := meta["name"].(string)

			// Check IngressClassName (tarak, tarak-cloudflare, tarak-tailscale, or empty default)
			className, _ := spec["ingressClassName"].(string)
			if className != "" && !strings.HasPrefix(className, "tarak") {
				continue
			}

			// Process Rules
			rules, _ := spec["rules"].([]interface{})
			for _, r := range rules {
				ruleMap, _ := r.(map[string]interface{})
				if ruleMap == nil {
					continue
				}
				host, _ := ruleMap["host"].(string)
				httpRule, _ := ruleMap["http"].(map[string]interface{})
				if httpRule == nil {
					continue
				}
				paths, _ := httpRule["paths"].([]interface{})
				for _, p := range paths {
					pathMap, _ := p.(map[string]interface{})
					if pathMap == nil {
						continue
					}
					pathStr, _ := pathMap["path"].(string)
					backend, _ := pathMap["backend"].(map[string]interface{})
					if backend == nil {
						continue
					}
					svcMap, _ := backend["service"].(map[string]interface{})
					if svcMap == nil {
						continue
					}
					svcName, _ := svcMap["name"].(string)
					var svcPort int
					if portMap, ok := svcMap["port"].(map[string]interface{}); ok {
						if num, ok := portMap["number"].(float64); ok {
							svcPort = int(num)
						} else if numStr, ok := portMap["number"].(string); ok {
							svcPort, _ = strconv.Atoi(numStr)
						}
					}
					if svcPort == 0 {
						svcPort = 80
					}

					// Resolve Service ClusterIP or Port
					backendTarget := c.resolveServiceBackend(ctx, ns, svcName, svcPort)
					if backendTarget != nil {
						allRoutes = append(allRoutes, Route{
							Host:        host,
							Path:        pathStr,
							BackendURL:  backendTarget,
							ServiceName: svcName,
							ServicePort: svcPort,
							Namespace:   ns,
						})
					}
				}
			}

			// Update Ingress Status with public hostname if available
			if publicHost != "" && name != "" {
				key := statestore.ResourceKey{
					Group:     group,
					Version:   "v1",
					Resource:  "ingresses",
					Namespace: ns,
					Name:      name,
				}
				c.updateIngressStatus(ctx, key, ing, env.ResourceVersion, publicHost)
			}
		}
	}

	c.router.UpdateRoutes(allRoutes)
	return nil
}

// resolveServiceBackend resolves a service to its loopback target or ClusterIP endpoint.
func (c *Controller) resolveServiceBackend(ctx context.Context, namespace, serviceName string, port int) *url.URL {
	// Look up the Service object in statestore
	key := statestore.ResourceKey{
		Group:     "",
		Version:   "v1",
		Resource:  "services",
		Namespace: namespace,
		Name:      serviceName,
	}

	env, err := c.store.Get(ctx, key)
	if err != nil {
		// Fallback to direct localhost port
		u, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
		return u
	}

	var svc map[string]interface{}
	if err := json.Unmarshal(env.Object, &svc); err == nil {
		spec, _ := svc["spec"].(map[string]interface{})
		if spec != nil {
			clusterIP, _ := spec["clusterIP"].(string)
			if clusterIP != "" && clusterIP != "None" {
				u, _ := url.Parse(fmt.Sprintf("http://%s:%d", clusterIP, port))
				return u
			}
		}
	}

	u, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	return u
}

func (c *Controller) updateIngressStatus(ctx context.Context, key statestore.ResourceKey, ing map[string]interface{}, rv int64, publicHost string) {
	status, _ := ing["status"].(map[string]interface{})
	if status == nil {
		status = make(map[string]interface{})
		ing["status"] = status
	}

	lb, _ := status["loadBalancer"].(map[string]interface{})
	if lb == nil {
		lb = make(map[string]interface{})
		status["loadBalancer"] = lb
	}

	lbIngress := []map[string]interface{}{
		{"hostname": publicHost},
	}
	lb["ingress"] = lbIngress

	raw, err := json.Marshal(ing)
	if err != nil {
		return
	}
	_, _ = c.store.StatusUpdate(ctx, key, raw, rv)
}
