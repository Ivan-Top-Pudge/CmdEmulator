package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/gui"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
)

// main создаёт приложение и запускает окно эмулятора до его закрытия; ничего не возвращает.
func main() {
	a := app.New()
	w := gui.NewWindow(a, "demo-vfs", shell.New())
	w.ShowAndRun()
}
