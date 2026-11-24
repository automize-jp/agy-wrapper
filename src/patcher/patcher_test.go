package patcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateBackup_SizeMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	realPath := filepath.Join(tmpDir, "agy")
	origPath := filepath.Join(tmpDir, "agy.original")

	realData := []byte("1234567890")
	origData := []byte("12345")
	stored := []byte("stored")

	_ = os.WriteFile(realPath, realData, 0755)
	_ = os.WriteFile(origPath, origData, 0755)

	ok, reason := ValidateBackup(realPath, origPath, realData, origData, stored)
	if ok {
		t.Fatalf("expected ValidateBackup to fail due to size mismatch, but it succeeded")
	}
	if !strings.Contains(reason, "size mismatch") {
		t.Fatalf("expected reason to contain 'size mismatch', got: %s", reason)
	}
}

func TestValidateBackup_PristinePromptMissing(t *testing.T) {
	tmpDir := t.TempDir()
	realPath := filepath.Join(tmpDir, "agy")
	origPath := filepath.Join(tmpDir, "agy.original")

	storedPrompt := []byte("<GRILL_ME>\nOriginal Prompt \n</GRILL_ME>")
	realData := []byte("<GRILL_ME>\nPatched Prompt  \n</GRILL_ME>")
	origData := []byte("<GRILL_ME>\nCorrupted Prompt\n</GRILL_ME>")

	_ = os.WriteFile(realPath, realData, 0755)
	_ = os.WriteFile(origPath, origData, 0755)

	ok, reason := ValidateBackup(realPath, origPath, realData, origData, storedPrompt)
	if ok {
		t.Fatalf("expected ValidateBackup to fail due to corrupted prompt in backup, but it succeeded")
	}
	if !strings.Contains(reason, "does not contain the pristine unpatched prompt") {
		t.Fatalf("expected reason to mention pristine unpatched prompt, got: %s", reason)
	}
}

func TestValidateBackup_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	realPath := filepath.Join(tmpDir, "agy")
	origPath := filepath.Join(tmpDir, "agy.original")

	storedPrompt := []byte("<GRILL_ME>\nOriginal Prompt\n</GRILL_ME>")
	realData := []byte("<GRILL_ME>\nPatched Prompt \n</GRILL_ME>")
	origData := []byte("<GRILL_ME>\nOriginal Prompt\n</GRILL_ME>")

	_ = os.WriteFile(realPath, realData, 0755)
	_ = os.WriteFile(origPath, origData, 0755)

	ok, reason := ValidateBackup(realPath, origPath, realData, origData, storedPrompt)
	if !ok {
		t.Fatalf("expected ValidateBackup to succeed, but failed with: %s", reason)
	}
}

func TestStringReplacements_EqualLengths(t *testing.T) {
	for _, rep := range StringReplacements {
		if len(rep.Old) != len(rep.New) {
			t.Fatalf("replacement %q has unequal lengths: old=%d, new=%d", rep.Name, len(rep.Old), len(rep.New))
		}
	}
}

func TestReplaceStringInPlace_ReplacesAllOccurrences(t *testing.T) {
	oldURL := "https://play.googleapis.com/log"
	newURL := "http://0.0.0.0/logging-disabled"
	orig := []byte("prefix\x00" + oldURL + "\x00middle\x00" + oldURL + "\x00suffix")

	dryBuf := append([]byte(nil), orig...)
	count, msg, err := ReplaceStringInPlace(dryBuf, oldURL, newURL, true)
	if err != nil {
		t.Fatalf("unexpected error in dry-run: %v", err)
	}
	if count != 2 || !strings.Contains(msg, "Replaced 2 occurrence(s)") {
		t.Fatalf("expected 2 occurrences in dry-run, got count=%d msg=%q", count, msg)
	}
	if string(dryBuf) != string(orig) {
		t.Fatalf("dry-run mutated buffer")
	}

	buf := append([]byte(nil), orig...)
	count, msg, err = ReplaceStringInPlace(buf, oldURL, newURL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 || !strings.Contains(msg, "Replaced 2 occurrence(s)") {
		t.Fatalf("expected 2 occurrences, got count=%d msg=%q", count, msg)
	}
	expected := "prefix\x00" + newURL + "\x00middle\x00" + newURL + "\x00suffix"
	if string(buf) != expected {
		t.Fatalf("unexpected buffer content:\ngot:  %q\nwant: %q", string(buf), expected)
	}
}

func TestReplaceStringInPlace_AlreadyPatchedAndMissing(t *testing.T) {
	oldURL := "https://play.googleapis.com/log"
	newURL := "http://0.0.0.0/logging-disabled"

	alreadyBuf := []byte("prefix\x00" + newURL + "\x00suffix")
	count, msg, err := ReplaceStringInPlace(alreadyBuf, oldURL, newURL, false)
	if err != nil || count != 0 || !strings.Contains(msg, "Already patched") {
		t.Fatalf("expected already patched, got count=%d msg=%q err=%v", count, msg, err)
	}

	missingBuf := []byte("prefix\x00something-else\x00suffix")
	count, msg, err = ReplaceStringInPlace(missingBuf, oldURL, newURL, false)
	if err != nil || count != 0 || !strings.Contains(msg, "not found (skipped)") {
		t.Fatalf("expected skipped, got count=%d msg=%q err=%v", count, msg, err)
	}
}

func TestReplaceStringInPlace_LengthMismatch(t *testing.T) {
	buf := []byte("abcdef")
	_, _, err := ReplaceStringInPlace(buf, "abc", "abcd", false)
	if err == nil {
		t.Fatalf("expected error on length mismatch, got nil")
	}
}
