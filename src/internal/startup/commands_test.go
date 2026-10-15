package startup_test

import (
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// TestStage4Demo исполняет реальный демонстрационный скрипт и проверяет ожидаемые ошибки и продолжение.
func TestStage4Demo(t *testing.T) {
	filesystem, err := vfs.LoadFile("../../../testdata/vfs/commands.xml")
	if err != nil {
		t.Fatal(err)
	}
	commands, err := startup.LoadFile("../../../scripts/stage4-startup.txt")
	if err != nil {
		t.Fatal(err)
	}
	expectedErrors := map[string]bool{
		"ls -z": true, "cd missing/../docs": true, "cd text.txt/..": true,
		"wc /docs": true, "wc missing.txt": true, "du missing": true,
	}
	var last shell.Result
	failures := 0
	display := func(command startup.Command, result shell.Result, err error) {
		if (err != nil) != expectedErrors[command.Text] {
			t.Errorf("строка %d (%s): неожиданный результат: %v", command.Line, command.Text, err)
		}
		if err != nil {
			failures++
			if !strings.Contains(err.Error(), "startup.Run: строка ") {
				t.Error(err)
			}
		}
		last = result
	}
	if startup.Run(commands, shell.New(filesystem), display) || failures != len(expectedErrors) {
		t.Fatal("сценарий должен продолжаться после ожидаемых ошибок")
	}
	if last.Output != "marker.txt" {
		t.Fatalf("последняя команда: %+v", last)
	}
}
