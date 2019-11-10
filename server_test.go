package torbula

import (
	"testing"
)

func TestTopmostPath(t *testing.T) {
	testTable := []struct {
		path     string
		expected string
	}{
		{"a/b/c.txt", "a"},
		{"/a/b/c.txt", "/"},
		{"a.txt", "a.txt"},
		{"../a/b/c.txt", ".."},
		{"a/b", "a"},
		{"a/b/", "a"},
		{"a/b/.", "a"},
		{"한글/경로/test.txt", "한글"},
	}
	for _, table := range testTable {
		result := topmostPath(table.path)
		if result != table.expected {
			t.Errorf("got rootDir(%q) = %q, want %q.",
				table.path, result, table.expected)
		}
	}
}
