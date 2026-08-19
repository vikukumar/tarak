// Package statestore — label/field index helpers.
//
// This file contains the index query logic. The actual index mutation
// (building index entries in BoltDB) lives in bolt.go.
package statestore

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// generateUID creates a new random UID in the Kubernetes UID format
// (8-4-4-4-12 hex characters separated by dashes).
func generateUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen on any real OS.
		panic("statestore: failed to generate random UID: " + err.Error())
	}
	s := hex.EncodeToString(b)
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32]
}

// ─── Index query helpers ───────────────────────────────────────────────────────

// labelIndexPrefix returns the BoltDB key prefix for scanning all resources
// with a specific label key=value pair.
//
// Index key format:
//
//	"<label_key>=<label_val>\x00<bucket_path>\x00<storage_key>"
func labelIndexPrefix(labelKey, labelVal string) string {
	return labelKey + "=" + labelVal + "\x00"
}

// parseLabelIndexKey extracts the bucket path and storage key from an index entry.
func parseLabelIndexKey(indexKey string) (bucketPath, storageKey string, ok bool) {
	parts := strings.SplitN(indexKey, "\x00", 3)
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// fieldIndexPrefix returns the BoltDB key prefix for scanning by field=value.
func fieldIndexPrefix(field, value string) string {
	return field + "=" + value + "\x00"
}

// parseFieldIndexKey extracts bucket path and storage key from a field index entry.
func parseFieldIndexKey(indexKey string) (bucketPath, storageKey string, ok bool) {
	parts := strings.SplitN(indexKey, "\x00", 3)
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[1], parts[2], true
}
