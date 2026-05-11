package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildInitPlan(t *testing.T) {
	plan, err := buildInitPlan("/etc/cdnfix", "/var/lib/cdnfix", "/var/log/cdnfix", "/etc/cdnfix/sites.yaml", "/etc/cdnfix/jobs.yaml", "prod-a")
	if err != nil {
		t.Fatalf("buildInitPlan failed: %v", err)
	}
	if plan.URLDir != "/etc/cdnfix/urls/prod-a" {
		t.Fatalf("unexpected url dir: %s", plan.URLDir)
	}
	if plan.RefreshFile != "/etc/cdnfix/urls/prod-a/refresh.txt" {
		t.Fatalf("unexpected refresh file: %s", plan.RefreshFile)
	}
}

func TestWriteScaffoldFilePreservesExistingWithoutForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sites.yaml")
	if err := os.WriteFile(path, []byte("original\n"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := writeScaffoldFile(path, "new\n", false); err != nil {
		t.Fatalf("writeScaffoldFile failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "original\n" {
		t.Fatalf("expected original content, got %q", string(data))
	}
}

func TestWriteScaffoldFileOverwritesWithForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sites.yaml")
	if err := os.WriteFile(path, []byte("original\n"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := writeScaffoldFile(path, "new\n", true); err != nil {
		t.Fatalf("writeScaffoldFile failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "new\n" {
		t.Fatalf("expected new content, got %q", string(data))
	}
}

func TestRenderJobsTemplate(t *testing.T) {
	content := renderJobsTemplate("prod-a")
	if !strings.Contains(content, "./urls/prod-a/refresh.txt") {
		t.Fatalf("missing refresh path in template: %s", content)
	}
	if !strings.Contains(content, "./urls/prod-a/push.txt") {
		t.Fatalf("missing push path in template: %s", content)
	}
}
