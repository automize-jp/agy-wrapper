package meta

import (
	"runtime"
	"testing"
)

func TestRuntimePlatform(t *testing.T) {
	platform := RuntimePlatform()
	if runtime.GOOS == "darwin" && platform != "macos" {
		t.Fatalf("expected macos for darwin, got %s", platform)
	}
	if runtime.GOOS == "linux" && platform != "linux" {
		t.Fatalf("expected linux for linux, got %s", platform)
	}
	if runtime.GOOS == "windows" && platform != "windows" {
		t.Fatalf("expected windows for windows, got %s", platform)
	}
}

func TestPromptsMatch(t *testing.T) {
	tests := []struct {
		name     string
		stored   map[string]string
		current  map[string]string
		expected bool
	}{
		{
			name:     "Both empty",
			stored:   map[string]string{},
			current:  map[string]string{},
			expected: true,
		},
		{
			name:     "Stored empty but current has values",
			stored:   map[string]string{},
			current:  map[string]string{"grill-me": "hash1"},
			expected: false,
		},
		{
			name:     "Hash mismatch",
			stored:   map[string]string{"grill-me": "hash1"},
			current:  map[string]string{"grill-me": "hash2"},
			expected: false,
		},
		{
			name:     "Exact match",
			stored:   map[string]string{"grill-me": "hash1"},
			current:  map[string]string{"grill-me": "hash1"},
			expected: true,
		},
		{
			name:     "Stored has extra keys but current matches all required",
			stored:   map[string]string{"grill-me": "hash1", "plan": "hash2"},
			current:  map[string]string{"grill-me": "hash1"},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PromptsMatch(tc.stored, tc.current)
			if got != tc.expected {
				t.Fatalf("PromptsMatch() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestGetTargetPromptsSHA256(t *testing.T) {
	prompts := GetTargetPromptsSHA256([]string{"grill-me"})
	hash, ok := prompts["grill-me"]
	if !ok || len(hash) != 64 {
		t.Fatalf("expected 64-char hex sha256 for grill-me, got: %s", hash)
	}
}

func TestGetTargetStringPatchesSHA256(t *testing.T) {
	patches := GetTargetStringPatchesSHA256()
	hash, ok := patches["disable-play-logging"]
	if !ok || len(hash) != 64 {
		t.Fatalf("expected 64-char hex sha256 for disable-play-logging, got: %s", hash)
	}
}
