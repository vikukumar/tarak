package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WebhookType represents the category of the webhook.
type WebhookType string

const (
	TypeEventWebhook     WebhookType = "EventNotification"
	TypeAdmissionWebhook WebhookType = "AdmissionValidation"
	TypeMutationWebhook  WebhookType = "AdmissionMutation"
)

// WebhookEndpoint defines a registered webhook URL and its trigger events.
type WebhookEndpoint struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	URL         string      `json:"url"`
	Type        WebhookType `json:"type"`
	Events      []string    `json:"events"` // "pod.crash", "node.notready", "policy.violation", "hpa.scale", "mesh.alert"
	SecretToken string      `json:"secretToken,omitempty"`
	Enabled     bool        `json:"enabled"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// DeliveryRecord tracks an individual webhook execution.
type DeliveryRecord struct {
	ID          string    `json:"id"`
	WebhookID   string    `json:"webhookId"`
	WebhookName string    `json:"webhookName"`
	Event       string    `json:"event"`
	StatusCode  int       `json:"statusCode"`
	Success     bool      `json:"success"`
	LatencyMs   float64   `json:"latencyMs"`
	Timestamp   time.Time `json:"timestamp"`
	Payload     string    `json:"payload"`
	Response    string    `json:"response,omitempty"`
}

// Manager handles webhook registrations, event subscriptions, and async deliveries.
type Manager struct {
	log         *zap.Logger
	mu          sync.RWMutex
	webhooks    map[string]*WebhookEndpoint
	deliveries  []DeliveryRecord
	client      *http.Client
}

// NewManager initializes the Webhook subsystem with default event handlers.
func NewManager(log *zap.Logger) *Manager {
	if log == nil {
		log = zap.NewNop()
	}

	m := &Manager{
		log:        log.Named("webhook-manager"),
		webhooks:   make(map[string]*WebhookEndpoint),
		deliveries: make([]DeliveryRecord, 0),
		client:     &http.Client{Timeout: 5 * time.Second},
	}

	m.seedDefaultWebhooks()
	return m
}

func (m *Manager) seedDefaultWebhooks() {
	now := time.Now()
	wh1 := &WebhookEndpoint{
		ID:        "wh-slack-alerts",
		Name:      "Slack Production Incident Alerts",
		URL:       "https://hooks.slack.com/services/T00/B00/XXXXX",
		Type:      TypeEventWebhook,
		Events:    []string{"pod.crash", "node.notready", "policy.violation"},
		Enabled:   true,
		CreatedAt: now.Add(-48 * time.Hour),
	}
	wh2 := &WebhookEndpoint{
		ID:        "wh-pagerduty-sev1",
		Name:      "PagerDuty Zero-Downtime Monitor",
		URL:       "https://events.pagerduty.com/v2/enqueue",
		Type:      TypeEventWebhook,
		Events:    []string{"node.notready", "mesh.partition"},
		Enabled:   true,
		CreatedAt: now.Add(-24 * time.Hour),
	}

	m.webhooks[wh1.ID] = wh1
	m.webhooks[wh2.ID] = wh2

	m.deliveries = append(m.deliveries,
		DeliveryRecord{
			ID:          "del-001",
			WebhookID:   wh1.ID,
			WebhookName: wh1.Name,
			Event:       "policy.violation",
			StatusCode:  200,
			Success:     true,
			LatencyMs:   124.5,
			Timestamp:   now.Add(-10 * time.Minute),
			Payload:     `{"event":"policy.violation","resource":"pod/default/frontend","rule":"disallow-default-namespace"}`,
			Response:    `{"ok":true}`,
		},
		DeliveryRecord{
			ID:          "del-002",
			WebhookID:   wh2.ID,
			WebhookName: wh2.Name,
			Event:       "hpa.scale",
			StatusCode:  202,
			Success:     true,
			LatencyMs:   88.2,
			Timestamp:   now.Add(-25 * time.Minute),
			Payload:     `{"event":"hpa.scale","resource":"deployment/production/storefront","fromReplicas":2,"toReplicas":5}`,
			Response:    `{"status":"accepted"}`,
		},
	)
}

// RegisterWebhook creates or updates a webhook endpoint.
func (m *Manager) RegisterWebhook(wh *WebhookEndpoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if wh.ID == "" {
		wh.ID = fmt.Sprintf("wh-%d", time.Now().UnixNano())
	}
	wh.CreatedAt = time.Now()
	m.webhooks[wh.ID] = wh
	m.log.Info("registered webhook endpoint", zap.String("id", wh.ID), zap.String("url", wh.URL))
}

// ListWebhooks returns all configured webhook endpoints.
func (m *Manager) ListWebhooks() []*WebhookEndpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*WebhookEndpoint, 0, len(m.webhooks))
	for _, wh := range m.webhooks {
		list = append(list, wh)
	}
	return list
}

// ListDeliveries returns recent webhook execution records.
func (m *Manager) ListDeliveries() []DeliveryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deliveries
}

// DispatchEvent asynchronously fires relevant webhooks for a given cluster event.
func (m *Manager) DispatchEvent(ctx context.Context, eventName string, payload interface{}) {
	m.mu.RLock()
	endpoints := make([]*WebhookEndpoint, 0)
	for _, wh := range m.webhooks {
		if !wh.Enabled {
			continue
		}
		for _, e := range wh.Events {
			if e == eventName || e == "*" {
				endpoints = append(endpoints, wh)
				break
			}
		}
	}
	m.mu.RUnlock()

	if len(endpoints) == 0 {
		return
	}

	payloadBytes, _ := json.Marshal(payload)

	for _, wh := range endpoints {
		go func(endpoint *WebhookEndpoint) {
			start := time.Now()
			req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint.URL, bytes.NewBuffer(payloadBytes))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "Tarak-Webhook-Engine/1.0.6")
			if endpoint.SecretToken != "" {
				req.Header.Set("X-Tarak-Token", endpoint.SecretToken)
			}

			resp, err := m.client.Do(req)
			latency := float64(time.Since(start).Microseconds()) / 1000.0

			rec := DeliveryRecord{
				ID:          fmt.Sprintf("del-%d", time.Now().UnixNano()),
				WebhookID:   endpoint.ID,
				WebhookName: endpoint.Name,
				Event:       eventName,
				LatencyMs:   latency,
				Timestamp:   time.Now(),
				Payload:     string(payloadBytes),
			}

			if err == nil && resp != nil {
				rec.StatusCode = resp.StatusCode
				rec.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
				_ = resp.Body.Close()
			} else {
				rec.StatusCode = 0
				rec.Success = false
				rec.Response = fmt.Sprintf("Error: %v", err)
			}

			m.mu.Lock()
			m.deliveries = append([]DeliveryRecord{rec}, m.deliveries...)
			if len(m.deliveries) > 100 {
				m.deliveries = m.deliveries[:100]
			}
			m.mu.Unlock()
		}(wh)
	}
}
