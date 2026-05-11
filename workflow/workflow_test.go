package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func TestSelectBatchJobsWithoutFilterReturnsSortedJobs(t *testing.T) {
	jobs := []Job{
		{Name: "z-job", Site: "prod-b", Action: "push"},
		{Name: "b-job", Site: "prod-a", Action: "refresh"},
		{Name: "a-job", Site: "prod-a", Action: "push"},
	}

	selected, err := SelectBatchJobs(jobs, ExecuteBatchOptions{})
	if err != nil {
		t.Fatalf("SelectBatchJobs failed: %v", err)
	}

	if len(selected) != 3 {
		t.Fatalf("unexpected selected job count: %d", len(selected))
	}
	if selected[0].Name != "a-job" || selected[1].Name != "b-job" || selected[2].Name != "z-job" {
		t.Fatalf("jobs not sorted as expected: %+v", selected)
	}
}

func TestSelectBatchJobsWithExactMatch(t *testing.T) {
	jobs := []Job{
		{Name: "prod-a-refresh", Site: "prod-a", Action: "refresh"},
		{Name: "prod-b-push", Site: "prod-b", Action: "push"},
	}

	selected, err := SelectBatchJobs(jobs, ExecuteBatchOptions{JobName: "prod-b-push"})
	if err != nil {
		t.Fatalf("SelectBatchJobs failed: %v", err)
	}

	if len(selected) != 1 || selected[0].Name != "prod-b-push" {
		t.Fatalf("unexpected selected jobs: %+v", selected)
	}
}

