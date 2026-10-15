// Package shell обрабатывает команды эмулятора независимо от GUI.
package shell

import (
	"fmt"
	"io/fs"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// Result содержит вывод команды и признак завершения сеанса.
type Result struct {
	Output string
	Exit   bool
}

// Shell хранит VFS и текущий каталог сеанса; команды не обращаются к реальной файловой системе.
type Shell struct {
	filesystem *vfs.FS
	current    []*vfs.Node
}

// New сохраняет загруженную VFS и возвращает обработчик команд эмулятора.
// При nil создаёт пустую VFS; начальный текущий каталог — корень.
func New(filesystem *vfs.FS) *Shell {
	if filesystem == nil {
		filesystem = &vfs.FS{Root: &vfs.Node{Name: "/", Mode: fs.ModeDir | 0755}}
	}
	return &Shell{filesystem: filesystem, current: []*vfs.Node{filesystem.Root}}
}

// Execute разбирает ввод и выполняет команду, возвращая её вывод, признак выхода и ошибку.
// Для пустого ввода возвращает пустой Result без ошибки; ошибки содержат имя операции.
func (sh *Shell) Execute(input string) (Result, error) {
	const op = "shell.Execute"

	args, err := parse(input)
	if err != nil {
		return Result{}, fmt.Errorf("%s: разбор команды: %w", op, err)
	}
	if len(args) == 0 {
		return Result{}, nil
	}
	result, err := sh.dispatch(args[0], args[1:])
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

// dispatch выбирает обработчик команды и возвращает результат либо ошибку аргументов или VFS.
func (sh *Shell) dispatch(command string, args []string) (Result, error) {
	const op = "shell.dispatch"
	switch command {
	case "ls":
		return sh.list(args)
	case "cd":
		return sh.changeDirectory(args)
	case "du":
		return sh.diskUsage(args)
	case "wc":
		return sh.wordCount(args)
	case "mkdir":
		return sh.makeDirectory(args)
	case "chmod":
		return sh.changeMode(args)
	case "exit":
		if len(args) != 0 {
			return Result{}, fmt.Errorf("%s: exit: команда не принимает аргументы", op)
		}
		return Result{Exit: true}, nil
	default:
		return Result{}, fmt.Errorf("%s: %q: неизвестная команда", op, command)
	}
}

// singlePath возвращает единственный путь либо fallback при отсутствии аргументов; опции запрещены.
func singlePath(args []string, fallback string) (string, error) {
	const op = "shell.singlePath"
	const maxPaths = 1
	if len(args) > maxPaths {
		return "", fmt.Errorf("%s: ожидается не более одного пути", op)
	}
	if len(args) == 0 {
		return fallback, nil
	}
	if args[0] == "" {
		return "", fmt.Errorf("%s: нужен непустой путь", op)
	}
	if args[0][0] == '-' {
		return "", fmt.Errorf("%s: опции не поддерживаются; используйте ./ для имени с дефисом", op)
	}
	return args[0], nil
}
