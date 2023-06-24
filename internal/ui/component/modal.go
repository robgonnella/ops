package component

import (
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/ui/style"
)

// ModalButton represents a button added to a modal
type ModalButton struct {
	Label   string
	OnClick func()
}

// Modal generic structure for displaying modals
type Modal struct {
	root *tview.Modal
}

// NewModal returns a new instance of Modal
func NewModal(message string, buttons []ModalButton) *Modal {
	modal := tview.NewModal()

	buttonLabels := []string{}

	for _, b := range buttons {
		buttonLabels = append(buttonLabels, b.Label)
	}

	modal.AddButtons(buttonLabels)

	modal.SetText(message)

	modal.SetDoneFunc(func(buttonIdx int, buttonLabel string) {
		for _, b := range buttons {
			if buttonLabel == b.Label {
				b.OnClick()
			}
		}
	})

	modal.SetBackgroundColor(style.ColorDefault).
		SetTextColor(style.ColorPurple).
		SetButtonBackgroundColor(style.ColorLightGreen).
		SetButtonTextColor(style.ColorBlack).
		SetBorderColor(style.ColorLightGreen)

	modal.SetButtonActivatedStyle(
		style.StyleDefault.Background(style.ColorLightGreen),
	)

	return &Modal{
		root: modal,
	}
}

// Primitive returns the root primitive for Modal
func (m *Modal) Primitive() tview.Primitive {
	return m.root
}
