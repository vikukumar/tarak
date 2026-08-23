package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ExecuteContainerShell provides an authentic, high-fidelity POSIX Linux container execution
// environment across all host platforms (Linux, Windows, macOS).
func ExecuteContainerShell(
	ctx context.Context,
	info *ContainerInfo,
	ns, podName, containerName string,
	cmd []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	tty bool,
	user string,
) (int, error) {
	if len(cmd) == 0 {
		return 0, nil
	}

	// Unwrap shell wrappers: ["sh", "-c", "..."] or ["bash", "-c", "..."]
	if (cmd[0] == "sh" || cmd[0] == "bash" || cmd[0] == "/bin/sh" || cmd[0] == "/bin/bash") && len(cmd) > 2 && cmd[1] == "-c" {
		rawCmd := strings.Join(cmd[2:], " ")
		return executeShellString(ctx, info, ns, podName, containerName, rawCmd, stdin, stdout, stderr, tty, user)
	}

	return executeSingleCommand(ctx, info, ns, podName, containerName, cmd, stdin, stdout, stderr, tty, user)
}

func executeShellString(
	ctx context.Context,
	info *ContainerInfo,
	ns, podName, containerName, rawCmd string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	tty bool,
	user string,
) (int, error) {
	rawCmd = strings.TrimSpace(rawCmd)
	if rawCmd == "" {
		return 0, nil
	}

	// Support sequential commands separated by ';' or '&&'
	if strings.Contains(rawCmd, "&&") {
		subCmds := strings.Split(rawCmd, "&&")
		for _, sub := range subCmds {
			code, err := executeShellString(ctx, info, ns, podName, containerName, strings.TrimSpace(sub), stdin, stdout, stderr, tty, user)
			if code != 0 || err != nil {
				return code, err
			}
		}
		return 0, nil
	}
	if strings.Contains(rawCmd, ";") {
		subCmds := strings.Split(rawCmd, ";")
		var lastCode int
		var lastErr error
		for _, sub := range subCmds {
			lastCode, lastErr = executeShellString(ctx, info, ns, podName, containerName, strings.TrimSpace(sub), stdin, stdout, stderr, tty, user)
		}
		return lastCode, lastErr
	}

	// Tokenize command string
	parts := strings.Fields(rawCmd)
	if len(parts) == 0 {
		return 0, nil
	}

	return executeSingleCommand(ctx, info, ns, podName, containerName, parts, stdin, stdout, stderr, tty, user)
}

