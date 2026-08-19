//go:build !windows

package runtime

func getSystemMemory(total, used, avail *uint64, pct *float64) {
	// Fallback logic in metrics_sys.go will handle this for now.
}
