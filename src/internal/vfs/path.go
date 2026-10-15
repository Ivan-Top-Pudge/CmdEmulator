package vfs

import (
	"fmt"
	"strings"
)

// Resolve последовательно обходит путь от current или корня для абсолютного пути.
// Возвращает новый стек узлов, не меняя current; . пропускается, .. поднимается до корня.
func (filesystem *FS) Resolve(current []*Node, path string) ([]*Node, error) {
	const op = "vfs.Resolve"
	if path == "" || strings.ContainsRune(path, '\x00') {
		return nil, fmt.Errorf("%s: нужен непустой путь без нулевого байта", op)
	}
	stack := append([]*Node(nil), current...)
	if strings.HasPrefix(path, "/") || len(stack) == 0 {
		stack = []*Node{filesystem.Root}
	}
	for _, part := range strings.Split(path, "/") {
		if err := walkPart(&stack, part); err != nil {
			return nil, fmt.Errorf("%s: путь %q: %w", op, path, err)
		}
	}
	return stack, nil
}

// walkPart добавляет потомка, пропускает . или поднимает стек для ..; возвращает ошибку перехода.
func walkPart(stack *[]*Node, part string) error {
	const op = "vfs.walkPart"
	node := (*stack)[len(*stack)-1]
	if !node.Mode.IsDir() {
		return fmt.Errorf("%s: %q: не каталог", op, node.Name)
	}
	switch part {
	case "", ".":
		return nil
	case "..":
		if len(*stack) > 1 {
			*stack = (*stack)[:len(*stack)-1]
		}
		return nil
	}
	for _, child := range node.Children {
		if child.Name == part {
			*stack = append(*stack, child)
			return nil
		}
	}
	return fmt.Errorf("%s: %q: путь не существует", op, part)
}
