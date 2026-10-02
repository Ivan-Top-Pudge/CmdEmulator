package gui_test

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/gui"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
)

// TestWindowCommandDialogue проверяет заголовок, отправку команд, историю и фокус поля ввода.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowCommandDialogue(t *testing.T) {
	window, input, output, button := newTestWindow(t)
	if !strings.Contains(window.Title(), "test-vfs") {
		t.Fatalf("имя VFS отсутствует в заголовке: %q", window.Title())
	}
	test.Type(input, "ls /")
	input.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if got := output.Text(); got != "$ ls /\nls: аргументы [\"/\"]" {
		t.Fatalf("диалог после Enter: %q", got)
	}
	if input.Text != "" || window.Canvas().Focused() != input {
		t.Fatal("после команды поле должно быть пустым и сохранять фокус")
	}
	test.Type(input, "cd folder")
	test.Tap(button)
	if !strings.HasSuffix(output.Text(), "$ cd folder\ncd: аргументы [\"folder\"]") {
		t.Fatalf("диалог после кнопки: %q", output.Text())
	}
}

// TestWindowErrorsAndEmptyInput проверяет пустой ввод, вывод ошибок и продолжение работы REPL.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowErrorsAndEmptyInput(t *testing.T) {
	_, input, output, button := newTestWindow(t)
	input.SetText(" \t ")
	test.Tap(button)
	if output.Text() != "" {
		t.Fatalf("пустой ввод изменил историю: %q", output.Text())
	}
	for _, command := range []string{"unknown", "exit now", `ls "unfinished`} {
		input.SetText(command)
		test.Tap(button)
		if !strings.Contains(output.Text(), "$ "+command+"\nОшибка: ") {
			t.Fatalf("нет сообщения об ошибке: %q", output.Text())
		}
	}
	input.SetText("ls")
	test.Tap(button)
	if !strings.HasSuffix(output.Text(), "ls: аргументы []") {
		t.Fatalf("REPL не продолжил работу после ошибки: %q", output.Text())
	}
}

// TestWindowExit проверяет, что окно закрывается только при корректной команде exit.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowExit(t *testing.T) {
	window, input, _, button := newTestWindow(t)
	closed := false
	window.SetOnClosed(func() { closed = true })
	input.SetText("exit extra")
	test.Tap(button)
	if closed {
		t.Fatal("exit с аргументами закрыл окно")
	}
	input.SetText("exit")
	test.Tap(button)
	if !closed {
		t.Fatal("exit не закрыл окно")
	}
}

// newTestWindow создаёт окно с тестовым драйвером и регистрирует очистку приложения.
// Возвращает окно, поле ввода, историю и кнопку; при отсутствии элементов завершает тест через t.
func newTestWindow(t *testing.T) (fyne.Window, *widget.Entry, *widget.TextGrid, *widget.Button) {
	t.Helper()
	application := test.NewApp()
	t.Cleanup(application.Quit)
	window := gui.NewWindow(application, "test-vfs", shell.New())
	var input *widget.Entry
	var output *widget.TextGrid
	var button *widget.Button
	for _, object := range test.LaidOutObjects(window.Content()) {
		switch object := object.(type) {
		case *widget.Entry:
			input = object
		case *widget.TextGrid:
			output = object
		case *widget.Button:
			button = object
		}
	}
	if input == nil || output == nil || button == nil {
		t.Fatal("в окне отсутствуют элементы REPL")
	}
	return window, input, output, button
}
