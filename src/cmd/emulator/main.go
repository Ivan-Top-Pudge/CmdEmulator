package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/config"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/gui"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
)

// main запускает приложение, выводит справку или ошибку запуска; ничего не возвращает.
// При ошибке завершает процесс с кодом 1, при справке или нормальном закрытии — с кодом 0.
func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Print(config.Usage)
			return
		}
		fmt.Fprintln(os.Stderr, "Ошибка запуска:", err)
		os.Exit(1)
	}
}

// run проверяет настройки и читает скрипт до создания GUI, затем запускает окно.
// Возвращает ошибку запуска либо nil после завершения сеанса; итоговые параметры выводит в stdout.
func run(args []string) error {
	const op = "main.run"
	settings, err := config.Load(args)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	commands, err := startup.LoadFile(settings.ScriptPath)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	fmt.Printf("Параметры запуска: %s\n", settings)
	a := app.New()
	w := gui.NewWindow(a, shell.New(), gui.Options{
		VFSName: settings.VFSName(), Prompt: settings.Prompt, Startup: commands,
	})
	w.ShowAndRun()
	return nil
}