func TestSelectBatchJobsReturnsErrorWhenMissing(t *testing.T) {
	jobs := []Job{
		{Name: "prod-a-refresh", Site: "prod-a", Action: "refresh"},
	}

	_, err := SelectBatchJobs(jobs, ExecuteBatchOptions{JobName: "missing-job"})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
	if err.Error() != `job "missing-job" not found in manifest` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectBatchJobsReturnsErrorWhenAmbiguous(t *testing.T) {
	jobs := []Job{
		{Name: "shared-job", Site: "prod-a", Action: "refresh"},
		{Name: "shared-job", Site: "prod-b", Action: "push"},
	}

	_, err := SelectBatchJobs(jobs, ExecuteBatchOptions{JobName: "shared-job"})
	if err == nil {
		t.Fatal("expected error for ambiguous job")
	}
	if err.Error() != `job "shared-job" is ambiguous: 2 matching jobs in manifest` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveLayoutUsesSystemDefaultsWithoutRoot(t *testing.T) {
	t.Setenv("CDNFIX_ROOT", "")
	t.Setenv("CDNFIX_CONFIG_DIR", "")
	t.Setenv("CDNFIX_STATE_DIR", "")
	t.Setenv("CDNFIX_LOG_DIR", "")

	execPath := filepath.Join(t.TempDir(), "bin", "cdnfix")
	layout, err := ResolveLayout(execPath, LayoutOptions{})
	if err != nil {
		t.Fatalf("ResolveLayout failed: %v", err)
	}

	if layout.ConfigDir != "/etc/cdnfix" {
		t.Fatalf("unexpected config dir: %q", layout.ConfigDir)
	}
	if layout.StateDir != "/var/lib/cdnfix" {
		t.Fatalf("unexpected state dir: %q", layout.StateDir)
	}
	if layout.LogDir != "/var/log/cdnfix" {
		t.Fatalf("unexpected log dir: %q", layout.LogDir)
	}
	if layout.CacheDir != "/var/lib/cdnfix/cache" {
		t.Fatalf("unexpected cache dir: %q", layout.CacheDir)
	}
	if layout.RunDir != "/var/lib/cdnfix/runs" {
		t.Fatalf("unexpected run dir: %q", layout.RunDir)
	}
	if layout.SiteConfig != "/etc/cdnfix/sites.yaml" {
		t.Fatalf("unexpected site config: %q", layout.SiteConfig)
	}
	if layout.Manifest != "/etc/cdnfix/jobs.yaml" {
		t.Fatalf("unexpected manifest path: %q", layout.Manifest)
	}
}

func TestResolveLayoutMapsPortableRootShortcut(t *testing.T) {
	root := t.TempDir()
	execPath := filepath.Join(root, "bin", "cdnfix")
	configDir := filepath.Join(root, "config")
	siteEnv := filepath.Join(configDir, ".env")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(siteEnv, []byte("SECRET_ID=test\n"), 0644); err != nil {
		t.Fatalf("failed to write site env: %v", err)
	}

	layout, err := ResolveLayout(execPath, LayoutOptions{RootDir: ".."})
	if err != nil {
		t.Fatalf("ResolveLayout failed: %v", err)
	}

	if layout.RootDir != root {
		t.Fatalf("unexpected root dir: got %q want %q", layout.RootDir, root)
	}
	if layout.ConfigDir != filepath.Join(root, "config") {
		t.Fatalf("unexpected config dir: %q", layout.ConfigDir)
	}
	if layout.StateDir != filepath.Join(root, "var", "lib") {
		t.Fatalf("unexpected state dir: %q", layout.StateDir)
	}
	if layout.LogDir != filepath.Join(root, "var", "log") {
		t.Fatalf("unexpected log dir: %q", layout.LogDir)
	}
	if layout.CacheDir != filepath.Join(root, "var", "lib", "cache") {
		t.Fatalf("unexpected cache dir: %q", layout.CacheDir)
	}
	if layout.RunDir != filepath.Join(root, "var", "lib", "runs") {
		t.Fatalf("unexpected run dir: %q", layout.RunDir)
	}
	if layout.SiteConfig != siteEnv {
		t.Fatalf("unexpected site config fallback: %q", layout.SiteConfig)
	}
	if layout.Manifest != filepath.Join(root, "config", "jobs.yaml") {
		t.Fatalf("unexpected manifest: %q", layout.Manifest)
	}
}

func TestResolveLayoutPrefersExplicitOptionsThenEnvThenRoot(t *testing.T) {
	root := t.TempDir()
	execPath := filepath.Join(root, "bin", "cdnfix")
	t.Setenv("CDNFIX_ROOT", "/env-root")
	t.Setenv("CDNFIX_CONFIG_DIR", "env-config")
	t.Setenv("CDNFIX_STATE_DIR", "env-state")
	t.Setenv("CDNFIX_LOG_DIR", "env-log")

	layout, err := ResolveLayout(execPath, LayoutOptions{
		RootDir:   "..",
		ConfigDir: "cli-config",
		LogDir:    "cli-log",
	})
	if err != nil {
		t.Fatalf("ResolveLayout failed: %v", err)
	}

	if layout.RootDir != root {
		t.Fatalf("unexpected root dir: got %q want %q", layout.RootDir, root)
	}
	if layout.ConfigDir != filepath.Join(root, "cli-config") {
		t.Fatalf("explicit config dir did not win: %q", layout.ConfigDir)
	}
	if layout.StateDir != filepath.Join(root, "env-state") {
		t.Fatalf("env state dir did not override root shortcut: %q", layout.StateDir)
	}
	if layout.LogDir != filepath.Join(root, "cli-log") {
		t.Fatalf("explicit log dir did not win: %q", layout.LogDir)
	}
	if layout.CacheDir != filepath.Join(root, "env-state", "cache") {
		t.Fatalf("unexpected cache dir: %q", layout.CacheDir)
	}
	if layout.RunDir != filepath.Join(root, "env-state", "runs") {
		t.Fatalf("unexpected run dir: %q", layout.RunDir)
	}
	if layout.SiteConfig != filepath.Join(root, "cli-config", "sites.yaml") {
		t.Fatalf("unexpected site config: %q", layout.SiteConfig)
	}
	if layout.Manifest != filepath.Join(root, "cli-config", "jobs.yaml") {
		t.Fatalf("unexpected manifest: %q", layout.Manifest)
	}
}

func TestNewRunUsesMonthlyJournalPath(t *testing.T) {
	paths := Paths{
		LogDir:   "/var/log/cdnfix",
		CacheDir: "/var/lib/cdnfix/cache",
		RunDir:   "/var/lib/cdnfix/runs",
	}
	now := time.Date(2026, time.May, 11, 9, 30, 0, 0, time.UTC)

	run := NewRun(paths, "prod-a", "refresh", "prod-a-refresh", "/tmp/urls.txt", now)

	if run.RunJournal != "/var/lib/cdnfix/runs/2026-05.jsonl" {
		t.Fatalf("unexpected run journal path: %q", run.RunJournal)
	}
	if run.LogFile != "/var/log/cdnfix/2026-05-11/prod-a.refresh.prod-a-refresh."+strconv.FormatInt(now.UnixNano(), 10)+".log" {
		t.Fatalf("unexpected log file path: %q", run.LogFile)
	}
}

func TestAppendRunAppendsJSONLines(t *testing.T) {
	root := t.TempDir()
	runJournal := filepath.Join(root, "runs", "2026-05.jsonl")
	run := RunRecord{
		ID:         "run-1",
		Site:       "prod-a",
		Action:     "push",
		LogFile:    filepath.Join(root, "logs", "run-1.log"),
		CacheFile:  filepath.Join(root, "cache", "prod-a", "push.tasks.json"),
		RunJournal: runJournal,
		StartedAt:  "2026-05-11T10:00:00Z",
		Status:     "running",
	}

	if err := AppendRun(run); err != nil {
		t.Fatalf("AppendRun() first call failed: %v", err)
	}
	run.Status = "submitted"
	run.FinishedAt = "2026-05-11T10:00:01Z"
	if err := AppendRun(run); err != nil {
		t.Fatalf("AppendRun() second call failed: %v", err)
	}

	content, err := os.ReadFile(runJournal)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 json lines, got %d: %q", len(lines), string(content))
	}

	var first RunRecord
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("failed to unmarshal first line: %v", err)
	}
	if first.Status != "running" {
		t.Fatalf("unexpected first status: %q", first.Status)
	}

	var second RunRecord
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("failed to unmarshal second line: %v", err)
	}
	if second.Status != "submitted" {
		t.Fatalf("unexpected second status: %q", second.Status)
	}
	if second.FinishedAt != "2026-05-11T10:00:01Z" {
		t.Fatalf("unexpected second finished_at: %q", second.FinishedAt)
	}
}
