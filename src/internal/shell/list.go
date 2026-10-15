package shell

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

type listOptions struct {
	long, human, all bool
	path             string
}

// list разбирает флаги -l, -h, -a и возвращает отсортированные имена или подробные строки узлов VFS.
func (sh *Shell) list(args []string) (Result, error) {
	const op = "shell.ls"
	options, err := parseListOptions(args)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	stack, err := sh.filesystem.Resolve(sh.current, options.path)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	node := stack[len(stack)-1]
	if !node.Mode.IsDir() {
		return Result{Output: listEntry(node, node.Name, options)}, nil
	}
	return Result{Output: listDirectory(stack, options)}, nil
}

// parseListOptions возвращает флаги и не более одного пути; -- завершает разбор опций.
func parseListOptions(args []string) (listOptions, error) {
	const op = "shell.parseListOptions"
	options := listOptions{path: "."}
	seenPath, flags := false, true
	for _, arg := range args {
		if flags && arg == "--" {
			flags = false
			continue
		}
		if flags && strings.HasPrefix(arg, "-") && arg != "-" {
			if err := options.addFlags(arg[1:]); err != nil {
				return options, err
			}
			continue
		}
		if seenPath {
			return options, fmt.Errorf("%s: ожидается не более одного пути", op)
		}
		if arg == "" {
			return options, fmt.Errorf("%s: нужен непустой путь", op)
		}
		options.path, seenPath = arg, true
	}
	return options, nil
}

// addFlags включает короткие флаги и возвращает ошибку для любого неизвестного символа.
func (options *listOptions) addFlags(flags string) error {
	const op = "shell.listOptions.addFlags"
	for _, flag := range flags {
		switch flag {
		case 'l':
			options.long = true
		case 'h':
			options.human = true
		case 'a':
			options.all = true
		default:
			return fmt.Errorf("%s: неизвестный флаг -%c", op, flag)
		}
	}
	return nil
}

// listDirectory возвращает строки каталога; с -a добавляет скрытые имена, . и .. с их метаданными.
func listDirectory(stack []*vfs.Node, options listOptions) string {
	node := stack[len(stack)-1]
	children := append([]*vfs.Node(nil), node.Children...)
	sort.Slice(children, func(i, j int) bool { return children[i].Name < children[j].Name })
	var lines []string
	if options.all {
		parent := stack[max(0, len(stack)-2)]
		lines = append(lines, listEntry(node, ".", options), listEntry(parent, "..", options))
	}
	for _, child := range children {
		if options.all || !strings.HasPrefix(child.Name, ".") {
			lines = append(lines, listEntry(child, child.Name, options))
		}
	}
	return strings.Join(lines, "\n")
}

// listEntry возвращает имя либо тип, права, размер и имя; размер каталогов равен нулю.
func listEntry(node *vfs.Node, name string, options listOptions) string {
	if !options.long {
		return name
	}
	size := strconv.FormatInt(node.Size(), 10)
	if options.human {
		size = humanSize(node.Size())
	}
	return fmt.Sprintf("%s %s %s", node.Mode.String(), size, name)
}

// humanSize возвращает размер в байтах либо единицах KiB–EiB с одним десятичным знаком без лишнего .0.
func humanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return strconv.FormatInt(size, 10) + "B"
	}
	value, index := float64(size), 0
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	for value >= unit && index < len(units)-1 {
		value /= unit
		index++
	}
	return strings.TrimSuffix(fmt.Sprintf("%.1f", value), ".0") + units[index]
}
