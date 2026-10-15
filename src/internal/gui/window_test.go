package gui_test

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/gui"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
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
	_, window, input, output, button := newConfiguredTestWindow(t, gui.Options{VFSName: "test-vfs", Prompt: "$ "})
	return window, input, output, button
}

// newConfiguredTestWindow создаёт окно с заданными настройками и регистрирует очистку приложения.
// Возвращает приложение, окно и элементы REPL; при отсутствии элементов завершает тест через t.
func newConfiguredTestWindow(t *testing.T, options gui.Options) (
	fyne.App, fyne.Window, *widget.Entry, *widget.TextGrid, *widget.Button,
) {
	t.Helper()
	application := test.NewApp()
	t.Cleanup(application.Quit)
	window := gui.NewWindow(application, shell.New(), options)
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
	return application, window, input, output, button
}

// TestWindowCustomPrompt проверяет пользовательское и явно пустое приглашение в истории.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowCustomPrompt(t *testing.T) {
	for _, prompt := range []string{"student> ", ""} {
		_, _, input, output, button := newConfiguredTestWindow(t, gui.Options{VFSName: "vfs.xml", Prompt: prompt})
		input.SetText("ls")
		test.Tap(button)
		if got, want := output.Text(), prompt+"ls\nls: аргументы []"; got != want {
			t.Fatalf("пользовательское приглашение: %q; want %q", got, want)
		}
	}
}

// TestWindowStartupThenInteractive проверяет запуск скрипта после старта GUI и возвращение к ручному вводу.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowStartupThenInteractive(t *testing.T) {
	commands, err := startup.Read(strings.NewReader("// comment\n/*\nexit\n*/\nls /\nunknown\nexit extra\ncd /\n"))
	if err != nil {
		t.Fatal(err)
	}
	application, window, input, output, button := newConfiguredTestWindow(t,
		gui.Options{VFSName: "script-vfs", Prompt: "script> ", Startup: commands})
	if output.Text() != "" || !input.Disabled() || !button.Disabled() {
		t.Fatal("скрипт не должен выполняться до запуска цикла Fyne")
	}
	window.Show()
	triggerStarted(t, application)
	for _, command := range []string{"ls /", "unknown", "exit extra", "cd /"} {
		if !strings.Contains(output.Text(), "script> "+command) {
			t.Fatalf("в истории нет команды %q: %q", command, output.Text())
		}
	}
	if !strings.Contains(output.Text(), "startup.Run: строка 6") {
		t.Fatalf("нет номера ошибочной строки: %q", output.Text())
	}
	assertInteractiveControls(t, window, input, button)
	input.SetText("ls folder")
	test.Tap(button)
	if !strings.HasSuffix(output.Text(), "script> ls folder\nls: аргументы [\"folder\"]") {
		t.Fatalf("нет результата ручной команды после скрипта: %q", output.Text())
	}
}

// TestWindowStartupExit проверяет закрытие окна из скрипта и пропуск строк после exit.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestWindowStartupExit(t *testing.T) {
	commands := []startup.Command{{Text: "ls", Line: 1}, {Text: "exit", Line: 2}, {Text: "cd /", Line: 3}}
	application, window, _, output, _ := newConfiguredTestWindow(t,
		gui.Options{VFSName: "script-vfs", Prompt: "$ ", Startup: commands})
	closed := false
	// OnClosed запоминает закрытие окна для проверки и ничего не возвращает.
	window.SetOnClosed(func() { closed = true })
	window.Show()
	triggerStarted(t, application)
	if !closed || strings.Contains(output.Text(), "cd /") || !strings.HasSuffix(output.Text(), "$ exit") {
		t.Fatalf("exit должен закрыть окно и остановить скрипт: closed=%v, history=%q", closed, output.Text())
	}
}

// triggerStarted вызывает сохранённый callback старта, поскольку тестовый драйвер не запускает настоящий цикл.
// Ничего не возвращает; при отсутствии callback завершает тест через t.
func triggerStarted(t *testing.T, application fyne.App) {
	t.Helper()
	lifecycle, ok := application.Lifecycle().(interface{ OnStarted() func() })
	if !ok || lifecycle.OnStarted() == nil {
		t.Fatal("не зарегистрирован обработчик запуска стартового скрипта")
	}
	lifecycle.OnStarted()()
}
