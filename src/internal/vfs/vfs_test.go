package vfs_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// TestParseTreeAndData проверяет типы узлов, права, текстовые и двоичные данные, а также размеры.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestParseTreeAndData(t *testing.T) {
	input := `<vfs name="test-vfs"><directory name="/" mode="750">
		<directory name="docs"><file name="empty.txt"/></directory>
		<file name="hello.txt" mode="640">SGVs bG8sIFZGUyEK</file>
		<file name="data.bin">AAEC/w==</file>
	</directory></vfs>`
	filesystem, err := vfs.Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if filesystem.Name != "test-vfs" || !filesystem.Root.Mode.IsDir() || filesystem.Root.Mode.Perm() != 0750 {
		t.Fatalf("неверные метаданные корня: %+v", filesystem)
	}
	docs := child(t, filesystem.Root, "docs")
	if !docs.Mode.IsDir() || docs.Mode.Perm() != 0755 || docs.Size() != 0 {
		t.Fatalf("неверные метаданные каталога: %+v", docs)
	}
	hello := child(t, filesystem.Root, "hello.txt")
	if hello.Mode.IsDir() || hello.Mode.Perm() != 0640 || string(hello.Data) != "Hello, VFS!\n" {
		t.Fatalf("неверные данные файла: %+v", hello)
	}
	binary := child(t, filesystem.Root, "data.bin")
	if !bytes.Equal(binary.Data, []byte{0, 1, 2, 255}) || binary.Mode.Perm() != 0644 || binary.Size() != 4 {
		t.Fatalf("неверные двоичные данные: %+v", binary)
	}
	if child(t, docs, "empty.txt").Size() != 0 || hello.Size() != int64(len(hello.Data)) {
		t.Fatal("размер должен соответствовать декодированным байтам")
	}
}

// TestParseInvalidDocuments проверяет синтаксис, структуру, имена, права, base64 и повторные имена.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestParseInvalidDocuments(t *testing.T) {
	tests := []struct{ input, message string }{
		{"", "неверный XML"},
		{`<vfs>`, "неверный XML"},
		{`<other/>`, "неверный XML"},
		{`<vfs><directory name="/"/></vfs>`, "имя VFS"},
		{`<vfs name="x"/>`, "корневой directory"},
		{`<vfs name="x"><directory name="a"/></vfs>`, "корневой directory"},
		{`<vfs name="x"><directory name="/"/><directory name="/"/></vfs>`, "корневой directory"},
		{`<vfs name="x"><unknown/></vfs>`, "неизвестный элемент"},
		{wrap(`unexpected`), "текст допустим"},
		{wrap(`<unknown/>`), "неизвестный элемент"},
		{wrap(`<directory name=".."/>`), "неверное имя"},
		{wrap(`<file name="a/b"/>`), "неверное имя"},
		{wrap(`<file/>`), "неверное имя"},
		{wrap(`<file name="a" mode="888"/>`), "неверные права"},
		{wrap(`<directory name="a" mode="0755"/>`), "три восьмеричные"},
		{wrap(`<file name="a">invalid!</file>`), "неверный base64"},
		{wrap(`<file name="a"><file name="b"/></file>`), "неизвестный элемент"},
		{wrap(`<file name="a"/><directory name="a"/>`), "повторяется имя"},
		{wrap("") + `<vfs/>`, "лишние данные"},
		{"text" + wrap(""), "ожидается корневой"},
	}
	for _, tt := range tests {
		filesystem, err := vfs.Parse([]byte(tt.input))
		if err == nil || filesystem != nil || !strings.Contains(err.Error(), tt.message) {
			t.Fatalf("Parse(%q) = %+v, %v; want %q", tt.input, filesystem, err, tt.message)
		}
	}
}

// TestExamples проверяет загрузку трёх примеров и наличие файлов в глубоко вложенном каталоге.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestExamples(t *testing.T) {
	for _, name := range []string{"minimal", "multiple", "deep"} {
		path := filepath.Join("../../..", "testdata/vfs", name+".xml")
		filesystem, err := vfs.LoadFile(path)
		if err != nil || filesystem.Name != name+"-vfs" {
			t.Fatalf("пример %s: %+v, %v", name, filesystem, err)
		}
		if name == "deep" {
			docs := child(t, filesystem.Root, "docs")
			reports := child(t, docs, "reports")
			year := child(t, reports, "2026")
			if string(child(t, year, "summary.txt").Data) != "summary\n" {
				t.Fatal("неверные данные вложенного файла")
			}
		}
	}
}

// TestLoadFileKeepsSource проверяет неизменность XML и отсутствие распаковки файлов на диск.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadFileKeepsSource(t *testing.T) {
	const fileMode = 0644
	dir := t.TempDir()
	path := filepath.Join(dir, "vfs.xml")
	data := []byte(wrap(`<file name="hello.txt">SGVsbG8=</file>`))
	if err := os.WriteFile(path, data, fileMode); err != nil {
		t.Fatal(err)
	}
	filesystem, err := vfs.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := shell.New(filesystem)
	for _, command := range []string{"ls /", "cd /", "unknown", "exit"} {
		_, _ = engine.Execute(command)
	}
	child(t, filesystem.Root, "hello.txt").Data[0] = 'X'
	filesystem.Root.Mode = filesystem.Root.Mode.Type() | 0700
	stored, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(stored, data) {
		t.Fatalf("исходный XML изменился: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "vfs.xml" {
		t.Fatalf("на диск извлечены дополнительные файлы: %v, %v", entries, err)
	}
}

// TestLoadMissingFile проверяет сохранение причины ошибки и имени операции при отсутствующем файле.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadMissingFile(t *testing.T) {
	_, err := vfs.LoadFile(filepath.Join(t.TempDir(), "missing.xml"))
	if !errors.Is(err, os.ErrNotExist) || !strings.HasPrefix(err.Error(), "vfs.LoadFile:") {
		t.Fatalf("причина ошибки загрузки потеряна: %v", err)
	}
}

// wrap возвращает минимальный XML-документ VFS с заданным содержимым корневого каталога.
func wrap(contents string) string {
	return `<vfs name="test"><directory name="/">` + contents + `</directory></vfs>`
}

// child находит дочерний узел по имени и возвращает его; при отсутствии завершает тест через t.
func child(t *testing.T, parent *vfs.Node, name string) *vfs.Node {
	t.Helper()
	for _, node := range parent.Children {
		if node.Name == name {
			return node
		}
	}
	t.Fatalf("в каталоге %q нет узла %q", parent.Name, name)
	return nil
}
