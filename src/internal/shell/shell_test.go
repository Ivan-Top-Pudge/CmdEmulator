package shell_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/shell"
)

// TestExecuteCommands проверяет команды, кавычки, экранирование и пустой ввод.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestExecuteCommands(t *testing.T) {
	tests := []struct {
		input string
		want  shell.Result
	}{
		{"", shell.Result{}},
		{" \t\n ", shell.Result{}},
		{"ls", shell.Result{Output: `ls: аргументы []`}},
		{"cd", shell.Result{Output: `cd: аргументы []`}},
		{"\tls\t/каталог\t", shell.Result{Output: `ls: аргументы ["/каталог"]`}},
		{`cd "my folder"`, shell.Result{Output: `cd: аргументы ["my folder"]`}},
		{`ls 'my folder'`, shell.Result{Output: `ls: аргументы ["my folder"]`}},
		{`ls my\ folder`, shell.Result{Output: `ls: аргументы ["my folder"]`}},
		{`ls pre"middle"'end'`, shell.Result{Output: `ls: аргументы ["premiddleend"]`}},
		{`ls "a\"b"`, shell.Result{Output: `ls: аргументы ["a\"b"]`}},
		{`ls '$HOME'`, shell.Result{Output: `ls: аргументы ["$HOME"]`}},
		{`ls \$HOME`, shell.Result{Output: `ls: аргументы ["$HOME"]`}},
		{`ls $`, shell.Result{Output: `ls: аргументы ["$"]`}},
		{"exit", shell.Result{Exit: true}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := shell.New().Execute(tt.input)
			if err != nil || got != tt.want {
				t.Fatalf("Execute(%q) = %+v, %v; want %+v", tt.input, got, err, tt.want)
			}
		})
	}
}

// TestExecuteEnvironment проверяет раскрытие переменных и сохранение значения в одном аргументе.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestExecuteEnvironment(t *testing.T) {
	t.Setenv("EMULATOR_TEST_PATH", "/my folder")
	t.Setenv("EMULATOR_TEST_EMPTY", "")
	t.Setenv("EMULATOR_TEST_COMMAND", "ls")
	tests := []struct{ input, output string }{
		{`ls $EMULATOR_TEST_PATH`, `ls: аргументы ["/my folder"]`},
		{`cd "${EMULATOR_TEST_PATH}/sub"`, `cd: аргументы ["/my folder/sub"]`},
		{`ls '${EMULATOR_TEST_PATH}'`, `ls: аргументы ["${EMULATOR_TEST_PATH}"]`},
		{`ls "\$EMULATOR_TEST_PATH"`, `ls: аргументы ["$EMULATOR_TEST_PATH"]`},
		{`ls "${EMULATOR_TEST_EMPTY}/sub"`, `ls: аргументы ["/sub"]`},
		{`$EMULATOR_TEST_COMMAND /`, `ls: аргументы ["/"]`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := shell.New().Execute(tt.input)
			if err != nil || got.Output != tt.output || got.Exit {
				t.Fatalf("Execute(%q) = %+v, %v; want %q", tt.input, got, err, tt.output)
			}
		})
	}
}

// TestExecuteErrorsAndRecovery проверяет контекст ошибок и выполнение команды после ошибочного ввода.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestExecuteErrorsAndRecovery(t *testing.T) {
	t.Setenv("EMULATOR_TEST_EMPTY", "")
	tests := []struct{ input, message string }{
		{"unknown", "неизвестная команда"},
		{"ls one two", "не более одного пути"},
		{"cd one two", "не более одного пути"},
		{"exit now", "не принимает аргументы"},
		{"ls -a", "опции пока не поддерживаются"},
		{`cd ""`, "непустой путь"},
		{`ls $EMULATOR_TEST_EMPTY`, "непустой путь"},
		{"$EMULATOR_TEST_EMPTY", "неизвестная команда"},
		{`ls "unfinished`, "незакрытая кавычка"},
		{`cd 'unfinished`, "незакрытая кавычка"},
		{`ls path\`, "после обратного слеша"},
		{`ls ${HOME`, "незакрытая переменная"},
		{`ls ${}`, "неверное имя переменной"},
		{`ls ${1HOME}`, "неверное имя переменной"},
		{`ls ${HOME-NAME}`, "неверное имя переменной"},
	}
	engine := shell.New()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := engine.Execute(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.message) || got != (shell.Result{}) {
				t.Fatalf("Execute(%q) = %+v, %v; want error containing %q", tt.input, got, err, tt.message)
			}
			if !strings.HasPrefix(err.Error(), "shell.Execute: ") {
				t.Fatalf("ошибка не содержит имя операции: %v", err)
			}
			if got, err := engine.Execute("ls /"); err != nil || got.Output != `ls: аргументы ["/"]` {
				t.Fatalf("команда после ошибки: %+v, %v", got, err)
			}
		})
	}
}

// TestExecuteUnsetEnvironment проверяет, что отсутствующая переменная раскрывается в пустую строку.
// Ничего не возвращает; о расхождениях сообщает через t.
func TestExecuteUnsetEnvironment(t *testing.T) {
	const missing = "EMULATOR_TEST_UNSET_VARIABLE"
	t.Setenv(missing, "")
	// Unsetenv проверяется отдельно от переменной с явно заданным пустым значением.
	if err := os.Unsetenv(missing); err != nil {
		t.Fatal(err)
	}
	got, err := shell.New().Execute("ls ${" + missing + "}/sub")
	if err != nil || got.Output != `ls: аргументы ["/sub"]` {
		t.Fatalf("неизвестная переменная: %+v, %v", got, err)
	}
}
