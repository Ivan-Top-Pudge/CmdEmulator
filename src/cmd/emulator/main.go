package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Эмулятор командной строки — demo-vfs")

	textArea := widget.NewTextGrid()

	input := widget.NewEntry()
	input.SetPlaceHolder("Введите команду...")

	button := widget.NewButton("Submit", func() {
		log.Println("User entered:", input.Text)
		textArea.Append(input.Text)
		input.SetText("") // Access the input value via .Text
	})

	w.SetContent(container.NewGridWithRows(2,
		textArea,
		container.NewGridWithColumns(2,
			input, button),
	),
	)

	w.ShowAndRun()
}
