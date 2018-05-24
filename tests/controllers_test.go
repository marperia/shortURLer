package controllers

import (
	"testing"
	"../controllers"
	"net/url"
	"reflect"
)

func TestFindParams(t *testing.T) {

	tables := []struct {
		firstParam url.Values
		secondtParam []string
		result map[string]string
	}{
		map[string][]string{"url": {"some param"}, "method": {"sha2"},},
		[]string{"url"},
		map[string]string{"url": "some param", "method": "sha2",},
		}

	for _, table := range tables {
		total := controllers.FindParams(table.firstParam, table.secondtParam)
		if !reflect.DeepEqual(table.result, total) {
			t.Errorf("Test of %d amd %d was incorrect, got: %d, want: %d.", table.firstParam, table.secondtParam, total, table.result)
		}
	}
}
