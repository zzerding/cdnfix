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
	manifest := filepath.Join(t.TempDir(), "jobs.yaml")
	content := []byte("jobs:\n  - name: a-refresh\n    site: prod-a\n    action: refresh\n    file: ./urls/prod-a/refresh.txt\n")
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
}
