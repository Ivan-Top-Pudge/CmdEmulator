package vfs

import (
	"fmt"
	"io/fs"
	"strings"
)

// MakeDirectory создаёт один каталог в существующем родителе только в памяти.
// Возвращает ошибку пути, имени или существующего узла; при ошибке дерево не меняется.
func (filesystem *FS) MakeDirectory(current []*Node, path string) error {
	const op = "vfs.MakeDirectory"
	parentPath, name := splitParent(path)
	if err := validateName(name); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stack, err := filesystem.Resolve(current, parentPath)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	parent := stack[len(stack)-1]
	if !parent.Mode.IsDir() {
		return fmt.Errorf("%s: %q: не каталог", op, parentPath)
	}
	for _, child := range parent.Children {
		if child.Name == name {
			return fmt.Errorf("%s: %q: уже существует", op, path)
		}
	}
	parent.Children = append(parent.Children, &Node{Name: name, Mode: fs.ModeDir | defaultDirectoryMode})
	return nil
}

// splitParent отделяет последнее имя от родителя без упрощения . и ..; завершающие / пропускает.
func splitParent(path string) (string, string) {
	trimmed := strings.TrimRight(path, "/")
	index := strings.LastIndex(trimmed, "/")
	if index < 0 {
		return ".", trimmed
	}
	if index == 0 {
		return "/", trimmed[1:]
	}
	return trimmed[:index], trimmed[index+1:]
}

// ChangeMode устанавливает три восьмеричные цифры прав файла или каталога только в памяти.
// Сохраняет тип узла; при неверном режиме или пути возвращает ошибку без изменения дерева.
func (filesystem *FS) ChangeMode(current []*Node, path, permissions string) error {
	const op = "vfs.ChangeMode"
	if len(permissions) != modeDigits {
		return fmt.Errorf("%s: права должны содержать три восьмеричные цифры", op)
	}
	mode, err := parseMode(permissions, 0)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stack, err := filesystem.Resolve(current, path)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	node := stack[len(stack)-1]
	node.Mode = node.Mode&^fs.ModePerm | mode
	return nil
}
