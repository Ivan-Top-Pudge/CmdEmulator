package startup_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Ivan-Top-Pudge/CmdEmulator/src/internal/startup"
)

// TestReadGoComments проверяет строковые, блочные и встроенные комментарии без вложенности.
func TestReadGoComments(t *testing.T) {
	tests := []struct {
		input string
		want  []startup.Command
	}{
		{"// line\n/* block */\n", nil},
		{"ls / // tail", []startup.Command{{Text: "ls / ", Line: 1}}},
		{"ls/* block */ /docs\n", []startup.Command{{Text: "ls  /docs", Line: 1}}},
		{"/* block */ls /", []startup.Command{{Text: " ls /", Line: 1}}},
		{"ls /* block */\"my folder\"", []startup.Command{{Text: `ls  "my folder"`, Line: 1}}},
		{"/* block\r\nignored\r\n*/\r\nls /\r\n", []startup.Command{{Text: "ls /", Line: 4}}},
		{"ls / /* block\n*/cd /\n", []startup.Command{{Text: "ls /  ", Line: 1}, {Text: "cd /", Line: 2}}},
		{"ls/* one *//* two *//docs", []startup.Command{{Text: "ls  /docs", Line: 1}}},
		{"/* outer /* inner */ls", []startup.Command{{Text: " ls", Line: 1}}},
		{"ls / // /*\ncd /\n", []startup.Command{{Text: "ls / ", Line: 1}, {Text: "cd /", Line: 2}}},
		{"/* // text\nstill block */ls\n", []startup.Command{{Text: "ls", Line: 2}}},
		{"# ordinary text\n", []startup.Command{{Text: "# ordinary text", Line: 1}}},
		{"ls / /* block */ // tail\n", []startup.Command{{Text: "ls /   ", Line: 1}}},
	}
	for _, tt := range tests {
		checkScriptCommands(t, tt.input, tt.want)
	}
}

// TestReadQuotedCommentMarkers проверяет защиту маркеров кавычками и обратным слешем.
func TestReadQuotedCommentMarkers(t *testing.T) {
	tests := []struct{ input, want string }{
		{`ls '//literal'`, `ls '//literal'`},
		{`ls "/*literal*/"`, `ls "/*literal*/"`},
		{`ls "a\"//name" // tail`, `ls "a\"//name" `},
		{`ls 'a\' // tail`, `ls 'a\' `},
		{`ls \/\/*name`, `ls \/\/*name`},
		{`ls \/\/name // tail`, `ls \/\/name `},
		{`ls 'one//two' "three/*four*/"`, `ls 'one//two' "three/*four*/"`},
		{`ls "two'//three"`, `ls "two'//three"`},
		{`ls 'two"/*three'`, `ls 'two"/*three'`},
	}
	for _, tt := range tests {
		checkScriptCommands(t, tt.input, []startup.Command{{Text: tt.want, Line: 1}})
	}
}

// TestReadQuoteStateIsPerLine оставляет ошибки кавычек обработчику команд, не затрагивая следующие строки.
func TestReadQuoteStateIsPerLine(t *testing.T) {
	input := "ls \"unfinished //literal\n// comment\nls /\n"
	want := []startup.Command{{Text: `ls "unfinished //literal`, Line: 1}, {Text: "ls /", Line: 3}}
	checkScriptCommands(t, input, want)
}

// TestReadUnclosedBlock проверяет ошибку с исходной строкой начала и отсутствие частично прочитанных команд.
func TestReadUnclosedBlock(t *testing.T) {
	for _, input := range []string{"ls /\n/* broken", "// /* ignored\n/*", "\n/* first\n/* not nested"} {
		commands, err := startup.Read(strings.NewReader(input))
		if commands != nil || err == nil {
			t.Fatalf("Read(%q) = %+v, %v; ожидалась ошибка без частичного результата", input, commands, err)
		}
		if !strings.Contains(err.Error(), "startup.Read: незакрытый блочный комментарий со строки 2") {
			t.Fatalf("нет исходного номера строки: %v", err)
		}
	}
}

// checkScriptCommands проверяет точные команды и исходные номера строк после обработки комментариев.
func checkScriptCommands(t *testing.T, input string, want []startup.Command) {
	t.Helper()
	got, err := startup.Read(strings.NewReader(input))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Read(%q) = %+v, %v; want %+v", input, got, err, want)
	}
}
