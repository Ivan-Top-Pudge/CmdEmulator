// Package gui предоставляет графический интерфейс эмулятора на Fyne.
package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
)

const (
	windowWidth  = 800
	windowHeight = 500
	prompt       = "$ "
)

// NewWindow создаёт и возвращает окно REPL с именем VFS в заголовке и обработчиками ввода.
// Окно не запускает цикл приложения: для этого вызывается ShowAndRun.
func NewWindow(application fyne.App, vfsName string, engine *shell.Shell) fyne.Window {
	window := application.NewWindow("Эмулятор командной строки — " + vfsName)
	output := widget.NewTextGrid()
	output.ShowLineNumbers = false
	output.Scroll = container.ScrollBoth
	input := widget.NewEntry()
	input.SetPlaceHolder("Введите команду...")
	// submit выполняет команду из поля ввода и ничего не возвращает.
	submit := func() { execute(window, engine, input, output) }
	// OnSubmitted отправляет команду по Enter и ничего не возвращает.
	input.OnSubmitted = func(string) { submit() }
	button := widget.NewButton("Выполнить", submit)
	controls := container.NewBorder(nil, nil, widget.NewLabel(prompt), button, input)
	window.SetContent(container.NewBorder(nil, controls, nil, nil, output))
	window.Resize(fyne.NewSize(windowWidth, windowHeight))
	window.Canvas().Focus(input)
	return window
}

// execute выполняет команду из поля ввода, добавляет результат или ошибку в историю и очищает поле.
// При exit закрывает окно; ничего не возвращает, поскольку результат отображается в GUI.
func execute(window fyne.Window, engine *shell.Shell, input *widget.Entry, output *widget.TextGrid) {
	command := input.Text
	input.SetText("")
	window.Canvas().Focus(input)
	if strings.TrimSpace(command) == "" {
		return
	}
	output.Append(prompt + command)
	result, err := engine.Execute(command)
	if err != nil {
		output.Append(fmt.Sprintf("Ошибка: %v", err))
	} else if result.Output != "" {
		output.Append(result.Output)
	}
	output.ScrollToBottom()
	if result.Exit {
		window.Close()
	}
}
