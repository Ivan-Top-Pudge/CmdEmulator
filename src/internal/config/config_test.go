package config_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/config"
)

// TestLoadDefaults проверяет обязательный путь, значения по умолчанию и отладочный вывод.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadDefaults(t *testing.T) {
	vfs := writeFile(t, t.TempDir(), "sample.xml", "<vfs/>")
	got, err := config.Load([]string{"--vfs", vfs})
	if err != nil {
		t.Fatal(err)
	}
	want := config.Settings{VFSPath: vfs, Prompt: config.DefaultPrompt}
	if got != want || got.VFSName() != "sample.xml" {
		t.Fatalf("Load = %+v; want %+v", got, want)
	}
	if !strings.Contains(got.String(), fmt.Sprintf("vfs=%q", vfs)) {
		t.Fatalf("в отладочном выводе нет пути: %s", got)
	}
}

// TestLoadJSONPaths проверяет пути относительно конфигурации и чтение всех её полей.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadJSONPaths(t *testing.T) {
	dir := t.TempDir()
	vfs := writeFile(t, dir, "data/vfs.xml", "<vfs/>")
	script := writeFile(t, dir, "commands.txt", "ls\n")
	path := writeFile(t, dir, "cfg/default.json",
		`{"vfs":"../data/vfs.xml","prompt":"student> ","script":"../commands.txt"}`)
	got, err := config.Load([]string{"--config", path})
	want := config.Settings{VFSPath: vfs, Prompt: "student> ", ScriptPath: script, ConfigPath: path}
	if err != nil || got != want {
		t.Fatalf("Load = %+v, %v; want %+v", got, err, want)
	}
}

// TestLoadCLIOverridesJSON проверяет приоритет CLI, включая пустые значения prompt и script.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadCLIOverridesJSON(t *testing.T) {
	dir := t.TempDir()
	vfs := writeFile(t, dir, "override.xml", "<vfs/>")
	path := writeFile(t, dir, "default.json",
		`{"vfs":"missing.xml","prompt":"json> ","script":"missing.txt"}`)
	got, err := config.Load([]string{"--config", path, "--vfs=" + vfs, "--prompt=", "--script", ""})
	want := config.Settings{VFSPath: vfs, Prompt: "", ScriptPath: "", ConfigPath: path}
	if err != nil || got != want {
		t.Fatalf("Load = %+v, %v; want %+v", got, err, want)
	}
}

// TestLoadCLIRelativePaths проверяет, что CLI-пути не привязываются к каталогу JSON.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadCLIRelativePaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "cli.xml", "<vfs/>")
	writeFile(t, dir, "commands.txt", "ls\n")
	path := writeFile(t, dir, "nested/default.json", `{"vfs":"missing.xml"}`)
	t.Chdir(dir)
	got, err := config.Load([]string{"--config", path, "--vfs", "cli.xml", "--script", "commands.txt"})
	if err != nil || got.VFSPath != "cli.xml" || got.ScriptPath != "commands.txt" {
		t.Fatalf("CLI-пути должны быть относительно cwd: %+v, %v", got, err)
	}
}

// TestLoadOptionalJSONFields проверяет значения по умолчанию и явно пустое приглашение в JSON.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadOptionalJSONFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "vfs.xml", "<vfs/>")
	for _, tt := range []struct{ body, prompt string }{
		{`{"vfs":"vfs.xml"}`, config.DefaultPrompt},
		{`{"vfs":"vfs.xml","prompt":"","script":""}`, ""},
	} {
		path := writeFile(t, dir, "default.json", tt.body)
		got, err := config.Load([]string{"--config=" + path})
		if err != nil || got.Prompt != tt.prompt || got.ScriptPath != "" {
			t.Fatalf("неверные значения по умолчанию: %+v, %v", got, err)
		}
	}
}

// TestLoadCLIErrorCases проверяет неверные, отсутствующие и повторные параметры, а также пути.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadCLIErrorCases(t *testing.T) {
	dir := t.TempDir()
	vfs := writeFile(t, dir, "vfs.xml", "<vfs/>")
	tests := []struct {
		args []string
		text string
	}{
		{nil, "укажите путь к VFS"},
		{[]string{"--unknown"}, "неизвестный параметр"},
		{[]string{"file.xml"}, "неизвестный параметр"},
		{[]string{"--vfs"}, "требуется значение"},
		{[]string{"--prompt", "--vfs", vfs}, "требуется значение"},
		{[]string{"--vfs", vfs, "--vfs", vfs}, "указан повторно"},
		{[]string{"--vfs="}, "укажите путь к VFS"},
		{[]string{"--config="}, "непустой путь"},
		{[]string{"--config", filepath.Join(dir, "missing.json")}, "открыть конфигурацию"},
		{[]string{"--vfs", filepath.Join(dir, "missing.xml")}, "открыть VFS"},
		{[]string{"--vfs", dir}, "обычным файлом"},
		{[]string{"--vfs", vfs, "--script", filepath.Join(dir, "missing.txt")}, "открыть стартовый"},
		{[]string{"--vfs", vfs, "--script", dir}, "обычным файлом"},
	}
	for _, tt := range tests {
		_, err := config.Load(tt.args)
		if err == nil || !strings.Contains(err.Error(), tt.text) || !strings.HasPrefix(err.Error(), "config.Load:") {
			t.Fatalf("Load(%q) = %v; want %q with operation", tt.args, err, tt.text)
		}
	}
}

// TestLoadInvalidJSON проверяет формат, типы, неизвестные поля и лишние данные в JSON.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	vfs := writeFile(t, dir, "vfs.xml", "<vfs/>")
	for _, body := range []string{
		`{`, `null`, `[]`, `"text"`, `{"unknown":"value"}`, `{"vfs":12}`,
		`{"prompt":null}`, `{"script":false}`, `{} {}`, `{} trailing`,
	} {
		t.Run(body, func(t *testing.T) {
			path := writeFile(t, dir, "invalid.json", body)
			_, err := config.Load([]string{"--config", path, "--vfs", vfs})
			if err == nil || !strings.Contains(err.Error(), "config.readJSON") {
				t.Fatalf("неверный JSON должен дать ошибку даже с CLI: %v", err)
			}
		})
	}
}

// TestLoadHelp проверяет успешное распознавание справки без обязательной VFS.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadHelp(t *testing.T) {
	for _, option := range []string{"--help", "-h"} {
		_, err := config.Load([]string{option})
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("Load(%s) = %v; want flag.ErrHelp", option, err)
		}
	}
}

// writeFile создаёт тестовый файл с родительскими каталогами и возвращает его путь.
// При ошибке завершает тест через t.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	const directoryMode = 0755
	const fileMode = 0644
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), directoryMode); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), fileMode); err != nil {
		t.Fatal(err)
	}
	return path
}
