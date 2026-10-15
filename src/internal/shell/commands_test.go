package shell_test

import (
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// newCommandShell загружает общую VFS и возвращает независимый сеанс для проверки команд.
func newCommandShell(t *testing.T) *shell.Shell {
	t.Helper()
	filesystem, err := vfs.LoadFile("../../../testdata/vfs/commands.xml")
	if err != nil {
		t.Fatal(err)
	}
	return shell.New(filesystem)
}

// expectCommand проверяет точный вывод команды и отсутствие ошибки и выхода.
func expectCommand(t *testing.T, engine *shell.Shell, command, want string) {
	t.Helper()
	got, err := engine.Execute(command)
	if err != nil || got.Output != want || got.Exit {
		t.Fatalf("Execute(%q) = %+v, %v; want %q", command, got, err, want)
	}
}

// TestListFlags проверяет все сочетания трёх флагов, отдельные опции и подробные размеры.
func TestListFlags(t *testing.T) {
	const visible = "-dash\nbig.bin\ndocs\nempty.txt\nmy folder\nother\ntext.txt"
	const all = ".\n..\n-dash\n.hidden\nbig.bin\ndocs\nempty.txt\nmy folder\nother\ntext.txt"
	for _, flags := range []string{"", "-h", "-a", "-ha", "-ah"} {
		want := visible
		if strings.Contains(flags, "a") {
			want = all
		}
		expectCommand(t, newCommandShell(t), "ls "+flags, want)
	}
	for _, flags := range []string{"-l", "-lh", "-la", "-lha", "-ahl", "-l -h -a"} {
		engine := newCommandShell(t)
		got, err := engine.Execute("ls " + flags)
		if err != nil {
			t.Fatal(err)
		}
		size := "2048"
		if strings.Contains(flags, "h") {
			size = "2KiB"
		}
		if !strings.Contains(got.Output, "-rw-r--r-- "+size+" big.bin") {
			t.Fatalf("нет размера для %s: %q", flags, got.Output)
		}
		if strings.Contains(got.Output, " .hidden") != strings.Contains(flags, "a") {
			t.Fatalf("скрытые файлы для %s: %q", flags, got.Output)
		}
	}
}

// TestListPaths проверяет файл, пустой каталог, порядок аргументов и имена с дефисом и пробелом.
func TestListPaths(t *testing.T) {
	engine := newCommandShell(t)
	for _, tt := range []struct{ command, want string }{
		{"ls /docs", "reports"}, {"ls /other", "marker.txt"},
		{"ls /empty.txt", "empty.txt"}, {`ls "my folder"`, ""},
		{"ls -- -dash", "-dash"}, {"ls ./-dash", "-dash"},
		{"ls big.bin -lh", "-rw-r--r-- 2KiB big.bin"},
		{"ls -lh empty.txt", "-rw-r--r-- 0B empty.txt"},
		{"ls -lh text.txt", "-rw-r--r-- 10B text.txt"},
	} {
		expectCommand(t, engine, tt.command, tt.want)
	}
}

// TestChangeDirectory проверяет . и .. в любом месте пути, корень и независимость сеансов.
func TestChangeDirectory(t *testing.T) {
	engine := newCommandShell(t)
	for _, command := range []string{"cd /docs", "cd .", "cd ./reports/../reports/2026/./.."} {
		expectCommand(t, engine, command, "")
	}
	expectCommand(t, engine, "ls", "2026")
	expectCommand(t, engine, "cd ../reports/../../other", "")
	expectCommand(t, engine, "ls", "marker.txt")
	expectCommand(t, engine, "cd //docs///reports/2026/../..//", "")
	expectCommand(t, engine, "ls", "reports")
	expectCommand(t, newCommandShell(t), "ls docs", "reports")
	expectCommand(t, engine, "cd", "")
	expectCommand(t, engine, "cd ../../../../", "")
	expectCommand(t, engine, "ls other", "marker.txt")
}

// TestFailedDirectoryChange проверяет последовательную проверку компонентов и сохранение текущего каталога.
func TestFailedDirectoryChange(t *testing.T) {
	engine := newCommandShell(t)
	expectCommand(t, engine, "cd /other", "")
	for _, command := range []string{
		"cd /missing/../docs", "cd marker.txt/..", "cd marker.txt/.", "cd marker.txt/", "cd marker.txt",
		"cd ../docs/missing", "cd /docs/../missing", "cd /text.txt/../docs",
	} {
		if _, err := engine.Execute(command); err == nil {
			t.Fatalf("%q должен завершиться ошибкой", command)
		}
		expectCommand(t, engine, "ls", "marker.txt")
	}
}

// TestReadCommands проверяет суммы байтов, Unicode, LF и пустой файл.
func TestReadCommands(t *testing.T) {
	engine := newCommandShell(t)
	for _, tt := range []struct{ command, want string }{
		{"du docs", "8\tdocs/reports/2026\n8\tdocs/reports\n8\tdocs"},
		{"du /docs/reports/../reports/2026", "8\t/docs/reports/../reports/2026"},
		{"du /", "8\t/docs/reports/2026\n8\t/docs/reports\n8\t/docs\n0\t/my folder\n3\t/other\n2076\t/"},
		{"du text.txt", "10\ttext.txt"}, {`du "my folder"`, "0\tmy folder"},
		{"wc text.txt", "1 3 10 text.txt"}, {"wc empty.txt", "0 0 0 empty.txt"},
		{"wc /docs/reports/2026/summary.txt", "1 1 8 /docs/reports/2026/summary.txt"},
		{"cd docs/reports/2026", ""}, {"du", "8\t."},
		{"wc summary.txt", "1 1 8 summary.txt"},
	} {
		expectCommand(t, engine, tt.command, tt.want)
	}
}

// TestCommandErrors проверяет неверные аргументы, неизвестные флаги, отсутствующие пути и восстановление.
func TestCommandErrors(t *testing.T) {
	engine := newCommandShell(t)
	for _, command := range []string{
		"unknown", "ls one two", "cd one two", "du one two", "wc one two", "exit now",
		"ls -z", "ls -lahz", "ls --long", "cd -a", "du -h", "wc -l", "wc", "wc /",
		`ls ""`, `cd ""`, `du ""`, `wc ""`, "ls missing", "du missing", "wc missing",
		"ls text.txt/", "du text.txt/..", "wc text.txt/.", "ls 'bad\x00name'",
	} {
		result, err := engine.Execute(command)
		if err == nil || !strings.HasPrefix(err.Error(), "shell.Execute: ") || result != (shell.Result{}) {
			t.Fatalf("Execute(%q) = %+v, %v; ожидалась ошибка с контекстом", command, result, err)
		}
		expectCommand(t, engine, "ls other", "marker.txt")
	}
	for _, command := range []string{"", " \t\n"} {
		expectCommand(t, engine, command, "")
	}
	if result, err := engine.Execute("exit"); err != nil || !result.Exit {
		t.Fatalf("exit = %+v, %v", result, err)
	}
}
