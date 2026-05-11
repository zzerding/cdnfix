package workflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTaskStateLifecycle(t *testing.T) {
	paths := Paths{
		LogDir:   filepath.Join(t.TempDir(), "logs"),
		CacheDir: filepath.Join(t.TempDir(), "cache"),
		RunDir:   filepath.Join(t.TempDir(), "runs"),
	}
	cacheFile := TaskStatePath(paths, "prod-a", "refresh")

	task := TaskRecord{
		ID:          "task-1",
		Action:      "refresh",
		Status:      "pending",
		RunID:       "run-1",
		URLCount:    2,
		SubmittedAt: time.Now().Format(time.RFC3339),
	}
	if err := AddTask(cacheFile, "prod-a", "refresh", task); err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	pending, err := PendingTasks(cacheFile, "prod-a", "refresh")
	if err != nil {
		t.Fatalf("PendingTasks failed: %v", err)
	}
	if len(pending) != 1 || pending[0].ID != "task-1" {
		t.Fatalf("unexpected pending tasks: %+v", pending)
	}

	if err := MarkTask(cacheFile, "prod-a", "refresh", "task-1", "completed", time.Now()); err != nil {
		t.Fatalf("MarkTask failed: %v", err)
	}

	pending, err = PendingTasks(cacheFile, "prod-a", "refresh")
	if err != nil {
		t.Fatalf("PendingTasks failed: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending tasks, got %+v", pending)
	}
}

func TestReadJobs(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	manifest := filepath.Join(configDir, "jobs.yaml")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	content := []byte("jobs:\n  - name: a-refresh\n    site: prod-a\n    action: refresh\n    file: ../urls/prod-a/refresh.txt\n")
	if err := os.WriteFile(manifest, content, 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	jobs, err := ReadJobs(manifest)
	if err != nil {
		t.Fatalf("ReadJobs failed: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Site != "prod-a" || jobs[0].Action != "refresh" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
	expectedFile := filepath.Join(root, "urls", "prod-a", "refresh.txt")
	if jobs[0].File != expectedFile {
		t.Fatalf("unexpected job file path: got %q want %q", jobs[0].File, expectedFile)
	}
}

func TestResolveLayoutDefaultsToExecutableDir(t *testing.T) {
	root := t.TempDir()
	execPath := filepath.Join(root, "cdnfix")
	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	siteConfig := filepath.Join(configDir, "sites.yaml")
	if err := os.WriteFile(siteConfig, []byte("sites: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write sites config: %v", err)
	}

	layout, err := ResolveLayout(execPath, LayoutOptions{})
	if err != nil {
		t.Fatalf("ResolveLayout failed: %v", err)
	}

	if layout.RootDir != root {
		t.Fatalf("unexpected root dir: got %q want %q", layout.RootDir, root)
	}
	if layout.ConfigDir != configDir {
		t.Fatalf("unexpected config dir: got %q want %q", layout.ConfigDir, configDir)
	}
	if layout.SiteConfig != siteConfig {
		t.Fatalf("unexpected site config: got %q want %q", layout.SiteConfig, siteConfig)
	}
	if layout.Manifest != filepath.Join(configDir, "jobs.yaml") {
		t.Fatalf("unexpected manifest path: %q", layout.Manifest)
	}
	if layout.Paths.LogDir != filepath.Join(root, "var", "logs") {
		t.Fatalf("unexpected log dir: %q", layout.Paths.LogDir)
	}
	if layout.Paths.CacheDir != filepath.Join(root, "var", "cache") {
		t.Fatalf("unexpected cache dir: %q", layout.Paths.CacheDir)
	}
	if layout.Paths.RunDir != filepath.Join(root, "var", "runs") {
		t.Fatalf("unexpected run dir: %q", layout.Paths.RunDir)
	}
}

func TestResolveLayoutResolvesOverridesAgainstRoot(t *testing.T) {
	root := t.TempDir()
	execPath := filepath.Join(root, "bin", "cdnfix")

	layout, err := ResolveLayout(execPath, LayoutOptions{
		RootDir:    "..",
		ConfigDir:  "settings",
		SiteConfig: "sites.prod.yaml",
		Manifest:   "jobs/prod.yaml",
		LogDir:     "runtime/logs",
		CacheDir:   "runtime/cache",
		RunDir:     "runtime/runs",
	})
	if err != nil {
		t.Fatalf("ResolveLayout failed: %v", err)
	}

	expectedRoot := filepath.Clean(root)
	if layout.RootDir != expectedRoot {
		t.Fatalf("unexpected root dir: got %q want %q", layout.RootDir, expectedRoot)
	}
	if layout.ConfigDir != filepath.Join(expectedRoot, "settings") {
		t.Fatalf("unexpected config dir: %q", layout.ConfigDir)
	}
	if layout.SiteConfig != filepath.Join(expectedRoot, "sites.prod.yaml") {
		t.Fatalf("unexpected site config: %q", layout.SiteConfig)
	}
	if layout.Manifest != filepath.Join(expectedRoot, "jobs", "prod.yaml") {
		t.Fatalf("unexpected manifest: %q", layout.Manifest)
	}
	if layout.Paths.LogDir != filepath.Join(expectedRoot, "runtime", "logs") {
		t.Fatalf("unexpected log dir: %q", layout.Paths.LogDir)
	}
	if layout.Paths.CacheDir != filepath.Join(expectedRoot, "runtime", "cache") {
		t.Fatalf("unexpected cache dir: %q", layout.Paths.CacheDir)
	}
	if layout.Paths.RunDir != filepath.Join(expectedRoot, "runtime", "runs") {
		t.Fatalf("unexpected run dir: %q", layout.Paths.RunDir)
	}
}
