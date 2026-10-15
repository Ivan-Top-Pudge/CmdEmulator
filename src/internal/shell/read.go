package shell

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// diskUsage возвращает размеры каталогов с потомками в байтах; для файла выводит его размер.
func (sh *Shell) diskUsage(args []string) (Result, error) {
	const op = "shell.du"
	name, err := singlePath(args, ".")
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	stack, err := sh.filesystem.Resolve(sh.current, name)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	var lines []string
	node := stack[len(stack)-1]
	total := directoryUsage(node, name, &lines)
	if !node.Mode.IsDir() {
		lines = append(lines, fmt.Sprintf("%d\t%s", total, name))
	}
	return Result{Output: strings.Join(lines, "\n")}, nil
}

// directoryUsage возвращает сумму байтов файлов; строки каталогов добавляет после их потомков.
func directoryUsage(node *vfs.Node, name string, lines *[]string) int64 {
	if !node.Mode.IsDir() {
		return node.Size()
	}
	children := append([]*vfs.Node(nil), node.Children...)
	sort.Slice(children, func(i, j int) bool { return children[i].Name < children[j].Name })
	var total int64
	for _, child := range children {
		total += directoryUsage(child, path.Join(name, child.Name), lines)
	}
	*lines = append(*lines, fmt.Sprintf("%d\t%s", total, name))
	return total
}

// wordCount возвращает количество LF, слов по пробелам Unicode и байтов одного обязательного файла.
func (sh *Shell) wordCount(args []string) (Result, error) {
	const op = "shell.wc"
	name, err := singlePath(args, "")
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	if name == "" {
		return Result{}, fmt.Errorf("%s: требуется путь к файлу", op)
	}
	stack, err := sh.filesystem.Resolve(sh.current, name)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	node := stack[len(stack)-1]
	if node.Mode.IsDir() {
		return Result{}, fmt.Errorf("%s: %q: это каталог", op, name)
	}
	lines := bytes.Count(node.Data, []byte{'\n'})
	words := len(strings.Fields(string(node.Data)))
	return Result{Output: fmt.Sprintf("%d %d %d %s", lines, words, node.Size(), name)}, nil
}
