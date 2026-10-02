package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunHelpAndConfigurationError проверяет возврат из run до создания GUI при справке и ошибке.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestRunHelpAndConfigurationError(t *testing.T) {
	if err := run([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("справка должна возвращать flag.ErrHelp: %v", err)
	}
	if err := run(nil); err == nil || !strings.HasPrefix(err.Error(), "main.run: config.Load:") {
		t.Fatalf("отсутствие VFS должно дать ошибку до запуска GUI: %v", err)
	}
}

// TestRunScriptReadError проверяет отказ от запуска GUI, если скрипт невозможно прочитать полностью.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestRunScriptReadError(t *testing.T) {
	const fileMode = 0644
	const oversizedLineLength = 1024 * 1024
	dir := t.TempDir()
	vfs := filepath.Join(dir, "vfs.xml")
	script := filepath.Join(dir, "script.txt")
	if err := os.WriteFile(vfs, []byte(`<vfs name="test"><directory name="/"/></vfs>`), fileMode); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte(strings.Repeat("x", oversizedLineLength)), fileMode); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"--vfs", vfs, "--script", script})
	if err == nil || !strings.Contains(err.Error(), "main.run: startup.LoadFile:") {
		t.Fatalf("ошибка чтения скрипта должна прервать запуск: %v", err)
	}
}

// TestRunInvalidVFS проверяет отказ от создания GUI при неверном XML.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestRunInvalidVFS(t *testing.T) {
	const fileMode = 0644
	path := filepath.Join(t.TempDir(), "broken.xml")
	if err := os.WriteFile(path, []byte("<vfs>"), fileMode); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"--vfs", path})
	if err == nil || !strings.HasPrefix(err.Error(), "main.run: vfs.LoadFile:") {
		t.Fatalf("неверная VFS должна прервать запуск до GUI: %v", err)
	}
}
