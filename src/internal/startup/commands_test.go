package startup_test

import (
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// TestCommandDemos исполняет реальные демонстрационные скрипты этапов 4 и 5.
func TestCommandDemos(t *testing.T) {
	tests := []struct {
		stage  string
		errors map[string]bool
		last   string
	}{
		{"4", map[string]bool{
			"ls -z": true, "cd missing/../docs": true, "cd text.txt/..": true,
			"wc /docs": true, "wc missing.txt": true, "du missing": true,
		}, "marker.txt"},
		{"5", map[string]bool{
			"mkdir /session-only": true, "mkdir /text.txt": true, "mkdir /missing/child": true,
			"mkdir /missing/../bad": true, "mkdir /text.txt/../bad": true,
			"chmod 888 /text.txt": true, "chmod a+x /text.txt": true,
			"chmod 755 /missing": true, "chmod 755 /text.txt/..": true,
		}, "drwxr-x--- 0 .\ndrwxr-xr-x 0 ..\ndrwxr-x--- 0 nested"},
	}
	for _, tt := range tests {
		t.Run("stage"+tt.stage, func(t *testing.T) { runCommandDemo(t, tt.stage, tt.errors, tt.last) })
	}
}

// runCommandDemo проверяет ошибочные строки, итоговый вывод и продолжение сеанса после скрипта.
func runCommandDemo(t *testing.T, stage string, expectedErrors map[string]bool, wantLast string) {
	t.Helper()
	filesystem, err := vfs.LoadFile("../../../testdata/vfs/commands.xml")
	if err != nil {
		t.Fatal(err)
	}
	commands, err := startup.LoadFile("../../../scripts/stage" + stage + "-startup.txt")
	if err != nil {
		t.Fatal(err)
	}
	var last shell.Result
	failures := 0
	display := func(command startup.Command, result shell.Result, err error) {
		wantError := expectedErrors[command.Text]
		if (err != nil) != wantError {
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
	if last.Output != wantLast {
		t.Fatalf("последняя команда: %+v; want %q", last, wantLast)
	}
}