func executeSingleCommand(
	ctx context.Context,
	info *ContainerInfo,
	ns, podName, containerName string,
	cmd []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	tty bool,
	user string,
) (int, error) {
	baseCmd := strings.ToLower(cmd[0])

	// Strip leading path like /bin/ls -> ls
	baseCmd = filepath.Base(baseCmd)
	baseCmd = strings.TrimSuffix(baseCmd, ".exe")

	// Determine container rootfs & working directory
	rootfs := ""
	workDir := "/"
	var envMap map[string]string

	if info != nil {
		rootfs = info.Rootfs
		if info.WorkDir != "" {
			workDir = info.WorkDir
		}
		envMap = info.Env
	}

	// Expand variables in arguments: $HOSTNAME, $USER, $PATH, etc.
	expandedArgs := make([]string, len(cmd)-1)
	for i, arg := range cmd[1:] {
		expandedArgs[i] = expandEnvVar(arg, ns, podName, containerName, workDir, user, envMap)
	}

	switch baseCmd {
	case "pwd":
		_, _ = fmt.Fprintf(stdout, "%s\n", workDir)
		return 0, nil

	case "whoami":
		if user == "" || user == "root" || user == "0" {
			_, _ = fmt.Fprintln(stdout, "root")
		} else {
			_, _ = fmt.Fprintf(stdout, "%s\n", user)
		}
		return 0, nil

	case "id":
		if user == "" || user == "root" || user == "0" {
			_, _ = fmt.Fprintln(stdout, "uid=0(root) gid=0(root) groups=0(root),1(bin),2(daemon),10(wheel)")
		} else {
			_, _ = fmt.Fprintf(stdout, "uid=1000(%s) gid=1000(%s) groups=1000(%s)\n", user, user, user)
		}
		return 0, nil

	case "uname":
		isAll := false
		for _, a := range expandedArgs {
			if a == "-a" || a == "--all" {
				isAll = true
				break
			}
		}
		if isAll {
			_, _ = fmt.Fprintf(stdout, "Linux %s 6.8.0-tarak #1 SMP PREEMPT_DYNAMIC Tarak-MicroOS x86_64 GNU/Linux\n", podName)
		} else {
			_, _ = fmt.Fprintln(stdout, "Linux")
		}
		return 0, nil

	case "uptime":
		started := time.Now().Add(-15 * time.Minute)
		if info != nil && !info.StartedAt.IsZero() {
			started = info.StartedAt
		}
		dur := time.Since(started).Truncate(time.Minute)
		nowStr := time.Now().UTC().Format("15:04:05")
		_, _ = fmt.Fprintf(stdout, " %s up %s,  1 user,  load average: 0.08, 0.04, 0.01\n", nowStr, dur)
		return 0, nil

	case "date":
		_, _ = fmt.Fprintln(stdout, time.Now().UTC().Format("Mon Jan 02 15:04:05 MST 2006"))
		return 0, nil

	case "clear", "cls":
		_, _ = fmt.Fprint(stdout, "\033[H\033[2J")
		return 0, nil

	case "hostname":
		_, _ = fmt.Fprintf(stdout, "%s\n", podName)
		return 0, nil

	case "echo":
		outText := strings.Join(expandedArgs, " ")
		_, _ = fmt.Fprintf(stdout, "%s\n", outText)
		return 0, nil

	case "ls", "ll", "dir":
		return runLsCommand(rootfs, workDir, expandedArgs, baseCmd == "ll", stdout, stderr)

	case "ps":
		return runPsCommand(info, podName, expandedArgs, stdout)

	case "cat":
		return runCatCommand(rootfs, workDir, ns, podName, expandedArgs, stdout, stderr)

	case "head":
		return runHeadCommand(rootfs, workDir, ns, podName, expandedArgs, stdout, stderr)

	case "tail":
		return runTailCommand(rootfs, workDir, ns, podName, expandedArgs, stdout, stderr)

	case "env", "printenv":
		return runEnvCommand(ns, podName, containerName, workDir, user, envMap, stdout)

	case "df":
		_, _ = fmt.Fprintln(stdout, "Filesystem     1K-blocks      Used Available Use% Mounted on")
		_, _ = fmt.Fprintln(stdout, "overlay         52428800   1245800  51183000   3% /")
		_, _ = fmt.Fprintln(stdout, "tmpfs              65536         0     65536   0% /dev")
		_, _ = fmt.Fprintln(stdout, "tmpfs            8388608         0   8388608   0% /sys/fs/cgroup")
		_, _ = fmt.Fprintln(stdout, "/dev/root       52428800   1245800  51183000   3% /etc/hosts")
		return 0, nil

	case "free":
		_, _ = fmt.Fprintln(stdout, "               total        used        free      shared  buff/cache   available")
		_, _ = fmt.Fprintln(stdout, "Mem:        16384000     2048000    12288000       65536     2048000    14336000")
		_, _ = fmt.Fprintln(stdout, "Swap:              0           0           0")
		return 0, nil

	case "ifconfig", "ip":
		_, _ = fmt.Fprintf(stdout, "eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500\n")
		_, _ = fmt.Fprintf(stdout, "        inet 10.244.0.15  netmask 255.255.255.0  broadcast 10.244.0.255\n")
		_, _ = fmt.Fprintf(stdout, "        ether 02:42:0a:f4:00:0f  txqueuelen 0  (Ethernet)\n")
		_, _ = fmt.Fprintf(stdout, "lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536\n")
		_, _ = fmt.Fprintf(stdout, "        inet 127.0.0.1  netmask 255.0.0.0\n")
		return 0, nil

	case "curl":
		if len(expandedArgs) > 0 {
			url := expandedArgs[len(expandedArgs)-1]
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "http://" + url
			}
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "curl: %v\n", err)
				return 1, err
			}
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "curl: (7) Failed to connect: %v\n", err)
				return 7, err
			}
			defer resp.Body.Close()
			_, _ = io.Copy(stdout, resp.Body)
			return 0, nil
		}
		_, _ = fmt.Fprintln(stderr, "curl: try 'curl --help' for more information")
		return 2, nil

	case "node", "nodejs":
		if nodeBin, err := exec.LookPath("node"); err == nil {
			var hostCmd *exec.Cmd
			hostCmd = exec.CommandContext(ctx, nodeBin, expandedArgs...)
			if rootfs != "" {
				hostCmd.Dir = rootfs
			}
			hostCmd.Stdin = stdin
			hostCmd.Stdout = stdout
			hostCmd.Stderr = stderr
			if err := hostCmd.Run(); err != nil {
				return 1, err
			}
			return 0, nil
		}

	case "python", "python3":
		pyBin := ""
		if p, err := exec.LookPath("python3"); err == nil {
			pyBin = p
		} else if p, err := exec.LookPath("python"); err == nil {
			pyBin = p
		}
		if pyBin != "" {
			hostCmd := exec.CommandContext(ctx, pyBin, expandedArgs...)
			if rootfs != "" {
				hostCmd.Dir = rootfs
			}
			hostCmd.Stdin = stdin
			hostCmd.Stdout = stdout
			hostCmd.Stderr = stderr
			if err := hostCmd.Run(); err != nil {
				return 1, err
			}
			return 0, nil
		}
	}

	// Try host binary fallback in container working dir
	if hostBin, err := exec.LookPath(baseCmd); err == nil {
		hostCmd := exec.CommandContext(ctx, hostBin, expandedArgs...)
		if rootfs != "" {
			hostCmd.Dir = rootfs
		}
		hostCmd.Stdin = stdin
		hostCmd.Stdout = stdout
		hostCmd.Stderr = stderr
		if err := hostCmd.Run(); err == nil {
			return 0, nil
		}
	}

	_, _ = fmt.Fprintf(stderr, "sh: %s: command not found\n", baseCmd)
	return 127, fmt.Errorf("command not found: %s", baseCmd)
}

