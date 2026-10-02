// Package vfs загружает виртуальную файловую систему из XML только в память.
package vfs

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// FS содержит имя виртуальной файловой системы и её корневой каталог.
type FS struct {
	Name string
	Root *Node
}

// Node представляет каталог или файл в памяти: имя, тип и права, байты и дочерние узлы.
// Для каталога Mode содержит fs.ModeDir, Data пуста; для файла Children пуст.
type Node struct {
	Name     string
	Mode     fs.FileMode
	Data     []byte
	Children []*Node
}

// Size возвращает размер содержимого файла в байтах; для каталога возвращает ноль.
func (n *Node) Size() int64 {
	if n.Mode.IsDir() {
		return 0
	}
	return int64(len(n.Data))
}

// LoadFile читает физический XML-файл и возвращает VFS в памяти либо ошибку.
// Исходный файл открывается только для чтения; файлы VFS на диск не извлекаются.
func LoadFile(path string) (*FS, error) {
	const op = "vfs.LoadFile"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: прочитать VFS %q: %w", op, path, err)
	}
	filesystem, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: файл %q: %w", op, path, err)
	}
	return filesystem, nil
}

// Parse разбирает XML и возвращает дерево с декодированными данными файлов либо ошибку формата.
// Требует имя VFS и ровно один корневой каталог с именем /.
func Parse(data []byte) (*FS, error) {
	const op = "vfs.Parse"
	document, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if strings.TrimSpace(document.Name) == "" {
		return nil, fmt.Errorf("%s: имя VFS не должно быть пустым", op)
	}
	if err := validateContent(document.Extra, document.Text); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	const rootCount = 1
	if len(document.Roots) != rootCount || document.Roots[0].Name != "/" {
		return nil, fmt.Errorf("%s: требуется один корневой directory с именем /", op)
	}
	root, err := buildDirectory(document.Roots[0], true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &FS{Name: document.Name, Root: root}, nil
}
