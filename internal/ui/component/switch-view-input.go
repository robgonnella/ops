package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

// SwitchViewInput toggle-able input for switching views
type SwitchViewInput struct {
	root     *tview.InputField
	showing  bool
	onSubmit func(text string)
}

// NewSwitchViewInput returns a new instance of SwitchViewInput
func NewSwitchViewInput(onSubmit func(text string)) *SwitchViewInput {

	input := tview.NewInputField()
	input.SetFieldStyle(style.StyleDefault.Dim(true))
	input.SetBorderPadding(0, 0, 1, 1)
	input.SetPlaceholderStyle(style.StyleDefault.Dim(true))

	// Show when focused
	input.SetFocusFunc(func() {
		input.SetBorder(true)
		input.SetBorderColor(style.ColorPurple)
		input.SetPlaceholder(
			"Enter view: servers, events, context, configure - type q | quit to quit",
		)
	})

	// hide when blurred
	input.SetBlurFunc(func() {
		input.SetBorder(false)
		input.SetPlaceholder("")
	})

	ai := &SwitchViewInput{
		root:     input,
		showing:  false,
		onSubmit: onSubmit,
	}

	// submit and then clear text when user presses enter
	ai.root.SetDoneFunc(func(k tcell.Key) {
		if k == key.KeyEnter {
			ai.onSubmit(ai.root.GetText())
			ai.root.SetText("")
		}
	})

	return ai
}

// Primitive returns the root primitive for SwitchViewInput
func (i *SwitchViewInput) Primitive() tview.Primitive {
	return i.root
}
