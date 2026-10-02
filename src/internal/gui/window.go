// Package gui предоставляет графический интерфейс эмулятора на Fyne.
package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
)

const (
	windowWidth  = 800
	windowHeight = 500
)

// Options задаёт имя VFS, приглашение и заранее прочитанные команды стартового скрипта.
type Options struct {
	VFSName string
	Prompt  string
	Startup []startup.Command
}

type terminal struct {
	window  fyne.Window
	engine  *shell.Shell
	input   *widget.Entry
	output  *widget.TextGrid
	button  *widget.Button
	options Options
}

// NewWindow создаёт и возвращает окно REPL с именем VFS в заголовке и обработчиками ввода.
// Скрипт выполнится при запуске цикла Fyne; до его окончания ручной ввод недоступен.
func NewWindow(application fyne.App, engine *shell.Shell, options Options) fyne.Window {
	t := &terminal{
		window: application.NewWindow("Эмулятор командной строки — " + options.VFSName),
		engine: engine, input: widget.NewEntry(), output: widget.NewTextGrid(), options: options,
	}
	t.output.ShowLineNumbers = false
	t.output.Scroll = container.ScrollBoth
	t.input.SetPlaceHolder("Введите команду...")
	// OnSubmitted отправляет команду по Enter и ничего не возвращает.
	t.input.OnSubmitted = func(string) { t.submit() }
	t.button = widget.NewButton("Выполнить", t.submit)
	controls := container.NewBorder(nil, nil, widget.NewLabel(options.Prompt), t.button, t.input)
	t.window.SetContent(container.NewBorder(nil, controls, nil, nil, t.output))
	t.window.Resize(fyne.NewSize(windowWidth, windowHeight))
	if len(options.Startup) != 0 {
		t.input.Disable()
		t.button.Disable()
		application.Lifecycle().SetOnStarted(t.runStartup)
	} else {
		t.window.Canvas().Focus(t.input)
	}
	return t.window
}

// submit выполняет команду из поля ввода, очищает поле и передаёт результат в историю.
// Пустой ввод игнорирует; ничего не возвращает, поскольку результат отображается в GUI.
func (t *terminal) submit() {
	command := t.input.Text
	t.input.SetText("")
	t.window.Canvas().Focus(t.input)
	if strings.TrimSpace(command) == "" {
		return
	}
	result, err := t.engine.Execute(command)
	t.display(command, result, err)
}

// display добавляет команду с приглашением и результат или ошибку в историю, затем прокручивает её вниз.
// При корректном exit закрывает окно; ничего не возвращает.
func (t *terminal) display(command string, result shell.Result, err error) {
	t.output.Append(t.options.Prompt + command)
	if err != nil {
		t.output.Append(fmt.Sprintf("Ошибка: %v", err))
	} else if result.Output != "" {
		t.output.Append(result.Output)
	}
	t.output.ScrollToBottom()
	if err == nil && result.Exit {
		t.window.Close()
	}
}

// runStartup выполняет скрипт после запуска Fyne и включает ручной ввод, если скрипт не завершил сеанс.
// Ничего не возвращает; каждая команда отображается тем же способом, что и ручной ввод.
func (t *terminal) runStartup() {
	// display показывает результат строки скрипта в истории и ничего не возвращает.
	display := func(command startup.Command, result shell.Result, err error) {
		t.display(command.Text, result, err)
	}
	if startup.Run(t.options.Startup, t.engine, display) {
		return
	}
	t.input.Enable()
	t.button.Enable()
	t.window.Canvas().Focus(t.input)
}
