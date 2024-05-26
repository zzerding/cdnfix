package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadURLs(t *testing.T) {
	// Test case 1: URLs provided as command line argument
	urls := "https://example.com/page1,https://example.com/page2"
	filePath := ""
	expected := []string{"https://example.com/page1", "https://example.com/page2"}
	result, err := readURLs(urls, filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test case 2: URLs provided in file
	urls = ""
	filePath = "test_urls.txt"
	expected = []string{"https://example.com/page3", "https://example.com/page4"}
	result, err = readURLs(urls, filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test case 3: Neither URLs nor file provided
	urls = ""
	filePath = ""
	result, err = readURLs(urls, filePath)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedErrMsg := "either --urls or --file must be provided"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
