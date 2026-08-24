//go:build !linux

package tcr

// OverlayMount stub for non-Linux platforms.
type OverlayMount struct {
	MergedDir string
	UpperDir  string
	WorkDir   string
	LowerDirs []string
}

func MountOverlay(baseDir, containerID string, lowerLayers []string) (*OverlayMount, error) {
	if len(lowerLayers) > 0 {
		return &OverlayMount{MergedDir: lowerLayers[0]}, nil
	}
	return &OverlayMount{MergedDir: baseDir}, nil
}

func (om *OverlayMount) UnmountOverlay() error {
	return nil
}
