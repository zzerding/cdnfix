package cmd

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"github.com/zzerding/cdnfix/workflow"
)

func TestReadURLs(t *testing.T) {
	urls := "https://example.com/page1, https://example.com/page2"
	filePath := ""
	expected := []string{"https://example.com/page1", "https://example.com/page2"}
	result, err := workflow.ReadURLs(urls, filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	urls = ""
	tmpFile, err := os.CreateTemp(t.TempDir(), "urls-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString("https://example.com/page3\nhttps://example.com/page4\n"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	filePath = tmpFile.Name()
	expected = []string{"https://example.com/page3", "https://example.com/page4"}
	result, err = workflow.ReadURLs(urls, filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	urls = ""
	filePath = ""
	result, err = workflow.ReadURLs(urls, filePath)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedErrMsg := "either --urls, --file, or stdin must be provided"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
