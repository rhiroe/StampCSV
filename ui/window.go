package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	stamper "StampCSV/csv"
)

// NewMainWindow はメインウィンドウを構築して返す。
func NewMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("StampCSV")
	w.Resize(fyne.NewSize(400, 200))

	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("CSVを保存するディレクトリを選択してください")

	// ディレクトリ選択ボタン
	browseBtn := widget.NewButtonWithIcon("選択", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dirEntry.SetText(uri.Path())
		}, w)
	})

	statusLabel := widget.NewLabel("")

	// スタンプボタン（IN）
	inBtn := widget.NewButton("IN（開始）", func() {
		dir := dirEntry.Text
		if dir == "" {
			statusLabel.SetText("ディレクトリを選択してください")
			return
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			statusLabel.SetText("ディレクトリ作成失敗: " + err.Error())
			return
		}
		if err := stamper.Stamp(dir, "in"); err != nil {
			statusLabel.SetText("エラー: " + err.Error())
			return
		}
		statusLabel.SetText("開始を記録しました")
	})

	// スタンプボタン（OUT）
	outBtn := widget.NewButton("OUT（終了）", func() {
		dir := dirEntry.Text
		if dir == "" {
			statusLabel.SetText("ディレクトリを選択してください")
			return
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			statusLabel.SetText("ディレクトリ作成失敗: " + err.Error())
			return
		}
		if err := stamper.Stamp(dir, "out"); err != nil {
			statusLabel.SetText("エラー: " + err.Error())
			return
		}
		statusLabel.SetText("終了を記録しました")
	})

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
	return w
}
