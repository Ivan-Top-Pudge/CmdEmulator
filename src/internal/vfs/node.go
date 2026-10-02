package vfs

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

const (
	defaultDirectoryMode fs.FileMode = 0755
	defaultFileMode      fs.FileMode = 0644
	modeDigits                       = 3
	modeBase                         = 8
	modeBits                         = 9
)

// buildDirectory рекурсивно создаёт каталог и его потомков в памяти.
// Возвращает узел либо ошибку имени, прав, содержимого или повторяющихся имён внутри каталога.
func buildDirectory(source directoryXML, root bool) (*Node, error) {
	const op = "vfs.buildDirectory"
	if !root {
		if err := validateName(source.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if err := validateContent(source.Extra, source.Text); err != nil {
		return nil, fmt.Errorf("%s: каталог %q: %w", op, source.Name, err)
	}
	mode, err := parseMode(source.Mode, defaultDirectoryMode)
	if err != nil {
		return nil, fmt.Errorf("%s: каталог %q: %w", op, source.Name, err)
	}
	node := &Node{Name: source.Name, Mode: mode | fs.ModeDir}
	for _, directory := range source.Directories {
		child, err := buildDirectory(directory, false)
		if err != nil {
			return nil, fmt.Errorf("%s: каталог %q: %w", op, source.Name, err)
		}
		node.Children = append(node.Children, child)
	}
	for _, file := range source.Files {
		child, err := buildFile(file)
		if err != nil {
			return nil, fmt.Errorf("%s: каталог %q: %w", op, source.Name, err)
		}
		node.Children = append(node.Children, child)
	}
	if err := validateChildren(node); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return node, nil
}

// buildFile создаёт файл с декодированными данными base64 и правами доступа.
// Возвращает узел либо ошибку имени, прав, вложенных элементов или base64.
func buildFile(source fileXML) (*Node, error) {
	const op = "vfs.buildFile"
	if err := validateName(source.Name); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := validateContent(source.Extra, ""); err != nil {
		return nil, fmt.Errorf("%s: файл %q: %w", op, source.Name, err)
	}
	mode, err := parseMode(source.Mode, defaultFileMode)
	if err != nil {
		return nil, fmt.Errorf("%s: файл %q: %w", op, source.Name, err)
	}
	encoded := strings.Join(strings.Fields(source.Data), "")
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%s: файл %q: неверный base64: %w", op, source.Name, err)
	}
	return &Node{Name: source.Name, Mode: mode, Data: data}, nil
}

// parseMode разбирает три восьмеричные цифры прав или использует значение по умолчанию.
// Возвращает права доступа либо ошибку формата; тип каталога устанавливается отдельно.
func parseMode(value string, fallback fs.FileMode) (fs.FileMode, error) {
	const op = "vfs.parseMode"
	if value == "" {
		return fallback, nil
	}
	if len(value) != modeDigits {
		return 0, fmt.Errorf("%s: mode должен содержать три восьмеричные цифры", op)
	}
	mode, err := strconv.ParseUint(value, modeBase, modeBits)
	if err != nil {
		return 0, fmt.Errorf("%s: неверные права %q: %w", op, value, err)
	}
	return fs.FileMode(mode), nil
}

// validateName проверяет непустое имя одного файла или каталога без /, нулевого байта, `.` и `..`.
// Возвращает nil при успехе либо ошибку имени.
func validateName(name string) error {
	const op = "vfs.validateName"
	if strings.TrimSpace(name) == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\x00") {
		return fmt.Errorf("%s: неверное имя %q", op, name)
	}
	return nil
}

// validateChildren проверяет уникальность имён файлов и каталогов у одного родителя.
// Возвращает nil при успехе либо ошибку с повторяющимся именем.
func validateChildren(node *Node) error {
	const op = "vfs.validateChildren"
	names := make(map[string]bool)
	for _, child := range node.Children {
		if names[child.Name] {
			return fmt.Errorf("%s: в каталоге %q повторяется имя %q", op, node.Name, child.Name)
		}
		names[child.Name] = true
	}
	return nil
}
