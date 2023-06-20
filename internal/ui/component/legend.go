package component

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/ui/style"
)

type Legend struct {
	root        *tview.Flex
	networkInfo *tview.TextView
}

func NewLegend(networkTarget string, views []string) *Legend {
	root := tview.NewFlex()

	left := tview.NewFlex().SetDirection(tview.FlexRow)
	right := tview.NewFlex().SetDirection(tview.FlexRow)

	colonInstructions := tview.NewTextView()
	colonInstructions.SetTextStyle(style.StyleDefault.Attributes(tcell.AttrDim))
	colonInstructions.SetText("\":\" show actions input")
	colonInstructions.SetBorderPadding(0, 0, 3, 0)

	sshInstructions := tview.NewTextView()
	sshInstructions.SetTextStyle(style.StyleDefault.Attributes(tcell.AttrDim))
	sshInstructions.SetText("\"s\" ssh to selected server")
	sshInstructions.SetBorderPadding(0, 0, 3, 0)

	actions := tview.NewTextView()
	actions.SetTextStyle(style.StyleDefault.Attributes(tcell.AttrDim))
	actions.SetText("actions: " + strings.Join(views, " "))
	actions.SetBorderPadding(0, 1, 0, 0)
	actions.SetBorderPadding(0, 0, 3, 0)

	networkInfo := tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetTextStyle(style.StyleDefault.Foreground(style.ColorLightGreen)).
		SetText(fmt.Sprintf("Scanning %s…", networkTarget))
	networkInfo.SetBorderPadding(0, 0, 0, 3)

	left.AddItem(colonInstructions, 0, 1, false)
	left.AddItem(sshInstructions, 0, 1, false)
	left.AddItem(actions, 0, 1, false)

	right.AddItem(networkInfo, 0, 1, false)

	root.AddItem(left, 0, 1, false)
	root.AddItem(right, 0, 1, false)

	return &Legend{
		root:        root,
		networkInfo: networkInfo,
	}
}

func (l *Legend) Primitive() tview.Primitive {
	return l.root
}

func (l *Legend) SetTargets(targets []string) {
	targetStr := strings.Join(targets, ",")
	l.networkInfo.SetText(fmt.Sprintf("Scanning %s…", targetStr))
}
