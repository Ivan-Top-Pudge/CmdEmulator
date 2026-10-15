package gui_test

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// assertInteractiveControls проверяет доступность поля и кнопки, а также фокус после стартового скрипта.
func assertInteractiveControls(t *testing.T, window fyne.Window, input *widget.Entry, button *widget.Button) {
	t.Helper()
	if input.Disabled() || button.Disabled() || window.Canvas().Focused() != input {
		t.Fatal("после скрипта ручной ввод должен быть доступен и иметь фокус")
	}
}
