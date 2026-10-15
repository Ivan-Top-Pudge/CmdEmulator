package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunUnclosedScriptComment проверяет отказ от запуска GUI до выполнения даже первой команды скрипта.
func TestRunUnclosedScriptComment(t *testing.T) {
	folder := t.TempDir()
	const fileMode = 0644
	files := map[string]string{
		"vfs.xml":    `<vfs name="test"><directory name="/"/></vfs>`,
		"script.txt": "exit\n/* незакрытый комментарий",
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(data), fileMode); err != nil {
			t.Fatal(err)
		}
	}
	err := run([]string{"--vfs", filepath.Join(folder, "vfs.xml"), "--script", filepath.Join(folder, "script.txt")})
	if err == nil || !strings.Contains(err.Error(), "startup.LoadFile:") || !strings.Contains(err.Error(), "строки 2") {
		t.Fatalf("незакрытый блок должен прервать запуск до GUI: %v", err)
	}
}
