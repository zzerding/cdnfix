package tencent

import (
	"reflect"
	"testing"
)

func TestSortedSiteNames(t *testing.T) {
	names := SortedSiteNames(map[string]Config{
		"prod-b": {Name: "prod-b"},
		"prod-a": {Name: "prod-a"},
	})
	if !reflect.DeepEqual(names, []string{"prod-a", "prod-b"}) {
		t.Fatalf("unexpected names: %v", names)
	}
}
