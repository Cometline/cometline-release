package process

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const fallbackPath = "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

// Env returns the current environment with a PATH suitable for subprocesses
// launched outside a user's interactive shell.
func Env() []string {
	path := AugmentedPath(os.Getenv("PATH"))
	env := os.Environ()
	for i, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			env[i] = "PATH=" + path
			return env
		}
	}
	return append(env, "PATH="+path)
}

// ResolveCommand finds command using the same augmented PATH passed to child processes.
func ResolveCommand(command string) (string, error) {
	if strings.ContainsRune(command, os.PathSeparator) {
		return command, nil
	}
	if path, err := exec.LookPath(command); err == nil {
		return path, nil
	}
	if path, ok := lookupInPath(command, AugmentedPath(os.Getenv("PATH"))); ok {
		return path, nil
	}
	return "", exec.ErrNotFound
}

// AugmentedPath preserves the inherited PATH and prepends common user tool locations.
func AugmentedPath(current string) string {
	entries := filepath.SplitList(current)
	if len(entries) == 0 {
		entries = filepath.SplitList(fallbackPath)
	}
	prefixes := []string{}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		prefixes = append(prefixes,
			filepath.Join(home, ".opencode", "bin"),
			filepath.Join(home, ".local", "bin"),
		)
	}
	if runtime.GOOS == "darwin" {
		prefixes = append(prefixes, "/opt/homebrew/bin", "/usr/local/bin")
	}
	entries = append(prefixes, entries...)
	entries = append(entries, filepath.SplitList(fallbackPath)...)

	seen := make(map[string]struct{}, len(entries))
	compact := entries[:0]
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		compact = append(compact, entry)
	}
	return strings.Join(compact, string(os.PathListSeparator))
}

func CommandNotFoundError(command string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return &exec.Error{Name: command, Err: exec.ErrNotFound}
	}
	return err
}

func lookupInPath(command, pathValue string) (string, bool) {
	for _, dir := range filepath.SplitList(pathValue) {
		candidate := filepath.Join(dir, command)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
			continue
		}
		return candidate, true
	}
	return "", false
}
