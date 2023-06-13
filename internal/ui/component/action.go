package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/ui/key"
	"github.com/robgonnella/opi/internal/ui/style"
)

type ActionInput struct {
	root     *tview.InputField
	showing  bool
	onSubmit func(text string)
}

func NewActionInput(onSubmit func(text string)) *ActionInput {

	input := tview.NewInputField()
	input.SetBorder(true)
	input.SetFieldStyle(style.StyleDefault.Dim(true))
	input.SetBorderPadding(0, 0, 1, 1)

	input.SetFocusFunc(func() {
		input.SetBorderColor(style.ColorPurple)
	})

	ai := &ActionInput{
		root:     input,
		showing:  false,
		onSubmit: onSubmit,
	}

	ai.root.SetDoneFunc(func(k tcell.Key) {
		if k == key.KeyEnter {
			ai.onSubmit(ai.root.GetText())
			ai.root.SetText("")
		}
	})

	return ai
}

func (i *ActionInput) Primitive() tview.Primitive {
	return i.root
}
