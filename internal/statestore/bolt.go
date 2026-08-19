// Package statestore — BoltDB implementation of the Store interface.
//
// Storage layout in BoltDB:
//
//	Bucket "meta"
//	  Key "revision" → uint64 big-endian    (global monotonic revision counter)
//
//	Bucket "resources/<group>/<version>/<resource>"
//	  Key "<namespace>/<name>"              → JSON(Envelope)
//	  Key "<name>"                          (cluster-scoped, no namespace segment)
//
//	Bucket "labelidx"
//	  Key "<label_key>=<label_val>\x00<bucket_path>\x00<storage_key>" → ""
//
//	Bucket "fieldidx"
//	  Key "<field>=<val>\x00<bucket_path>\x00<storage_key>" → ""
//
// Every write is a BoltDB transaction that:
//  1. Reads and increments the global revision.
//  2. Writes the updated Envelope under its bucket/key.
//  3. Updates the label and field indexes.
//  4. After commit, publishes a WatchEvent to the bus.
package statestore

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

// ─── Bucket names ─────────────────────────────────────────────────────────────

var (
	bucketMeta      = []byte("meta")
	bucketResources = []byte("resources")
	bucketLabelIdx  = []byte("labelidx")
	bucketFieldIdx  = []byte("fieldidx")
	keyRevision     = []byte("revision")
)

// ─── BoltStore ────────────────────────────────────────────────────────────────

// BoltStore is the BoltDB-backed implementation of Store.
type BoltStore struct {
	db     *bbolt.DB
	bus    *watchBus
	log    *zap.Logger
	closed chan struct{}
}

// Options configures a BoltStore.
type Options struct {
	// Path is the file path of the BoltDB database.
	Path string
	// Logger is the structured logger.
	Logger *zap.Logger
	// Timeout is the BoltDB open timeout (0 = 5s default).
	Timeout time.Duration
}

