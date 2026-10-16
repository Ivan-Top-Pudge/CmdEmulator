package vfs_test

import (
	"bytes"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// assertParsedDirectories проверяет имя VFS, тип и права каталогов, а также нулевые размеры.
func assertParsedDirectories(t *testing.T, filesystem *vfs.FS) {
	t.Helper()
	const (
		wantRootPermissions = 0750
		wantDocsPermissions = 0755
	)
	if filesystem.Name != "test-vfs" || !filesystem.Root.Mode.IsDir() || filesystem.Root.Mode.Perm() != wantRootPermissions {
		t.Fatalf("неверные метаданные корня: %+v", filesystem)
	}
	docs := child(t, filesystem.Root, "docs")
	if !docs.Mode.IsDir() || docs.Mode.Perm() != wantDocsPermissions || docs.Size() != 0 {
		t.Fatalf("неверные метаданные каталога: %+v", docs)
	}
	if child(t, docs, "empty.txt").Size() != 0 {
		t.Fatal("пустой файл должен иметь нулевой размер")
	}
}

// assertParsedFiles проверяет тип, права, содержимое и декодированные размеры текстового и двоичного файлов.
func assertParsedFiles(t *testing.T, root *vfs.Node) {
	t.Helper()
	const (
		wantTextPermissions   = 0640
		wantBinaryPermissions = 0644
	)
	hello := child(t, root, "hello.txt")
	if hello.Mode.IsDir() || hello.Mode.Perm() != wantTextPermissions || string(hello.Data) != "Hello, VFS!\n" {
		t.Fatalf("неверные данные файла: %+v", hello)
	}
	binary := child(t, root, "data.bin")
	wantBinaryData := []byte{0, 1, 2, 255}
	if !bytes.Equal(binary.Data, wantBinaryData) || binary.Mode.Perm() != wantBinaryPermissions || binary.Size() != int64(len(wantBinaryData)) {
		t.Fatalf("неверные двоичные данные: %+v", binary)
	}
	if hello.Size() != int64(len(hello.Data)) {
		t.Fatal("размер должен соответствовать декодированным байтам")
	}
}
