package shell

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestParse проверяет кавычки, экранирование и сохранение границ аргументов.
func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil}, {" \t\n ", nil}, {"ls", []string{"ls"}},
		{"\tls\t/каталог\t", []string{"ls", "/каталог"}},
		{`cd "my folder"`, []string{"cd", "my folder"}},
		{`ls 'my folder'`, []string{"ls", "my folder"}},
		{`ls my\ folder`, []string{"ls", "my folder"}},
		{`ls pre"middle"'end'`, []string{"ls", "premiddleend"}},
		{`ls "a\"b"`, []string{"ls", `a"b`}},
		{`ls '$HOME'`, []string{"ls", "$HOME"}},
		{`ls \$HOME`, []string{"ls", "$HOME"}},
		{`ls $`, []string{"ls", "$"}},
		{`ls ""`, []string{"ls", ""}},
	}
	for _, tt := range tests {
		got, err := parse(tt.input)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parse(%q) = %q, %v; want %q", tt.input, got, err, tt.want)
		}
	}
}

// TestParseEnvironment проверяет раскрытие переменных без повторного разделения их значений.
func TestParseEnvironment(t *testing.T) {
	t.Setenv("EMULATOR_TEST_PATH", "/my folder")
	t.Setenv("EMULATOR_TEST_EMPTY", "")
	t.Setenv("EMULATOR_TEST_COMMAND", "ls")
	tests := []struct {
		input string
		want  []string
	}{
		{`ls $EMULATOR_TEST_PATH`, []string{"ls", "/my folder"}},
		{`cd "${EMULATOR_TEST_PATH}/sub"`, []string{"cd", "/my folder/sub"}},
		{`ls '${EMULATOR_TEST_PATH}'`, []string{"ls", "${EMULATOR_TEST_PATH}"}},
		{`ls "\$EMULATOR_TEST_PATH"`, []string{"ls", "$EMULATOR_TEST_PATH"}},
		{`ls "${EMULATOR_TEST_EMPTY}/sub"`, []string{"ls", "/sub"}},
		{`$EMULATOR_TEST_COMMAND /`, []string{"ls", "/"}},
	}
	for _, tt := range tests {
		got, err := parse(tt.input)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parse(%q) = %q, %v; want %q", tt.input, got, err, tt.want)
		}
	}
}

// TestParseUnsetEnvironment проверяет пустую подстановку отсутствующей переменной.
func TestParseUnsetEnvironment(t *testing.T) {
	const missing = "EMULATOR_TEST_UNSET_VARIABLE"
	t.Setenv(missing, "")
	if err := os.Unsetenv(missing); err != nil {
		t.Fatal(err)
	}
	got, err := parse("ls ${" + missing + "}/sub")
	if err != nil || !reflect.DeepEqual(got, []string{"ls", "/sub"}) {
		t.Fatalf("неизвестная переменная: %q, %v", got, err)
	}
}

// TestExecuteSyntaxRecovery проверяет ошибки парсера и продолжение работы после них.
func TestExecuteSyntaxRecovery(t *testing.T) {
	tests := []struct{ input, message string }{
		{`ls "unfinished`, "незакрытая кавычка"},
		{`cd 'unfinished`, "незакрытая кавычка"},
		{`ls path\`, "после обратного слеша"},
		{`ls ${HOME`, "незакрытая переменная"},
		{`ls ${}`, "неверное имя переменной"},
		{`ls ${1HOME}`, "неверное имя переменной"},
		{`ls ${HOME-NAME}`, "неверное имя переменной"},
	}
	engine := New(nil)
	for _, tt := range tests {
		result, err := engine.Execute(tt.input)
		if err == nil || !strings.Contains(err.Error(), tt.message) || result != (Result{}) {
			t.Fatalf("Execute(%q) = %+v, %v", tt.input, result, err)
		}
		if !strings.HasPrefix(err.Error(), "shell.Execute: ") {
			t.Fatal(err)
		}
		if _, err := engine.Execute("ls /"); err != nil {
			t.Fatal(err)
		}
	}
}
