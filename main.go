package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"StampCSV/config"
	stamper "StampCSV/csv"
)

func main() {
	a := app.New()
	w := a.NewWindow("StampCSV")
	w.Resize(fyne.NewSize(420, 240))

	csvDir := config.LoadDir()

	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("CSVを保存するディレクトリを選択してください")
	dirEntry.SetText(csvDir)
	dirEntry.OnChanged = func(s string) {
		csvDir = s
		_ = config.SaveDir(s)
	}

	browseBtn := widget.NewButtonWithIcon("選択", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dirEntry.SetText(uri.Path())
			csvDir = uri.Path()
			_ = config.SaveDir(csvDir)
		}, w)
	})

	statusLabel := widget.NewLabel("")

	stamp := func(kind string) {
		dir := csvDir
		if dir == "" {
			statusLabel.SetText("ディレクトリを選択してください")
			return
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			statusLabel.SetText("ディレクトリ作成失敗: " + err.Error())
			return
		}
		if err := stamper.Stamp(dir, kind); err != nil {
			statusLabel.SetText("エラー: " + err.Error())
			return
		}
		if kind == "in" {
			statusLabel.SetText("開始を記録しました")
		} else {
			statusLabel.SetText("終了を記録しました")
		}
	}

	inBtn := widget.NewButton("IN（開始）", func() { stamp("in") })
	outBtn := widget.NewButton("OUT（終了）", func() { stamp("out") })
	inBtn.Importance = widget.HighImportance
	outBtn.Importance = widget.MediumImportance

	dirRow := container.NewBorder(nil, nil, nil, browseBtn, dirEntry)
	btnRow := container.NewGridWithColumns(2, inBtn, outBtn)
	content := container.NewVBox(
		widget.NewLabel("保存先ディレクトリ"),
		dirRow,
		btnRow,
		statusLabel,
	)

	w.SetContent(content)
	w.ShowAndRun()
}
