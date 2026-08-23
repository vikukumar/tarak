package mesh

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Engine is the central Service Mesh controller managing routes, policies, and circuit breaking.
type Engine struct {
	mu             sync.RWMutex
	log            *zap.Logger
	routes         map[string]*TrafficRoute // key: ns/name
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
}

// NewEngine creates a new Service Mesh Engine.
func NewEngine(log *zap.Logger) *Engine {
	return &Engine{
		log:            log.Named("mesh"),
		routes:         make(map[string]*TrafficRoute),
		circuitBreaker: NewCircuitBreaker(5, 10*time.Second),
		rateLimiter:    NewRateLimiter(),
	}
}

// RegisterRoute adds or updates a traffic route.
func (e *Engine) RegisterRoute(route *TrafficRoute) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s/%s", route.Namespace, route.Name)
	e.routes[key] = route
	e.log.Info("registered mesh traffic route", zap.String("key", key), zap.String("host", route.Host))
}

// SelectDestination picks a canary destination based on defined weights.
func (e *Engine) SelectDestination(route *TrafficRoute) *Destination {
	if len(route.Destinations) == 0 {
		return nil
	}
	if len(route.Destinations) == 1 {
		return &route.Destinations[0]
	}

	totalWeight := 0
	for _, d := range route.Destinations {
		w := d.Weight
		if w <= 0 {
			w = 100 / len(route.Destinations)
		}
		totalWeight += w
	}

	if totalWeight == 0 {
		return &route.Destinations[0]
	}

	rnd := rand.Intn(totalWeight)
	curr := 0
	for i, d := range route.Destinations {
		w := d.Weight
		if w <= 0 {
			w = 100 / len(route.Destinations)
		}
		curr += w
		if rnd < curr {
			return &route.Destinations[i]
		}
	}

	return &route.Destinations[0]
}

// HandleListRoutes serves the active mesh routes via HTTP API.
func (e *Engine) HandleListRoutes(w http.ResponseWriter, r *http.Request) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]*TrafficRoute, 0, len(e.routes))
	for _, r := range e.routes {
		list = append(list, r)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

// HandleCreateRoute creates a new mesh traffic route via HTTP API.
func (e *Engine) HandleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var req TrafficRoute
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Name == "" {
		http.Error(w, "missing route name", http.StatusBadRequest)
		return
	}

	e.RegisterRoute(&req)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

// HandleDeleteRoute removes a route by namespace and name.
func (e *Engine) HandleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	key := fmt.Sprintf("%s/%s", ns, name)

	e.mu.Lock()
	delete(e.routes, key)
	e.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