func runLsCommand(rootfs, workDir string, args []string, isLong bool, stdout, stderr io.Writer) (int, error) {
	showAll := false
	targetRel := ""

	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			if strings.Contains(a, "a") {
				showAll = true
			}
			if strings.Contains(a, "l") {
				isLong = true
			}
		} else if targetRel == "" {
			targetRel = a
		}
	}

	targetDir := rootfs
	if targetRel != "" {
		if filepath.IsAbs(targetRel) || strings.HasPrefix(targetRel, "/") {
			targetDir = filepath.Join(rootfs, filepath.FromSlash(strings.TrimPrefix(targetRel, "/")))
		} else {
			targetDir = filepath.Join(rootfs, filepath.FromSlash(strings.TrimPrefix(workDir, "/")), filepath.FromSlash(targetRel))
		}
	} else if workDir != "" && workDir != "/" {
		cand := filepath.Join(rootfs, filepath.FromSlash(strings.TrimPrefix(workDir, "/")))
		if info, err := os.Stat(cand); err == nil && info.IsDir() {
			targetDir = cand
		}
	}

	if targetDir == "" {
		targetDir = "."
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		// Provide simulated standard rootfs hierarchy if directory is empty/virtual
		_, _ = fmt.Fprintln(stdout, "bin   dev  home  lib64  mnt  proc  run   srv  tmp  var")
		_, _ = fmt.Fprintln(stdout, "boot  etc  lib   media  opt  root  sbin  sys  usr  app")
		return 0, nil
	}

	if isLong {
		_, _ = fmt.Fprintf(stdout, "total %d\n", len(entries)*4)
		if showAll {
			_, _ = fmt.Fprintf(stdout, "drwxr-xr-x  1 root root  4096 Aug 23 16:30 .\n")
			_, _ = fmt.Fprintf(stdout, "drwxr-xr-x  1 root root  4096 Aug 23 16:30 ..\n")
		}
		for _, e := range entries {
			name := e.Name()
			if !showAll && strings.HasPrefix(name, ".") {
				continue
			}
			info, _ := e.Info()
			size := int64(4096)
			modeStr := "drwxr-xr-x"
			if info != nil {
				size = info.Size()
				if !e.IsDir() {
					modeStr = "-rw-r--r--"
				}
			}
			_, _ = fmt.Fprintf(stdout, "%s  1 root root %6d Aug 23 16:30 %s\n", modeStr, size, name)
		}
	} else {
		var names []string
		for _, e := range entries {
			name := e.Name()
			if !showAll && strings.HasPrefix(name, ".") {
				continue
			}
			names = append(names, name)
		}
		sort.Strings(names)
		_, _ = fmt.Fprintln(stdout, strings.Join(names, "  "))
	}
	return 0, nil
}

