package component

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/ui/style"
)

const appText = `
 ██████╗ ██████╗ ███████╗
██╔═══██╗██╔══██╗██╔════╝
██║   ██║██████╔╝███████╗
██║   ██║██╔═══╝ ╚════██║
╚██████╔╝██║     ███████║
 ╚═════╝ ╚═╝     ╚══════╝`

// Header shown above all views. Includes app title and dynamic key legend
type Header struct {
	root            *tview.Flex
	legendContainer *tview.Flex
	legendCol1      *tview.Flex
	legendCol2      *tview.Flex
	switchViewInput *SwitchViewInput
	cidr            string
	extraLegendMap  map[string]tview.Primitive
}

// NewHeader returns a new instance of Header
func NewHeader(
	userIP string,
	cidr string,
	onViewSwitch func(text string),
) *Header {
	h := &Header{}

	h.root = tview.NewFlex().SetDirection(tview.FlexRow)

	h.legendContainer = tview.NewFlex().SetDirection(tview.FlexColumn)

	h.legendCol1 = tview.NewFlex()

	h.legendCol2 = tview.NewFlex().SetDirection(tview.FlexRow)

	title := tview.NewTextView().
		SetText(appText).
		SetTextColor(style.ColorPurple)

	h.legendCol1.AddItem(title, 0, 1, false)

	emptyText := tview.NewTextView().SetText("")

	viewSwitchLegend := tview.NewTextView().SetText("\":\" - change views")
	viewSwitchLegend.SetTextColor(style.ColorOrange)
	viewSwitchLegend.SetTextAlign(tview.AlignLeft)

	h.legendCol2.AddItem(emptyText, 0, 1, false)
	h.legendCol2.AddItem(viewSwitchLegend, 0, 1, false)

	h.legendContainer.AddItem(h.legendCol1, 60, 1, false)
	h.legendContainer.AddItem(h.legendCol2, 0, 1, false)

	h.root.AddItem(h.legendContainer, 0, 1, false)

	switchViewInput := NewSwitchViewInput(onViewSwitch)

	h.switchViewInput = switchViewInput

	currentTarget := tview.NewTextView().
		SetText(
			fmt.Sprintf(
				"IP: %s, Network Target: %s",
				userIP,
				cidr,
			),
		)

	currentTarget.SetTextColor(style.ColorLightGreen)
	currentTarget.SetTextAlign(tview.AlignLeft)

	h.root.AddItem(currentTarget, 1, 1, false)
	h.root.AddItem(h.switchViewInput.Primitive(), 3, 1, false)

	h.cidr = cidr

	h.extraLegendMap = map[string]tview.Primitive{}

	return h
}

// Primitive returns the root primitive for Header
func (h *Header) Primitive() tview.Primitive {
	return h.root
}

// AddLegendKey adds a new key and description to the legend
func (h *Header) AddLegendKey(key, description string) {
	v := tview.NewTextView().
		SetText(fmt.Sprintf("\"%s\" - %s", key, description)).
		SetTextColor(style.ColorOrange).
		SetTextAlign(tview.AlignLeft)

	h.extraLegendMap[key] = v

	h.legendCol2.AddItem(v, 0, 1, false)
}

// RemoveLegendKey removes key and description from legend
func (h *Header) RemoveLegendKey(key string) {
	for k, primitive := range h.extraLegendMap {
		if k == key {
			h.legendCol2.RemoveItem(primitive)
			delete(h.extraLegendMap, key)
		}
	}
}

// RemoveAllExtraLegendKeys removes all non-default keys and descriptions
// from legend
func (h *Header) RemoveAllExtraLegendKeys() {
	for k, primitive := range h.extraLegendMap {
		h.legendCol2.RemoveItem(primitive)
		delete(h.extraLegendMap, k)
	}
}

// SwitchViewInput returns access to the Header's SwitchViewInput component
func (h *Header) SwitchViewInput() *SwitchViewInput {
	return h.switchViewInput
}
