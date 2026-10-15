package shell_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/vfs"
)

// TestMakeDirectory проверяет абсолютные/относительные пути, пробелы, . и .., default mode и отсутствие -p.
func TestMakeDirectory(t *testing.T) {
	engine := newCommandShell(t)
	expectCommand(t, engine, "mkdir /new", "")
	expectCommand(t, engine, "cd /new", "")
	expectCommand(t, engine, `mkdir "my child/"`, "")
	expectCommand(t, engine, "ls -l", "drwxr-xr-x 0 my child")
	expectCommand(t, engine, "mkdir ../docs/../sibling", "")
	expectCommand(t, engine, "mkdir ./../docs/reports/../extra", "")
	expectCommand(t, engine, "ls /docs", "extra\nreports")
	expectCommand(t, engine, "cd /sibling", "")
	expectCommand(t, engine, "ls -a", ".\n..")
	expectCommand(t, engine, "du /new", "0\t/new/my child\n0\t/new")
}

// TestChangeMode проверяет все классы прав, корень, текущий каталог и сохранение типа и данных.
func TestChangeMode(t *testing.T) {
	engine := newCommandShell(t)
	for _, tt := range []struct{ command, want string }{
		{"chmod 600 text.txt", ""}, {"ls -l text.txt", "-rw------- 10 text.txt"},
		{"wc text.txt", "1 3 10 text.txt"}, {"chmod 000 text.txt", ""},
		{"ls -l text.txt", "---------- 10 text.txt"}, {"chmod 777 text.txt", ""},
		{"ls -l text.txt", "-rwxrwxrwx 10 text.txt"}, {"chmod 700 docs", ""},
		{"cd docs", ""}, {"chmod 711 .", ""}, {"ls -la", "drwx--x--x 0 .\ndrwxr-xr-x 0 ..\ndrwxr-x--- 0 reports"},
		{"chmod 000 reports/2026/..", ""}, {"cd reports", ""},
		{"ls -la", "d--------- 0 .\ndrwx--x--x 0 ..\ndrwxr-xr-x 0 2026"},
		{"chmod 750 /", ""}, {"cd /", ""},
	} {
		expectCommand(t, engine, tt.command, tt.want)
	}
	result, err := engine.Execute("ls -la")
	if err != nil || !strings.HasPrefix(result.Output, "drwxr-x--- 0 .\ndrwxr-x--- 0 ..\n") {
		t.Fatalf("тип и права корня: %+v, %v", result, err)
	}
}

// TestMutationErrorsAreAtomic проверяет аргументы и отсутствие частичных изменений при ошибке.
func TestMutationErrorsAreAtomic(t *testing.T) {
	engine := newCommandShell(t)
	before, err := engine.Execute("ls -la")
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		"mkdir", `mkdir ""`, "mkdir one two", "mkdir -p", "mkdir -p nested/new",
		"mkdir missing/new", "mkdir missing/../new", "mkdir text.txt/new", "mkdir text.txt/../new",
		"mkdir docs", "mkdir text.txt", "mkdir /", "mkdir .", "mkdir ..", "mkdir docs/.",
		"mkdir 'bad\x00name'", "mkdir '   '", "chmod", "chmod 755", "chmod 755 docs other",
		"chmod 888 docs", "chmod 77 docs", "chmod 0755 docs", "chmod +77 docs", `chmod "" docs`,
		"chmod a+x docs", "chmod 755 missing", `chmod 755 ""`, "chmod 755 text.txt/..",
		"chmod 755 missing/../docs", "chmod -R docs", "chmod 755 -dash", "ls -ld docs",
	} {
		result, err := engine.Execute(command)
		if err == nil || result != (shell.Result{}) || !strings.HasPrefix(err.Error(), "shell.Execute: ") {
			t.Fatalf("Execute(%q) = %+v, %v", command, result, err)
		}
		expectCommand(t, engine, "ls -la", before.Output)
	}
	expectCommand(t, engine, "mkdir ./-folder", "")
	expectCommand(t, engine, "chmod 700 ./-folder", "")
	expectCommand(t, engine, "chmod 600 ./-dash", "")
	expectCommand(t, engine, "ls -l ./-dash", "-rw------- 0 -dash")
}

// TestMutationsStayInMemory проверяет исходный XML, отсутствие распаковки и сброс изменений при загрузке.
func TestMutationsStayInMemory(t *testing.T) {
	original, err := os.ReadFile("../../../testdata/vfs/commands.xml")
	if err != nil {
		t.Fatal(err)
	}
	folder := t.TempDir()
	path := filepath.Join(folder, "source.xml")
	const fileMode = 0644
	if err := os.WriteFile(path, original, fileMode); err != nil {
		t.Fatal(err)
	}
	filesystem, err := vfs.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := shell.New(filesystem)
	expectCommand(t, engine, "mkdir session-only", "")
	expectCommand(t, engine, "chmod 700 text.txt", "")
	expectCommand(t, engine, "ls -l text.txt", "-rwx------ 10 text.txt")
	after, err := os.ReadFile(path)
	if err != nil || !reflect.DeepEqual(after, original) {
		t.Fatalf("XML изменился: %v", err)
	}
	assertFreshFilesystem(t, path)
	files, err := os.ReadDir(folder)
	if err != nil || len(files) != 1 || files[0].Name() != "source.xml" {
		t.Fatalf("созданы физические файлы VFS: %+v, %v", files, err)
	}
}

// assertFreshFilesystem проверяет, что повторная загрузка не содержит изменений завершённого сеанса.
func assertFreshFilesystem(t *testing.T, path string) {
	t.Helper()
	filesystem, err := vfs.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fresh := shell.New(filesystem)
	expectCommand(t, fresh, "ls -l text.txt", "-rw-r--r-- 10 text.txt")
	if _, err := fresh.Execute("ls session-only"); err == nil {
		t.Fatal("изменения прошлого сеанса попали в исходную VFS")
	}
}
