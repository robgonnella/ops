package component

import (
	"fmt"
	"strings"

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

type Header struct {
	root            *tview.Flex
	legendContainer *tview.Flex
	legendCol1      *tview.Flex
	legendCol2      *tview.Flex
	switchViewInput *SwitchViewInput
	targets         []string
	extraLegendMap  map[string]tview.Primitive
}

func NewHeader(userIP string, targets []string, onViewSwitch func(text string)) *Header {
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
				"IP: %s, Network Targets: %s",
				userIP,
				strings.Join(targets, ","),
			),
		)

	currentTarget.SetTextColor(style.ColorLightGreen)
	currentTarget.SetTextAlign(tview.AlignLeft)

	h.root.AddItem(currentTarget, 1, 1, false)
	h.root.AddItem(h.switchViewInput.Primitive(), 3, 1, false)

	h.targets = targets

	h.extraLegendMap = map[string]tview.Primitive{}

	return h
}

func (h *Header) Primitive() tview.Primitive {
	return h.root
}

func (h *Header) AddLegendKey(key, description string) {
	v := tview.NewTextView().
		SetText(fmt.Sprintf("\"%s\" - %s", key, description)).
		SetTextColor(style.ColorOrange).
		SetTextAlign(tview.AlignLeft)

	h.extraLegendMap[key] = v

	h.legendCol2.AddItem(v, 0, 1, false)
}

func (h *Header) RemoveLegendKey(key string) {
	for k, primitive := range h.extraLegendMap {
		if k == key {
			h.legendCol2.RemoveItem(primitive)
			delete(h.extraLegendMap, key)
		}
	}
}

func (h *Header) RemoveAllExtraLegendKeys() {
	for k, primitive := range h.extraLegendMap {
		h.legendCol2.RemoveItem(primitive)
		delete(h.extraLegendMap, k)
	}
}

func (h *Header) SwitchViewInput() *SwitchViewInput {
	return h.switchViewInput
}
