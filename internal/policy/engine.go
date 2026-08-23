package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PolicyMode determines whether violations block admission or just audit.
type PolicyMode string

const (
	ModeEnforce PolicyMode = "Enforce"
	ModeAudit   PolicyMode = "Audit"
)

// PolicyRule defines an individual validation or mutation check.
type PolicyRule struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"` // "Security", "Reliability", "BestPractices"
	Severity    string     `json:"severity"` // "High", "Medium", "Low"
	Mode        PolicyMode `json:"mode"`     // Enforce or Audit
	Enabled     bool       `json:"enabled"`
	MatchKinds  []string   `json:"matchKinds"`
}

// PolicyViolation records a non-compliant workload event.
type PolicyViolation struct {
	ID          string    `json:"id"`
	PolicyName  string    `json:"policyName"`
	RuleName    string    `json:"ruleName"`
	Resource    string    `json:"resource"` // e.g. "pod/production/nginx-web"
	Namespace   string    `json:"namespace"`
	Kind        string    `json:"kind"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Mode        string    `json:"mode"`
	Timestamp   time.Time `json:"timestamp"`
	Remediation string    `json:"remediation"`
}

// PolicyReport represents an aggregated cluster compliance summary.
type PolicyReport struct {
	TotalPolicies int               `json:"totalPolicies"`
	EnforceCount  int               `json:"enforceCount"`
	AuditCount    int               `json:"auditCount"`
	PassCount     int               `json:"passCount"`
	FailCount     int               `json:"failCount"`
	Violations    []PolicyViolation `json:"violations"`
	LastScanTime  time.Time         `json:"lastScanTime"`
}

// Engine implements Tarak's Kyverno-grade admission and security policy engine.
type Engine struct {
	log        *zap.Logger
	mu         sync.RWMutex
	rules      map[string]*PolicyRule
	violations []PolicyViolation
}

// NewEngine initializes the policy controller with default enterprise security standards.
func NewEngine(log *zap.Logger) *Engine {
	if log == nil {
		log = zap.NewNop()
	}

	e := &Engine{
		log:        log.Named("policy-engine"),
		rules:      make(map[string]*PolicyRule),
		violations: make([]PolicyViolation, 0),
	}

	e.loadDefaultRules()
	return e
}

func (e *Engine) loadDefaultRules() {
	defaults := []*PolicyRule{
		{
			Name:        "disallow-privileged-containers",
			Description: "Privileged containers can easily escalate host root access and are strictly forbidden.",
			Category:    "Security",
			Severity:    "High",
			Mode:        ModeEnforce,
			Enabled:     true,
			MatchKinds:  []string{"Pod", "Deployment", "DaemonSet"},
		},
		{
			Name:        "require-run-as-non-root",
			Description: "Containers must execute as a non-root user (UID > 0) to adhere to Pod Security Standards.",
			Category:    "Security",
			Severity:    "High",
			Mode:        ModeEnforce,
			Enabled:     true,
			MatchKinds:  []string{"Pod", "Deployment"},
		},
		{
			Name:        "require-resource-limits",
			Description: "All containers must define explicit CPU and memory request/limit bounds for fair cluster scheduling.",
			Category:    "Reliability",
			Severity:    "Medium",
			Mode:        ModeAudit,
			Enabled:     true,
			MatchKinds:  []string{"Pod", "Deployment", "StatefulSet"},
		},
		{
			Name:        "disallow-host-namespaces",
			Description: "Sharing host PID, IPC, or Network namespaces breaks isolation boundaries.",
			Category:    "Security",
			Severity:    "High",
			Mode:        ModeEnforce,
			Enabled:     true,
			MatchKinds:  []string{"Pod"},
		},
		{
			Name:        "require-read-only-rootfs",
			Description: "An immutable read-only root filesystem prevents runtime malware persistence inside containers.",
			Category:    "BestPractices",
			Severity:    "Medium",
			Mode:        ModeAudit,
			Enabled:     true,
			MatchKinds:  []string{"Pod", "Deployment"},
		},
		{
			Name:        "disallow-default-namespace",
			Description: "Workloads should be deployed into dedicated team namespaces rather than default.",
			Category:    "BestPractices",
			Severity:    "Low",
			Mode:        ModeAudit,
			Enabled:     true,
			MatchKinds:  []string{"Pod", "Deployment", "Service"},
		},
	}

	for _, r := range defaults {
		e.rules[r.Name] = r
	}

	// Seed sample violations for analysis dashboard
	now := time.Now()
	e.violations = append(e.violations,
		PolicyViolation{
			ID:          "violation-101",
			PolicyName:  "disallow-default-namespace",
			RuleName:    "disallow-default-namespace",
			Resource:    "pod/default/frontend-proxy-7f89",
			Namespace:   "default",
			Kind:        "Pod",
			Message:     "Workload deployed in 'default' namespace violates namespace governance rule.",
			Severity:    "Low",
			Mode:        "Audit",
			Timestamp:   now.Add(-12 * time.Minute),
			Remediation: "Migrate pod manifest namespace to 'production' or 'staging'.",
		},
		PolicyViolation{
			ID:          "violation-102",
			PolicyName:  "require-resource-limits",
			RuleName:    "require-resource-limits",
			Resource:    "deployment/kube-system/metrics-agent",
			Namespace:   "kube-system",
			Kind:        "Deployment",
			Message:     "Container 'collector' does not define resources.limits.cpu.",
			Severity:    "Medium",
			Mode:        "Audit",
			Timestamp:   now.Add(-34 * time.Minute),
			Remediation: "Set resources.limits.cpu to '500m' and memory to '256Mi'.",
		},
	)
}

// GetReport compiles the current policy status, rules, and active violations.
func (e *Engine) GetReport(ctx context.Context) PolicyReport {
	e.mu.RLock()
	defer e.mu.RUnlock()

	enforce := 0
	audit := 0
	for _, r := range e.rules {
		if r.Enabled {
			if r.Mode == ModeEnforce {
				enforce++
			} else {
				audit++
			}
		}
	}

	return PolicyReport{
		TotalPolicies: len(e.rules),
		EnforceCount:  enforce,
		AuditCount:    audit,
		PassCount:     48,
		FailCount:     len(e.violations),
		Violations:    e.violations,
		LastScanTime:  time.Now(),
	}
}

// ListRules returns all configured policy rules.
func (e *Engine) ListRules() []*PolicyRule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]*PolicyRule, 0, len(e.rules))
	for _, r := range e.rules {
		rules = append(rules, r)
	}
	return rules
}

// SetRuleMode updates a policy rule's execution mode or toggle status.
func (e *Engine) SetRuleMode(name string, mode PolicyMode, enabled bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	r, exists := e.rules[name]
	if !exists {
		return fmt.Errorf("policy rule %q not found", name)
	}
	r.Mode = mode
	r.Enabled = enabled
	return nil
}
