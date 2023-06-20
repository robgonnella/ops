package component

import (
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/ui/style"
)

type ModalButton struct {
	Label   string
	OnClick func()
}

type Modal struct {
	root *tview.Modal
}

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

	return &Modal{
		root: modal,
	}
}

func (m *Modal) Primitive() tview.Primitive {
	return m.root
}
