//go:build linux

package tcr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// RunContainerInit is called when the tarak binary is re-executed with
// __tcr_init__ as argv[1]. It runs inside freshly created Linux namespaces
// (created by the parent via SysProcAttr.Cloneflags).
//
// Sequence:
//  1. Read ContainerConfig from TARAK_TCR_CONFIG env var
//  2. Make the mount namespace private (so mounts don't leak to host)
//  3. Mount /proc, /sys, /dev inside the container rootfs
//  4. Set the container hostname
//  5. chroot into the rootfs
//  6. chdir to the working directory
//  7. syscall.Exec the actual container process (replaces this process)
//
// After step 7, the container process runs as PID 1 inside its own PID
// namespace with full filesystem isolation — no Docker needed.
func RunContainerInit() error {
	cfgJSON := os.Getenv(EnvConfig)
	if cfgJSON == "" {
		return fmt.Errorf("TARAK_TCR_CONFIG not set")
	}

	cfg, err := UnmarshalConfig(cfgJSON)
	if err != nil {
		return fmt.Errorf("parse container config: %w", err)
	}

	if len(cfg.Command) == 0 {
		return fmt.Errorf("empty command in container config")
	}

	rootfs := cfg.Rootfs
	if rootfs == "" {
		return fmt.Errorf("empty rootfs in container config")
	}

	// ── 1. Make the mount namespace private ──────────────────────────────────
	// Without this, mounts propagate to the host mount namespace.
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		// Non-fatal on some kernel configs — log and continue
		fmt.Fprintf(os.Stderr, "tcr-init: MS_PRIVATE: %v (non-fatal)\n", err)
	}

	// ── 2. Mount /proc inside rootfs ─────────────────────────────────────────
	procDir := filepath.Join(rootfs, "proc")
	if mkErr := os.MkdirAll(procDir, 0755); mkErr == nil {
		if mntErr := syscall.Mount("proc", procDir, "proc", 0, ""); mntErr != nil {
			fmt.Fprintf(os.Stderr, "tcr-init: mount proc: %v\n", mntErr)
		}
	}

	// ── 3. Mount /sys inside rootfs ──────────────────────────────────────────
	sysDir := filepath.Join(rootfs, "sys")
	if mkErr := os.MkdirAll(sysDir, 0755); mkErr == nil {
		if mntErr := syscall.Mount("sysfs", sysDir, "sysfs", syscall.MS_RDONLY, ""); mntErr != nil {
			fmt.Fprintf(os.Stderr, "tcr-init: mount sys: %v\n", mntErr)
		}
	}

	// ── 4. Mount /dev inside rootfs ──────────────────────────────────────────
	devDir := filepath.Join(rootfs, "dev")
	if mkErr := os.MkdirAll(devDir, 0755); mkErr == nil {
		// Try devtmpfs first, fall back to bind-mounting /dev from host
		if mntErr := syscall.Mount("devtmpfs", devDir, "devtmpfs", 0, ""); mntErr != nil {
			// Bind-mount host /dev as read-write
			_ = syscall.Mount("/dev", devDir, "", syscall.MS_BIND|syscall.MS_REC, "")
		}
	}

	// Create /dev/pts for terminal support
	devptsDir := filepath.Join(rootfs, "dev", "pts")
	if mkErr := os.MkdirAll(devptsDir, 0755); mkErr == nil {
		_ = syscall.Mount("devpts", devptsDir, "devpts", 0, "newinstance,ptmxmode=0666")
	}

	// ── 5. Set hostname ───────────────────────────────────────────────────────
	hostname := cfg.Hostname
	if hostname == "" {
		hostname = cfg.ID
	}
	if hsErr := syscall.Sethostname([]byte(hostname)); hsErr != nil {
		fmt.Fprintf(os.Stderr, "tcr-init: sethostname: %v\n", hsErr)
	}

	// ── 6. chroot into rootfs ─────────────────────────────────────────────────
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot %q: %w", rootfs, err)
	}

	// ── 7. chdir to working directory ─────────────────────────────────────────
	workDir := cfg.WorkingDir
	if workDir == "" {
		workDir = "/"
	}
	if err := syscall.Chdir(workDir); err != nil {
		// Fall back to root if working dir doesn't exist
		_ = syscall.Chdir("/")
	}

	// ── 8. exec the container process ────────────────────────────────────────
	// Resolve the binary path inside the chroot. Since we've already chrooted,
	// standard PATH resolution applies within the container filesystem.
	binary := cfg.Command[0]
	if !strings.HasPrefix(binary, "/") {
		// Search PATH inside the container
		for _, dir := range []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"} {
			candidate := filepath.Join(dir, binary)
			if _, err := os.Stat(candidate); err == nil {
				binary = candidate
				break
			}
		}
	}

	execErr := syscall.Exec(binary, cfg.Command, cfg.Env)
	if execErr != nil {
		// Fallback: try via /bin/sh or /bin/bash
		for _, sh := range []string{"/bin/sh", "/bin/bash", "/usr/bin/sh", "/usr/bin/bash"} {
			if _, statErr := os.Stat(sh); statErr == nil {
				shCmd := []string{sh, "-c", strings.Join(cfg.Command, " ")}
				_ = syscall.Exec(sh, shCmd, cfg.Env)
			}
		}
	}
	return execErr
}

