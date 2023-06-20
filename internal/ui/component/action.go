package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/ui/style"
)

type ActionInput struct {
	root     *tview.InputField
	showing  bool
	OnSubmit func(text string)
}

func NewActionInput() *ActionInput {

	input := tview.NewInputField()
	input.SetBorder(true)
	input.SetFieldStyle(style.StyleDefault.Dim(true))
	input.SetBorderPadding(0, 0, 1, 1)

	input.SetFocusFunc(func() {
		input.SetBorderColor(style.ColorPurple)
	})

	ai := &ActionInput{
		root:    input,
		showing: false,
	}

	ai.setDoneFunc()

	return ai
}

func (i *ActionInput) Primitive() tview.Primitive {
	return i.root
}

func (i *ActionInput) setDoneFunc() {
	i.root.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter && i.OnSubmit != nil {
			i.OnSubmit(i.root.GetText())
			i.root.SetText("")
		}
	})
}
