package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/ui/key"
)

type ConfigContext struct {
	root *tview.Table
}

func NewConfigContext(confs []*config.Config, onSelect func(name string), onDelete func(name string)) *ConfigContext {
	colHeaders := []string{"Name", "Target", "SSH-User", "SSH-Identity", "Overrides"}
	table := createTable("Context", colHeaders)

	table.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if evt.Rune() == key.RuneD {
			row, _ := table.GetSelection()
			name := table.GetCell(row, 0).Text
			onDelete(name)
			return nil
		}

		if evt.Key() == key.KeyEnter {
			row, _ := table.GetSelection()
			name := table.GetCell(row, 0).Text
			onSelect(name)
			return nil
		}

		return evt
	})

	for rowIdx, c := range confs {
		name := c.Name
		target := strings.Join(c.Targets, ",")
		sshUser := c.SSH.User
		sshIdentity := c.SSH.Identity
		overrides := "N"

		if len(c.SSH.Overrides) > 0 {
			overrides = "Y"
		}

		row := []string{name, target, sshUser, sshIdentity, overrides}

		for col, text := range row {
			cell := tview.NewTableCell(text)
			cell.SetExpansion(1)
			cell.SetAlign(tview.AlignLeft)
			table.SetCell(rowIdx+2, col, cell)
		}
	}

	return &ConfigContext{root: table}
}

func (c *ConfigContext) Primitive() tview.Primitive {
	return c.root
}
