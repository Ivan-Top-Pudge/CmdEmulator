// Package startup читает и выполняет стартовые скрипты независимо от GUI.
package startup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
)

const (
	initialBufferSize = 4096
	maxLineSize       = 1024 * 1024
)

// Command содержит текст исполняемой строки и её исходный номер в скрипте.
type Command struct {
	Text string
	Line int
}

// Executor выполняет одну команду и возвращает вывод, признак выхода и ошибку.
type Executor interface {
	Execute(string) (shell.Result, error)
}

// LoadFile читает стартовый скрипт целиком до запуска GUI.
// Возвращает команды с номерами строк либо ошибку; пустой путь означает отсутствие скрипта.
func LoadFile(path string) ([]Command, error) {
	const op = "startup.LoadFile"
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: открыть скрипт %q: %w", op, path, err)
	}
	defer file.Close()
	commands, err := Read(file)
	if err != nil {
		return nil, fmt.Errorf("%s: скрипт %q: %w", op, path, err)
	}
	return commands, nil
}

// Read читает по одной команде на строку, пропуская пустые строки и комментарии с # в начале.
// Возвращает команды либо ошибку чтения; максимальная длина строки — менее 1 МиБ.
func Read(reader io.Reader) ([]Command, error) {
	const op = "startup.Read"
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, initialBufferSize), maxLineSize)
	var commands []Command
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		commands = append(commands, Command{Text: text, Line: line})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: чтение строки %d: %w", op, line+1, err)
	}
	return commands, nil
}

// Run последовательно выполняет команды и передаёт каждый результат в display.
// Ошибки команд не прерывают скрипт; возвращает true только при корректном exit.
func Run(commands []Command, executor Executor, display func(Command, shell.Result, error)) bool {
	const op = "startup.Run"
	for _, command := range commands {
		result, err := executor.Execute(command.Text)
		if err != nil {
			err = fmt.Errorf("%s: строка %d: %w", op, command.Line, err)
		}
		display(command, result, err)
		if err == nil && result.Exit {
			return true
		}
	}
	return false
}
