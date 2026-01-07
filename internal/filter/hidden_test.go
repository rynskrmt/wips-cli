package filter

import (
	"testing"
)

func TestIsHiddenDir(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		hiddenDirs []string
		want       bool
	}{
		{
			name:       "exact match",
			path:       "/home/user/secret",
			hiddenDirs: []string{"/home/user/secret"},
			want:       true,
		},
		{
			name:       "subdirectory match",
			path:       "/home/user/secret/nested",
			hiddenDirs: []string{"/home/user/secret"},
			want:       true,
		},
		{
			name:       "not hidden when no match",
			path:       "/home/user/public",
			hiddenDirs: []string{"/home/user/secret"},
			want:       false,
		},
		{
			name:       "prefix but not subdirectory",
			path:       "/home/user/secret-other",
			hiddenDirs: []string{"/home/user/secret"},
			want:       false,
		},
		{
			name:       "empty hidden dirs",
			path:       "/any/path",
			hiddenDirs: []string{},
			want:       false,
		},
		{
			name:       "nil hidden dirs",
			path:       "/any/path",
			hiddenDirs: nil,
			want:       false,
		},
		{
			name:       "multiple hidden dirs - first matches",
			path:       "/home/user/secret/file",
			hiddenDirs: []string{"/home/user/secret", "/home/user/private"},
			want:       true,
		},
		{
			name:       "multiple hidden dirs - second matches",
			path:       "/home/user/private/file",
			hiddenDirs: []string{"/home/user/secret", "/home/user/private"},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHiddenDir(tt.path, tt.hiddenDirs)
			if got != tt.want {
				t.Errorf("IsHiddenDir(%q, %v) = %v, want %v", tt.path, tt.hiddenDirs, got, tt.want)
			}
		})
	}
}
