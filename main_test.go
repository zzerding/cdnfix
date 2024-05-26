package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestReadURLs(t *testing.T) {
	// Test case 1: URLs provided as command line argument
	urls := "https://example.com/page1,https://example.com/page2"
	urlList, err := readURLs(urls, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := []string{"https://example.com/page1", "https://example.com/page2"}
	if !reflect.DeepEqual(urlList, expected) {
		t.Errorf("Expected %v, got %v", expected, urlList)
	}

	// Test case 2: URLs provided in file
	file, err := ioutil.TempFile("", "urls")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString("https://example.com/page3\nhttps://example.com/page4")
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	urlList, err = readURLs("", file.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected = []string{"https://example.com/page3", "https://example.com/page4"}
	if !reflect.DeepEqual(urlList, expected) {
		t.Errorf("Expected %v, got %v", expected, urlList)
	}

	// Test case 3: Neither URLs nor file provided
	_, err = readURLs("", "")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedErrMsg := "either --urls or --file must be provided"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