// RunContainerExec is called when the tarak binary is re-executed with
// __tcr_exec__ as argv[1]. It enters the container's mount namespace
// (via /proc/<pid>/ns/mnt) and exec's the requested command.
func RunContainerExec() error {
	cfgJSON := os.Getenv(EnvExecConfig)
	if cfgJSON == "" {
		return fmt.Errorf("TARAK_TCR_EXEC_CONFIG not set")
	}

	cfg, err := UnmarshalExecConfig(cfgJSON)
	if err != nil {
		return fmt.Errorf("parse exec config: %w", err)
	}

	if len(cfg.Command) == 0 {
		return fmt.Errorf("empty exec command")
	}

	// Enter the container's mount namespace via /proc/<pid>/ns/mnt
	mntNSPath := fmt.Sprintf("/proc/%d/ns/mnt", cfg.TargetPID)
	f, err := os.Open(mntNSPath)
	if err != nil {
		// Container may have exited or we don't have permission
		goto fallback
	}

	// setns — enter the mount namespace using SYS_SETNS directly
	{
		_, _, errno := syscall.RawSyscall(setns, f.Fd(), syscall.CLONE_NEWNS, 0)
		if errno != 0 {
			f.Close()
			fmt.Fprintf(os.Stderr, "tcr-exec: setns mnt: %v, falling back to chroot\n", errno)
			goto fallback
		}
	}
	f.Close()

	// Also enter UTS namespace for correct hostname
	if utsF, openErr := os.Open(fmt.Sprintf("/proc/%d/ns/uts", cfg.TargetPID)); openErr == nil {
		_, _, _ = syscall.RawSyscall(setns, utsF.Fd(), syscall.CLONE_NEWUTS, 0)
		utsF.Close()
	}

	if cfg.WorkingDir != "" {
		_ = syscall.Chdir(cfg.WorkingDir)
	} else {
		_ = syscall.Chdir("/")
	}

	return syscall.Exec(cfg.Command[0], cfg.Command, cfg.Env)

fallback:
	if cfg.Rootfs != "" {
		_ = syscall.Chroot(cfg.Rootfs)
	}
	if cfg.WorkingDir != "" {
		_ = syscall.Chdir(cfg.WorkingDir)
	} else {
		_ = syscall.Chdir("/")
	}
	return syscall.Exec(cfg.Command[0], cfg.Command, cfg.Env)
}

// setns is the Linux SYS_SETNS syscall number.
const setns uintptr = 308 // SYS_SETNS on amd64; also valid on arm64 (274) handled by kernel