// Open opens or creates a BoltStore at the given path.
func Open(opts Options) (*BoltStore, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("statestore.Open: Path must not be empty")
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	db, err := bbolt.Open(opts.Path, 0600, &bbolt.Options{Timeout: timeout})
	if err != nil {
		return nil, fmt.Errorf("statestore.Open: bbolt.Open %q: %w", opts.Path, err)
	}

	// Ensure required top-level buckets exist.
	if err := db.Update(func(tx *bbolt.Tx) error {
		for _, b := range [][]byte{bucketMeta, bucketResources, bucketLabelIdx, bucketFieldIdx} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return fmt.Errorf("create bucket %q: %w", b, err)
			}
		}
		// Initialise revision to 0 if not present.
		m := tx.Bucket(bucketMeta)
		if m.Get(keyRevision) == nil {
			if err := m.Put(keyRevision, uint64ToBytes(0)); err != nil {
				return fmt.Errorf("initialise revision: %w", err)
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("statestore.Open: bootstrap: %w", err)
	}

	s := &BoltStore{
		db:     db,
		bus:    newWatchBus(),
		log:    opts.Logger,
		closed: make(chan struct{}),
	}
	return s, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *BoltStore) Create(ctx context.Context, key ResourceKey, obj RawObject) (*Envelope, error) {
	if err := key.ValidateWithName(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var env *Envelope
	var watchEv WatchEvent

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := resourceBucket(tx, key)
		if err != nil {
			return err
		}
		sk := []byte(key.StorageKey())
		if bkt.Get(sk) != nil {
			return &AlreadyExistsError{Key: key}
		}

		rev, err := incrementRevision(tx)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		uid := generateUID()

		env = &Envelope{
			ResourceVersion: rev,
			Generation:      1,
			UID:             uid,
			CreatedAt:       now,
			UpdatedAt:       now,
			Object:          obj,
		}

		// Parse labels, namespace, name from the object for indexing.
		if err := populateEnvelopeFromObject(env, obj, key); err != nil {
			return fmt.Errorf("parse object: %w", err)
		}

		// Inject server-set fields back into the object JSON.
		patchedObj, err := injectServerFields(obj, uid, rev, now, now)
		if err != nil {
			return fmt.Errorf("inject server fields: %w", err)
		}
		env.Object = patchedObj

		data, err := json.Marshal(env)
		if err != nil {
			return fmt.Errorf("marshal envelope: %w", err)
		}
		if err := bkt.Put(sk, data); err != nil {
			return err
		}

		// Update indexes.
		if err := updateLabelIndex(tx, key, nil, env.Labels); err != nil {
			return err
		}
		if err := updateFieldIndex(tx, key, nil, extractFields(env)); err != nil {
			return err
		}

		watchEv = WatchEvent{Type: EventAdded, Envelope: env, Key: key}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.publish(watchEv)
	return env, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *BoltStore) Update(ctx context.Context, key ResourceKey, obj RawObject, resourceVersion int64) (*Envelope, error) {
	if err := key.ValidateWithName(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var env *Envelope
	var watchEv WatchEvent

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := resourceBucket(tx, key)
		if err != nil {
			return err
		}
		sk := []byte(key.StorageKey())
		existing := bkt.Get(sk)
		if existing == nil {
			return &NotFoundError{Key: key}
		}

		var prev Envelope
		if err := json.Unmarshal(existing, &prev); err != nil {
			return fmt.Errorf("unmarshal existing: %w", err)
		}

		if resourceVersion != 0 && prev.ResourceVersion != resourceVersion {
			return &ConflictError{
				Key:             key,
				ProvidedVersion: resourceVersion,
				CurrentVersion:  prev.ResourceVersion,
			}
		}

		rev, err := incrementRevision(tx)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		env = &Envelope{
			ResourceVersion: rev,
			Generation:      prev.Generation + 1,
			UID:             prev.UID,
			CreatedAt:       prev.CreatedAt,
			UpdatedAt:       now,
			DeletionTimestamp: prev.DeletionTimestamp,
			Object:          obj,
		}
		if err := populateEnvelopeFromObject(env, obj, key); err != nil {
			return fmt.Errorf("parse object: %w", err)
		}

		patchedObj, err := injectServerFieldsUpdate(obj, prev.UID, rev, prev.CreatedAt, now)
		if err != nil {
			return fmt.Errorf("inject server fields: %w", err)
		}
		env.Object = patchedObj

		data, err := json.Marshal(env)
		if err != nil {
			return fmt.Errorf("marshal envelope: %w", err)
		}
		if err := bkt.Put(sk, data); err != nil {
			return err
		}

		// Update indexes (remove old labels, add new).
		if err := updateLabelIndex(tx, key, prev.Labels, env.Labels); err != nil {
			return err
		}
		if err := updateFieldIndex(tx, key, extractFields(&prev), extractFields(env)); err != nil {
			return err
		}

		watchEv = WatchEvent{Type: EventModified, Envelope: env, Key: key}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.publish(watchEv)
	return env, nil
}

// ─── StatusUpdate ─────────────────────────────────────────────────────────────

func (s *BoltStore) StatusUpdate(ctx context.Context, key ResourceKey, obj RawObject, resourceVersion int64) (*Envelope, error) {
	if err := key.ValidateWithName(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var env *Envelope
	var watchEv WatchEvent

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := resourceBucket(tx, key)
		if err != nil {
			return err
		}
		sk := []byte(key.StorageKey())
		existing := bkt.Get(sk)
		if existing == nil {
			return &NotFoundError{Key: key}
		}

		var prev Envelope
		if err := json.Unmarshal(existing, &prev); err != nil {
			return fmt.Errorf("unmarshal existing: %w", err)
		}

		if resourceVersion != 0 && prev.ResourceVersion != resourceVersion {
			return &ConflictError{
				Key:             key,
				ProvidedVersion: resourceVersion,
				CurrentVersion:  prev.ResourceVersion,
			}
		}

		rev, err := incrementRevision(tx)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		// For status updates, Generation does NOT change.
		env = &Envelope{
			ResourceVersion:   rev,
			Generation:        prev.Generation,
			UID:               prev.UID,
			CreatedAt:         prev.CreatedAt,
			UpdatedAt:         now,
			DeletionTimestamp: prev.DeletionTimestamp,
			Labels:            prev.Labels,
			Namespace:         prev.Namespace,
			Name:              prev.Name,
			Finalizers:        prev.Finalizers,
			Object:            obj,
		}

		patchedObj, err := injectServerFieldsUpdate(obj, prev.UID, rev, prev.CreatedAt, now)
		if err != nil {
			return fmt.Errorf("inject server fields: %w", err)
		}
		env.Object = patchedObj

		data, err := json.Marshal(env)
		if err != nil {
			return fmt.Errorf("marshal envelope: %w", err)
		}
		if err := bkt.Put(sk, data); err != nil {
			return err
		}

		watchEv = WatchEvent{Type: EventModified, Envelope: env, Key: key}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.publish(watchEv)
	return env, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *BoltStore) Delete(ctx context.Context, key ResourceKey, resourceVersion int64) (*Envelope, error) {
	if err := key.ValidateWithName(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var env *Envelope
	var watchEv WatchEvent

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := resourceBucket(tx, key)
		if err != nil {
			return err
		}
		sk := []byte(key.StorageKey())
		existing := bkt.Get(sk)
		if existing == nil {
			return &NotFoundError{Key: key}
		}

		var prev Envelope
		if err := json.Unmarshal(existing, &prev); err != nil {
			return fmt.Errorf("unmarshal existing: %w", err)
		}

		if resourceVersion != 0 && prev.ResourceVersion != resourceVersion {
			return &ConflictError{
				Key:             key,
				ProvidedVersion: resourceVersion,
				CurrentVersion:  prev.ResourceVersion,
			}
		}

		// Increment revision for the delete event.
		rev, err := incrementRevision(tx)
		if err != nil {
			return err
		}
		prev.ResourceVersion = rev

		if err := bkt.Delete(sk); err != nil {
			return err
		}

		// Clean up indexes.
		if err := updateLabelIndex(tx, key, prev.Labels, nil); err != nil {
			return err
		}
		if err := updateFieldIndex(tx, key, extractFields(&prev), nil); err != nil {
			return err
		}

		env = &prev
		watchEv = WatchEvent{Type: EventDeleted, Envelope: env, Key: key}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.bus.publish(watchEv)
	return env, nil
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (s *BoltStore) Get(ctx context.Context, key ResourceKey) (*Envelope, error) {
	if err := key.ValidateWithName(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var env Envelope
	err := s.db.View(func(tx *bbolt.Tx) error {
		bkt := resourceBucketRead(tx, key)
		if bkt == nil {
			return &NotFoundError{Key: key}
		}
		data := bkt.Get([]byte(key.StorageKey()))
		if data == nil {
			return &NotFoundError{Key: key}
		}
		return json.Unmarshal(data, &env)
	})
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *BoltStore) List(ctx context.Context, query ListQuery) ([]*Envelope, int64, error) {
	if err := query.Key.Validate(); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}

	var results []*Envelope
	var currentRev int64

	err := s.db.View(func(tx *bbolt.Tx) error {
		// Read current revision.
		m := tx.Bucket(bucketMeta)
		if m != nil {
			currentRev = int64(bytesToUint64(m.Get(keyRevision)))
		}

		bkt := resourceBucketRead(tx, query.Key)
		if bkt == nil {
			// Bucket doesn't exist yet — return empty list.
			return nil
		}

		prefix := []byte(query.Key.PrefixForNamespace())

		c := bkt.Cursor()
		var k, v []byte
		if len(prefix) == 0 {
			k, v = c.First()
		} else {
			k, v = c.Seek(prefix)
		}

		// Handle continue token (it's the last-seen storage key).
		if query.Continue != "" {
			// Advance past the continue key.
			_, _ = c.Seek([]byte(query.Continue))
			k, v = c.Next()
		}

		for ; k != nil; k, v = c.Next() {
			// Namespace prefix check.
			if len(prefix) > 0 && !strings.HasPrefix(string(k), string(prefix)) {
				break
			}

			var env Envelope
			if err := json.Unmarshal(v, &env); err != nil {
				s.log.Warn("corrupt envelope, skipping", zap.String("key", string(k)), zap.Error(err))
				continue
			}

			// Apply label selector.
			if query.LabelSelector != nil && !query.LabelSelector.Matches(env.Labels) {
				continue
			}

			// Apply field selector.
			if query.FieldSelector != nil && !query.FieldSelector.Matches(extractFields(&env)) {
				continue
			}

			results = append(results, &env)

			// Limit check.
			if query.Limit > 0 && int64(len(results)) >= query.Limit {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return results, currentRev, nil
}

// ─── Watch ────────────────────────────────────────────────────────────────────

func (s *BoltStore) Watch(ctx context.Context, query WatchQuery) (<-chan WatchEvent, error) {
	if err := query.Key.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyInvalid, err)
	}
	return s.bus.subscribe(ctx, query)
}

// ─── CurrentRevision ──────────────────────────────────────────────────────────

func (s *BoltStore) CurrentRevision(ctx context.Context) (int64, error) {
	var rev int64
	err := s.db.View(func(tx *bbolt.Tx) error {
		m := tx.Bucket(bucketMeta)
		if m == nil {
			return nil
		}
		rev = int64(bytesToUint64(m.Get(keyRevision)))
		return nil
	})
	return rev, err
}

// ─── Close ────────────────────────────────────────────────────────────────────

func (s *BoltStore) Close() error {
	select {
	case <-s.closed:
		return ErrStoreClosed
	default:
		close(s.closed)
	}
	return s.db.Close()
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// resourceBucket returns (or creates) the nested bucket for a resource type.
// Call within an Update transaction.
func resourceBucket(tx *bbolt.Tx, key ResourceKey) (*bbolt.Bucket, error) {
	root := tx.Bucket(bucketResources)
	if root == nil {
		return nil, fmt.Errorf("resources bucket missing")
	}
	path := key.BucketPath()
	bkt, err := root.CreateBucketIfNotExists([]byte(path))
	if err != nil {
		return nil, fmt.Errorf("create resource bucket %q: %w", path, err)
	}
	return bkt, nil
}

// resourceBucketRead returns the nested bucket for a resource type for reading.
// Returns nil (not an error) if the bucket doesn't exist yet.
func resourceBucketRead(tx *bbolt.Tx, key ResourceKey) *bbolt.Bucket {
	root := tx.Bucket(bucketResources)
	if root == nil {
		return nil
	}
	return root.Bucket([]byte(key.BucketPath()))
}

// incrementRevision atomically increments and returns the new global revision.
func incrementRevision(tx *bbolt.Tx) (int64, error) {
	m := tx.Bucket(bucketMeta)
	if m == nil {
		return 0, fmt.Errorf("meta bucket missing")
	}
	cur := bytesToUint64(m.Get(keyRevision))
	cur++
	if err := m.Put(keyRevision, uint64ToBytes(cur)); err != nil {
		return 0, fmt.Errorf("increment revision: %w", err)
	}
	return int64(cur), nil
}

// updateLabelIndex removes oldLabels from the index and adds newLabels.
// Pass nil for oldLabels to only add; nil for newLabels to only remove.
func updateLabelIndex(tx *bbolt.Tx, key ResourceKey, oldLabels, newLabels map[string]string) error {
	bkt := tx.Bucket(bucketLabelIdx)
	if bkt == nil {
		return fmt.Errorf("labelidx bucket missing")
	}
	suffix := "\x00" + key.BucketPath() + "\x00" + key.StorageKey()

	for k, v := range oldLabels {
		if err := bkt.Delete([]byte(k + "=" + v + suffix)); err != nil {
			return err
		}
	}
	for k, v := range newLabels {
		if err := bkt.Put([]byte(k+"="+v+suffix), []byte{}); err != nil {
			return err
		}
	}
	return nil
}

// updateFieldIndex removes old field entries and adds new ones.
func updateFieldIndex(tx *bbolt.Tx, key ResourceKey, oldFields, newFields map[string]string) error {
	bkt := tx.Bucket(bucketFieldIdx)
	if bkt == nil {
		return fmt.Errorf("fieldidx bucket missing")
	}
	suffix := "\x00" + key.BucketPath() + "\x00" + key.StorageKey()

	for f, v := range oldFields {
		if err := bkt.Delete([]byte(f + "=" + v + suffix)); err != nil {
			return err
		}
	}
	for f, v := range newFields {
		if err := bkt.Put([]byte(f+"="+v+suffix), []byte{}); err != nil {
			return err
		}
	}
	return nil
}

// extractFields returns the indexed fields for a stored envelope.
func extractFields(env *Envelope) map[string]string {
	return map[string]string{
		"metadata.name":      env.Name,
		"metadata.namespace": env.Namespace,
	}
}

// populateEnvelopeFromObject parses the raw JSON object to extract metadata
// fields that are denormalised into the Envelope.
func populateEnvelopeFromObject(env *Envelope, obj []byte, key ResourceKey) error {
	// Use a minimal struct for fast parsing without allocating the full type.
	var meta struct {
		Metadata struct {
			Name        string            `json:"name"`
			Namespace   string            `json:"namespace"`
			Labels      map[string]string `json:"labels"`
			Finalizers  []string          `json:"finalizers"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(obj, &meta); err != nil {
		return err
	}
	env.Name = meta.Metadata.Name
	env.Namespace = meta.Metadata.Namespace
	env.Labels = meta.Metadata.Labels
	env.Finalizers = meta.Metadata.Finalizers

	// Fall back to key values if object doesn't set them.
	if env.Name == "" {
		env.Name = key.Name
	}
	if env.Namespace == "" {
		env.Namespace = key.Namespace
	}
	return nil
}

// injectServerFields merges server-assigned metadata into an object's JSON.
func injectServerFields(obj []byte, uid string, rv int64, createdAt, updatedAt time.Time) ([]byte, error) {
	return injectMetaFields(obj, uid, rv, createdAt, updatedAt, 1)
}

func injectServerFieldsUpdate(obj []byte, uid string, rv int64, createdAt, updatedAt time.Time) ([]byte, error) {
	return injectMetaFields(obj, uid, rv, createdAt, updatedAt, -1)
}

// injectMetaFields merges server-controlled metadata fields into a raw JSON object.
// generation = -1 means preserve existing generation.
func injectMetaFields(obj []byte, uid string, rv int64, createdAt, updatedAt time.Time, generation int64) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(obj, &raw); err != nil {
		return nil, err
	}

	meta, ok := raw["metadata"].(map[string]interface{})
	if !ok {
		meta = make(map[string]interface{})
		raw["metadata"] = meta
	}

	meta["uid"] = uid
	meta["resourceVersion"] = fmt.Sprintf("%d", rv)
	meta["creationTimestamp"] = createdAt.UTC().Format(time.RFC3339)

	if generation > 0 {
		meta["generation"] = generation
	}

	return json.Marshal(raw)
}

// ─── Binary encoding helpers ──────────────────────────────────────────────────

func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func bytesToUint64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}
