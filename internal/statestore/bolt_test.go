// Package statestore — BoltDB integration tests.
package statestore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func openTestStore(t *testing.T) *statestore.BoltStore {
	t.Helper()
	dir := t.TempDir()
	store, err := statestore.Open(statestore.Options{
		Path:   filepath.Join(dir, "test.db"),
		Logger: zap.NewNop(),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func podKey(ns, name string) statestore.ResourceKey {
	return statestore.ResourceKey{
		Group:     "",
		Version:   "v1",
		Resource:  "pods",
		Namespace: ns,
		Name:      name,
	}
}

func testPodJSON(ns, name string) []byte {
	pod := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": ns,
			"labels":    map[string]string{"app": "test", "env": "prod"},
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{"name": "nginx", "image": "nginx:latest"},
			},
		},
	}
	data, _ := json.Marshal(pod)
	return data
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	env, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	assert.NotEmpty(t, env.UID)
	assert.Equal(t, int64(1), env.ResourceVersion)
	assert.Equal(t, int64(1), env.Generation)
	assert.Equal(t, "nginx", env.Name)
	assert.Equal(t, "default", env.Namespace)
	assert.Equal(t, "test", env.Labels["app"])
}

func TestCreateAlreadyExists(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	_, err = store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.Error(t, err)
	assert.ErrorIs(t, err, statestore.ErrAlreadyExists)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func TestGet(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	got, err := store.Get(ctx, podKey("default", "nginx"))
	require.NoError(t, err)
	assert.Equal(t, created.UID, got.UID)
	assert.Equal(t, created.ResourceVersion, got.ResourceVersion)
}

func TestGetNotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, err := store.Get(ctx, podKey("default", "missing"))
	require.Error(t, err)
	assert.ErrorIs(t, err, statestore.ErrNotFound)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	updated, err := store.Update(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"), created.ResourceVersion)
	require.NoError(t, err)

	assert.Greater(t, updated.ResourceVersion, created.ResourceVersion)
	assert.Equal(t, int64(2), updated.Generation)
	assert.Equal(t, created.UID, updated.UID)
}

func TestUpdateConflict(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	// First update succeeds.
	updated, err := store.Update(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"), created.ResourceVersion)
	require.NoError(t, err)

	// Second update with old RV fails.
	_, err = store.Update(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"), created.ResourceVersion)
	require.Error(t, err)
	assert.ErrorIs(t, err, statestore.ErrConflict)
	_ = updated
}

func TestUpdateForceNoRV(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	// resourceVersion 0 means force update (no optimistic lock).
	_, err = store.Update(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"), 0)
	require.NoError(t, err)
}

// ─── StatusUpdate ─────────────────────────────────────────────────────────────

func TestStatusUpdate(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	updated, err := store.StatusUpdate(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"), created.ResourceVersion)
	require.NoError(t, err)

	// Generation must NOT change on status update.
	assert.Equal(t, created.Generation, updated.Generation)
	// ResourceVersion must change.
	assert.Greater(t, updated.ResourceVersion, created.ResourceVersion)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	_, err = store.Delete(ctx, podKey("default", "nginx"), created.ResourceVersion)
	require.NoError(t, err)

	// Should be gone.
	_, err = store.Get(ctx, podKey("default", "nginx"))
	assert.ErrorIs(t, err, statestore.ErrNotFound)
}

func TestDeleteNotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, err := store.Delete(ctx, podKey("default", "missing"), 0)
	assert.ErrorIs(t, err, statestore.ErrNotFound)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestList(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		_, err := store.Create(ctx, podKey("default", name), testPodJSON("default", name))
		require.NoError(t, err)
	}

	query := statestore.ListQuery{
		Key: statestore.ResourceKey{Version: "v1", Resource: "pods", Namespace: "default"},
	}
	results, rev, err := store.List(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Greater(t, rev, int64(0))
}

func TestListCrossNamespace(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, _ = store.Create(ctx, podKey("ns1", "pod1"), testPodJSON("ns1", "pod1"))
	_, _ = store.Create(ctx, podKey("ns2", "pod2"), testPodJSON("ns2", "pod2"))
	_, _ = store.Create(ctx, podKey("ns1", "pod3"), testPodJSON("ns1", "pod3"))

	// Query specific namespace.
	query := statestore.ListQuery{
		Key: statestore.ResourceKey{Version: "v1", Resource: "pods", Namespace: "ns1"},
	}
	results, _, err := store.List(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestListEmpty(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	query := statestore.ListQuery{
		Key: statestore.ResourceKey{Version: "v1", Resource: "pods"},
	}
	results, _, err := store.List(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

// ─── Watch ────────────────────────────────────────────────────────────────────

func TestWatchCreate(t *testing.T) {
	store := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := statestore.WatchQuery{
		Key: statestore.ResourceKey{Version: "v1", Resource: "pods", Namespace: "default"},
	}
	ch, err := store.Watch(ctx, query)
	require.NoError(t, err)

	// Create an object — should generate an ADDED event.
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, _ = store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	}()

	select {
	case ev := <-ch:
		assert.Equal(t, statestore.EventAdded, ev.Type)
		assert.Equal(t, "nginx", ev.Envelope.Name)
	case <-ctx.Done():
		t.Fatal("timeout waiting for ADDED event")
	}
}

func TestWatchDelete(t *testing.T) {
	store := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create first.
	env, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	query := statestore.WatchQuery{
		Key: statestore.ResourceKey{Version: "v1", Resource: "pods"},
	}
	ch, err := store.Watch(ctx, query)
	require.NoError(t, err)

	// Delete.
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, _ = store.Delete(ctx, podKey("default", "nginx"), env.ResourceVersion)
	}()

	select {
	case ev := <-ch:
		assert.Equal(t, statestore.EventDeleted, ev.Type)
	case <-ctx.Done():
		t.Fatal("timeout waiting for DELETED event")
	}
}

func TestWatchConcurrentWatchers(t *testing.T) {
	store := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const numWatchers = 10
	var wg sync.WaitGroup
	received := make([]int, numWatchers)

	for i := 0; i < numWatchers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ch, err := store.Watch(ctx, statestore.WatchQuery{
				Key: statestore.ResourceKey{Version: "v1", Resource: "pods"},
			})
			if err != nil {
				return
			}
			for ev := range ch {
				if ev.Type == statestore.EventAdded {
					received[idx]++
					return
				}
			}
		}(i)
	}

	// Allow watchers to subscribe.
	time.Sleep(100 * time.Millisecond)

	_, err := store.Create(ctx, podKey("default", "nginx"), testPodJSON("default", "nginx"))
	require.NoError(t, err)

	wg.Wait()

	// All watchers should have received the ADDED event.
	for i, count := range received {
		assert.Equal(t, 1, count, "watcher %d didn't receive event", i)
	}
}

// ─── Revision ─────────────────────────────────────────────────────────────────

func TestRevisionMonotonicallyIncreases(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	var prevRev int64
	for i := 0; i < 5; i++ {
		_, err := store.Create(ctx, podKey("default", fmt.Sprintf("pod-%d", i)), testPodJSON("default", fmt.Sprintf("pod-%d", i)))
		require.NoError(t, err)

		rev, err := store.CurrentRevision(ctx)
		require.NoError(t, err)
		assert.Greater(t, rev, prevRev, "revision should increase after write %d", i)
		prevRev = rev
	}
}

// ─── Persistence ──────────────────────────────────────────────────────────────

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Write to store.
	{
		store, err := statestore.Open(statestore.Options{Path: dbPath, Logger: zap.NewNop()})
		require.NoError(t, err)
		_, err = store.Create(context.Background(), podKey("default", "nginx"), testPodJSON("default", "nginx"))
		require.NoError(t, err)
		_ = store.Close()
	}

	// Re-open and verify data survived.
	{
		store, err := statestore.Open(statestore.Options{Path: dbPath, Logger: zap.NewNop()})
		require.NoError(t, err)
		defer store.Close()

		env, err := store.Get(context.Background(), podKey("default", "nginx"))
		require.NoError(t, err)
		assert.Equal(t, "nginx", env.Name)
	}
}

// ─── Key validation ───────────────────────────────────────────────────────────

func TestKeyValidation(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// Missing version.
	_, err := store.Create(ctx, statestore.ResourceKey{Resource: "pods", Name: "x"}, []byte("{}"))
	assert.Error(t, err)

	// Missing resource.
	_, err = store.Create(ctx, statestore.ResourceKey{Version: "v1", Name: "x"}, []byte("{}"))
	assert.Error(t, err)
}
