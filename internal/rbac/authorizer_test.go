package rbac

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestRBAC_SuperuserAuthorize(t *testing.T) {
	log, _ := zap.NewDevelopment()
	auth := NewAuthorizer(nil, log)

	// Superuser test
	allowed := auth.Authorize(context.Background(), "admin", []string{"system:masters"}, "create", "apps", "deployments", "default", "web")
	if !allowed {
		t.Fatal("expected admin to be authorized")
	}

	// Read operations test
	allowedRead := auth.Authorize(context.Background(), "viewer", []string{"system:authenticated"}, "get", "", "pods", "default", "my-pod")
	if !allowedRead {
		t.Fatal("expected viewer to be authorized for get")
	}
}
