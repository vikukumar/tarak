// Package statestore — compaction of the watch event ring buffer.
//
// The ring buffer in watchBus is purely in-memory; it does not need explicit
// compaction because it is circular and overwrites old entries automatically.
//
// This file provides compaction for the BoltDB-backed event log (future use when
// we persist watch events to survive restarts) and a utility to compact the
// resources bucket by removing tombstones.
//
// Phase 1: BoltDB watch events are NOT persisted to disk; they live only in
// the in-memory ring buffer.  This file documents where persistence would slot in.
package statestore

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// CompactRevisionsBefore removes completed-delete tombstones from the store
// for resources whose deletion was completed before the given time.
// This is a maintenance operation run by a background goroutine.
//
// Phase 1: no-op — BoltDB deletes are immediate, no tombstones.
// Phase 8+: when Raft log compaction is introduced, this becomes a real operation.
func (s *BoltStore) CompactRevisionsBefore(_ context.Context, _ time.Time) error {
	s.log.Debug("CompactRevisionsBefore: no-op in Phase 1")
	return nil
}

// RunCompaction starts a background goroutine that periodically compacts
// old revisions from the watch ring buffer.
//
// Phase 1: ring buffer self-manages; this is a placeholder for future phases.
func (s *BoltStore) RunCompaction(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				if err := s.CompactRevisionsBefore(ctx, t.Add(-interval)); err != nil {
					s.log.Warn("compaction error", zap.Error(err))
				}
			}
		}
	}()
}
