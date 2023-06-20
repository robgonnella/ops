package component

import "github.com/rivo/tview"

type UIComponent interface {
	Primitive() tview.Primitive
}
