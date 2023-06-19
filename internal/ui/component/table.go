package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/ui/style"
)

func createTable(title string, columnHeaders []string) *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetFixed(2, 0).
		SetSelectable(true, false).
		SetSelectedStyle(style.StyleDefault.Background(style.ColorLightGreen).Bold(true))

	table.SetBorder(true)

	table.SetBorderPadding(2, 2, 2, 2)

	for c, h := range columnHeaders {
		cell := tview.NewTableCell(h)
		cell.SetExpansion(1)
		cell.SetAlign(tview.AlignLeft)
		cell.SetTextColor(style.ColorPurple)
		cell.SetSelectable(false)
		cell.SetAttributes(tcell.AttrBold)
		table.SetCell(0, c, cell)
	}

	for c := range columnHeaders {
		cell := tview.NewTableCell("")
		cell.SetExpansion(1)
		cell.SetAlign(tview.AlignLeft)
		cell.SetTextColor(style.ColorPurple)
		cell.SetSelectable(false)
		table.SetCell(1, c, cell)
	}

	table.SetBlurFunc(func() {
		table.SetBorderColor(style.ColorDefault)
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(style.ColorPurple)
	})

	table.SetTitle(title)
	table.SetTitleColor(style.ColorLightGreen)

	return table
}
