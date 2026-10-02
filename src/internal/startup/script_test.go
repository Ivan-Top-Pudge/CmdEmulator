package startup_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
)

// TestReadCommentsAndLines проверяет CRLF, пропуск комментариев и сохранение номеров строк.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestReadCommentsAndLines(t *testing.T) {
	input := "# комментарий\r\n\r\n  # ещё комментарий\r\nls /\r\ncd '#folder'\r\n  ls\r\n"
	got, err := startup.Read(strings.NewReader(input))
	want := []startup.Command{{Text: "ls /", Line: 4}, {Text: "cd '#folder'", Line: 5}, {Text: "  ls", Line: 6}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Read = %+v, %v; want %+v", got, err, want)
	}
}

// TestRunErrorsAndExit проверяет продолжение после ошибок и остановку на корректном exit.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestRunErrorsAndExit(t *testing.T) {
	commands, err := startup.Read(strings.NewReader("unknown\nexit extra\nls /\nexit\ncd /\n"))
	if err != nil {
		t.Fatal(err)
	}
	var displayed []string
	var messages []error
	// display запоминает результаты команд для проверки порядка; ничего не возвращает.
	display := func(command startup.Command, result shell.Result, err error) {
		displayed = append(displayed, command.Text)
		messages = append(messages, err)
	}
	if !startup.Run(commands, shell.New(), display) {
		t.Fatal("exit должен завершить сеанс")
	}
	if want := []string{"unknown", "exit extra", "ls /", "exit"}; !reflect.DeepEqual(displayed, want) {
		t.Fatalf("порядок команд: %q; want %q", displayed, want)
	}
	if messages[0] == nil || !strings.Contains(messages[0].Error(), "строка 1") || messages[1] == nil {
		t.Fatalf("ошибки должны содержать номера строк: %v", messages)
	}
	if messages[2] != nil || messages[3] != nil {
		t.Fatalf("штатные команды не должны давать ошибки: %v", messages)
	}
}

// TestRunKeepsSession проверяет возврат в интерактивный режим после обычного или пустого скрипта.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestRunKeepsSession(t *testing.T) {
	for _, input := range []string{"", "# only comment\n", "unknown\nls\n"} {
		commands, err := startup.Read(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		// display принимает результат без сохранения; ничего не возвращает.
		display := func(startup.Command, shell.Result, error) {}
		if startup.Run(commands, shell.New(), display) {
			t.Fatalf("скрипт без exit завершил сеанс: %q", input)
		}
	}
}

// TestReadFailure проверяет ошибку чтения и слишком длинную строку без частичного выполнения.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestReadFailure(t *testing.T) {
	const oversizedLineLength = 1024 * 1024
	_, err := startup.Read(strings.NewReader(strings.Repeat("x", oversizedLineLength)))
	if err == nil || !strings.Contains(err.Error(), "startup.Read: чтение строки 1") {
		t.Fatalf("слишком длинная строка должна дать ошибку: %v", err)
	}
	want := errors.New("ошибка источника")
	_, err = startup.Read(failingReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("причина ошибки чтения потеряна: %v", err)
	}
}

type failingReader struct{ err error }

// Read имитирует сбой источника и возвращает ноль прочитанных байтов с заданной ошибкой.
func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

// TestLoadFile проверяет чтение файла, отсутствие скрипта и ошибку отсутствующего файла.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestLoadFile(t *testing.T) {
	if commands, err := startup.LoadFile(""); err != nil || len(commands) != 0 {
		t.Fatalf("пустой путь: %+v, %v", commands, err)
	}
	path := filepath.Join(t.TempDir(), "commands.txt")
	if _, err := startup.LoadFile(path); err == nil || !strings.Contains(err.Error(), "startup.LoadFile") {
		t.Fatalf("отсутствующий файл: %v", err)
	}
	const fileMode = 0644
	if err := os.WriteFile(path, []byte("ls\n"), fileMode); err != nil {
		t.Fatal(err)
	}
	commands, err := startup.LoadFile(path)
	if err != nil || len(commands) != 1 || commands[0].Text != "ls" {
		t.Fatalf("чтение скрипта: %+v, %v", commands, err)
	}
}