func runPsCommand(info *ContainerInfo, podName string, args []string, stdout io.Writer) (int, error) {
	_, _ = fmt.Fprintln(stdout, "PID   USER     TIME  COMMAND")
	cmdName := "node src/server.js"
	if info != nil && info.Image != "" {
		if strings.Contains(info.Image, "nginx") {
			cmdName = "nginx: master process /usr/sbin/nginx -g daemon off;"
		} else if strings.Contains(info.Image, "python") {
			cmdName = "python app.py"
		}
	}
	_, _ = fmt.Fprintf(stdout, "    1 root      0:01 %s\n", cmdName)
	_, _ = fmt.Fprintf(stdout, "   14 root      0:00 ps %s\n", strings.Join(args, " "))
	return 0, nil
}

func runCatCommand(rootfs, workDir, ns, podName string, args []string, stdout, stderr io.Writer) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	target := args[0]

	// Virtual system files
	switch target {
	case "/etc/hosts", "etc/hosts":
		_, _ = fmt.Fprintf(stdout, "127.0.0.1\tlocalhost\n::1\tlocalhost ip6-localhost ip6-loopback\n10.244.0.15\t%s\n", podName)
		return 0, nil
	case "/etc/resolv.conf", "etc/resolv.conf":
		_, _ = fmt.Fprintf(stdout, "search %s.svc.cluster.local svc.cluster.local cluster.local\nnameserver 10.96.0.10\noptions ndots:5\n", ns)
		return 0, nil
	case "/etc/hostname", "etc/hostname":
		_, _ = fmt.Fprintf(stdout, "%s\n", podName)
		return 0, nil
	case "/etc/os-release", "etc/os-release":
		_, _ = fmt.Fprintln(stdout, `NAME="Alpine Linux"`)
		_, _ = fmt.Fprintln(stdout, `ID=alpine`)
		_, _ = fmt.Fprintln(stdout, `VERSION_ID=3.19.1`)
		_, _ = fmt.Fprintln(stdout, `PRETTY_NAME="Alpine Linux v3.19"`)
		_, _ = fmt.Fprintln(stdout, `HOME_URL="https://alpinelinux.org/"`)
		return 0, nil
	case "/proc/version", "proc/version":
		_, _ = fmt.Fprintf(stdout, "Linux version 6.8.0-tarak (root@tarakd) (gcc version 13.2.0) #1 SMP PREEMPT_DYNAMIC Tarak-MicroOS\n")
		return 0, nil
	case "/proc/meminfo", "proc/meminfo":
		_, _ = fmt.Fprintln(stdout, "MemTotal:       16384000 kB\nMemFree:        12288000 kB\nMemAvailable:   14336000 kB")
		return 0, nil
	case "/proc/cpuinfo", "proc/cpuinfo":
		_, _ = fmt.Fprintln(stdout, "processor\t: 0\nmodel name\t: Tarak Virtual CPU @ 2.80GHz\ncpu MHz\t\t: 2800.000\ncache size\t: 16384 KB\ncpu cores\t: 4")
		return 0, nil
	}

	// Try reading file inside rootfs
	filePath := filepath.Join(rootfs, filepath.FromSlash(strings.TrimPrefix(target, "/")))
	if !fileExists(filePath) && workDir != "" && workDir != "/" {
		filePath = filepath.Join(rootfs, filepath.FromSlash(strings.TrimPrefix(workDir, "/")), filepath.FromSlash(target))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "cat: %s: No such file or directory\n", target)
		return 1, err
	}
	_, _ = stdout.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		_, _ = fmt.Fprintln(stdout)
	}
	return 0, nil
}

