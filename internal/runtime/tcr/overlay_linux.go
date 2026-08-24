//go:build linux

package tcr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// OverlayMount configures a copy-on-write OverlayFS rootfs mount.
type OverlayMount struct {
	MergedDir string
	UpperDir  string
	WorkDir   string
	LowerDirs []string
}

// MountOverlay creates upper/work/merged directories and mounts OverlayFS.
func MountOverlay(baseDir, containerID string, lowerLayers []string) (*OverlayMount, error) {
	cDir := filepath.Join(baseDir, "overlay", containerID)
	upper := filepath.Join(cDir, "upper")
	work := filepath.Join(cDir, "work")
	merged := filepath.Join(cDir, "merged")

	for _, d := range []string{upper, work, merged} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("create overlay dir %s: %w", d, err)
		}
	}

	lowerStr := strings.Join(lowerLayers, ":")
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerStr, upper, work)

	if err := syscall.Mount("overlay", merged, "overlay", 0, opts); err != nil {
		return nil, fmt.Errorf("mount overlayfs (%s): %w", opts, err)
	}

	return &OverlayMount{
		MergedDir: merged,
		UpperDir:  upper,
		WorkDir:   work,
		LowerDirs: lowerLayers,
	}, nil
}

// UnmountOverlay unmounts and cleans up an OverlayFS mount.
func (om *OverlayMount) UnmountOverlay() error {
	if om.MergedDir == "" {
		return nil
	}
	_ = syscall.Unmount(om.MergedDir, syscall.MNT_DETACH)
	return nil
}
