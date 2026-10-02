// Package shell обрабатывает команды эмулятора независимо от GUI.
package shell

import (
	"fmt"
	"strings"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// Result содержит вывод команды и признак завершения сеанса.
type Result struct {
	Output string
	Exit   bool
}

// Shell хранит VFS в памяти и выполняет команды эмулятора; до этапа 4 ls и cd остаются заглушками.
type Shell struct {
	filesystem *vfs.FS
}

// New сохраняет загруженную VFS и возвращает обработчик команд эмулятора.
// Для проверки заглушек без файловой системы допускается nil.
func New(filesystem *vfs.FS) *Shell {
	return &Shell{filesystem: filesystem}
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
	switch args[0] {
	case "ls", "cd":
		result, err := executeStub(args[0], args[1:])
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", op, err)
		}
		return result, nil
	case "exit":
		if len(args) != 1 {
			return Result{}, fmt.Errorf("%s: exit: команда не принимает аргументы", op)
		}
		return Result{Exit: true}, nil
	default:
		return Result{}, fmt.Errorf("%s: %q: неизвестная команда", op, args[0])
	}
}

// executeStub проверяет аргументы ls или cd и возвращает имя команды со списком аргументов.
// При неверных аргументах возвращает пустой Result и ошибку с именем операции.
func executeStub(command string, args []string) (Result, error) {
	const op = "shell.executeStub"
	const maxPaths = 1
	if len(args) > maxPaths {
		return Result{}, fmt.Errorf("%s: %s: ожидается не более одного пути", op, command)
	}
	if len(args) == maxPaths && (args[0] == "" || strings.HasPrefix(args[0], "-")) {
		return Result{}, fmt.Errorf("%s: %s: нужен непустой путь; опции пока не поддерживаются", op, command)
	}
	return Result{Output: fmt.Sprintf("%s: аргументы %q", command, args)}, nil
}
