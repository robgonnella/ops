package component

import (
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/ui/style"
)

const appText = `
 ██████╗ ██████╗ ██╗
██╔═══██╗██╔══██╗██║
██║   ██║██████╔╝██║
██║   ██║██╔═══╝ ██║
╚██████╔╝██║     ██║
 ╚═════╝ ╚═╝     ╚═╝`

type Header struct {
	root              *tview.Flex
	legendContainer   *tview.Flex
	legendCol1        *tview.Flex
	legendCol2        *tview.Flex
	legendCol3        *tview.Flex
	switchViewInput   *ActionInput
	viewsText         *tview.TextView
	extraLegendMap    map[string]tview.Primitive
	showingSwitchView bool
}

func NewHeader(onViewSwitch func(text string)) *Header {
	h := &Header{}

	h.root = tview.NewFlex().SetDirection(tview.FlexRow)

	h.legendContainer = tview.NewFlex().SetDirection(tview.FlexColumn)

	h.legendCol1 = tview.NewFlex()

	h.legendCol2 = tview.NewFlex().SetDirection(tview.FlexRow)

	h.legendCol3 = tview.NewFlex().SetDirection(tview.FlexRow)

	emptyText := tview.NewTextView().SetText("")
	h.legendCol3.AddItem(emptyText, 0, 1, false)

	h.setDefaultLegend()

	h.root.AddItem(h.legendContainer, 0, 1, false)

	viewsText := tview.NewTextView().
		SetText("views: servers, events, context, configure")
	viewsText.SetTextColor(style.ColorOrange)

	switchViewInput := NewActionInput(onViewSwitch)

	h.switchViewInput = switchViewInput

	h.root.AddItem(h.switchViewInput.Primitive(), 3, 1, false)

	h.viewsText = viewsText

	h.showingSwitchView = false

	return h
}

func (h *Header) Primitive() tview.Primitive {
	return h.root
}

func (h *Header) ShowSwitchViewInput() {
	if !h.showingSwitchView {
		h.legendCol2.AddItem(h.viewsText, 0, 1, false)
		h.showingSwitchView = true
	}
}

func (h *Header) HideSwitchViewInput() {
	if h.showingSwitchView {
		h.legendCol2.RemoveItem(h.viewsText)
		h.showingSwitchView = false
	}
}

func (h *Header) ShowExtraLegend(legend map[string]string) {
	primitives := map[string]tview.Primitive{}

	for key, value := range legend {
		v := tview.NewTextView().
			SetText(key + " - " + value).
			SetTextStyle(style.StyleDefault.Dim(true))

		primitives[key] = v
		h.legendCol3.AddItem(v, 0, 1, false)
	}

	h.extraLegendMap = primitives
}

func (h *Header) RemoveExtraLegend() {
	for _, primitive := range h.extraLegendMap {
		h.legendCol3.RemoveItem(primitive)
	}

	h.extraLegendMap = map[string]tview.Primitive{}
}

func (h *Header) IsShowingSwitchViewInput() bool {
	return h.showingSwitchView
}

func (h *Header) SwitchViewInput() *ActionInput {
	return h.switchViewInput
}

func (h *Header) setDefaultLegend() {
	title := tview.NewTextView().
		SetText(appText).
		SetTextColor(style.ColorPurple)

	h.legendCol1.AddItem(title, 0, 1, false)

	emptyText := tview.NewTextView().SetText("")
	viewSwitchLegend := tview.NewTextView().SetText("type \":\" to change views")
	viewSwitchLegend.SetTextColor(style.ColorOrange)

	h.legendCol2.AddItem(emptyText, 0, 1, false)
	h.legendCol2.AddItem(viewSwitchLegend, 0, 1, false)

	h.legendContainer.AddItem(h.legendCol1, 36, 1, false)
	h.legendContainer.AddItem(h.legendCol2, 50, 1, false)
	h.legendContainer.AddItem(h.legendCol3, 0, 1, false)
}