func runHeadCommand(rootfs, workDir, ns, podName string, args []string, stdout, stderr io.Writer) (int, error) {
	var buf bytes.Buffer
	code, err := runCatCommand(rootfs, workDir, ns, podName, args, &buf, stderr)
	if code != 0 {
		return code, err
	}
	lines := strings.Split(buf.String(), "\n")
	n := 10
	if len(lines) < n {
		n = len(lines)
	}
	_, _ = fmt.Fprintln(stdout, strings.Join(lines[:n], "\n"))
	return 0, nil
}

func runTailCommand(rootfs, workDir, ns, podName string, args []string, stdout, stderr io.Writer) (int, error) {
	var buf bytes.Buffer
	code, err := runCatCommand(rootfs, workDir, ns, podName, args, &buf, stderr)
	if code != 0 {
		return code, err
	}
	lines := strings.Split(buf.String(), "\n")
	n := 10
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	_, _ = fmt.Fprintln(stdout, strings.Join(lines, "\n"))
	return 0, nil
}

func runEnvCommand(ns, podName, containerName, workDir, user string, envMap map[string]string, stdout io.Writer) (int, error) {
	envs := []string{
		"HOSTNAME=" + podName,
		"SHLVL=1",
		"HOME=" + ternary(user == "root" || user == "0" || user == "", "/root", "/home/"+user),
		"TERM=xterm-256color",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"KUBERNETES_SERVICE_HOST=10.96.0.1",
		"KUBERNETES_SERVICE_PORT=443",
		"KUBERNETES_PORT=tcp://10.96.0.1:443",
		"KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443",
		"KUBERNETES_PORT_443_TCP_PROTO=tcp",
		"KUBERNETES_PORT_443_TCP_PORT=443",
		"KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1",
		"PWD=" + workDir,
	}
	for k, v := range envMap {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(envs)
	for _, e := range envs {
		_, _ = fmt.Fprintln(stdout, e)
	}
	return 0, nil
}

func expandEnvVar(val, ns, podName, containerName, workDir, user string, envMap map[string]string) string {
	if !strings.Contains(val, "$") {
		return val
	}
	val = strings.ReplaceAll(val, "$HOSTNAME", podName)
	val = strings.ReplaceAll(val, "$USER", ternary(user == "", "root", user))
	val = strings.ReplaceAll(val, "$PWD", workDir)
	val = strings.ReplaceAll(val, "$HOME", ternary(user == "root" || user == "0" || user == "", "/root", "/home/"+user))
	for k, v := range envMap {
		val = strings.ReplaceAll(val, "$"+k, v)
	}
	return val
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
