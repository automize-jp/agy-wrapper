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
