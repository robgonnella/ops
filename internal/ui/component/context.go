package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

type ConfigContext struct {
	root *tview.Table
}

func NewConfigContext(
	current string,
	confs []*config.Config,
	onSelect func(name string),
	onDelete func(name string),
) *ConfigContext {
	colHeaders := []string{"Name", "Target", "SSH-User", "SSH-Identity", "Overrides"}
	table := createTable("Context", colHeaders)

	table.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if evt.Key() == key.KeyCtrlD {
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

	c := &ConfigContext{root: table}
	c.UpdateConfigs(current, confs)

	return c
}

func (c *ConfigContext) UpdateConfigs(current string, confs []*config.Config) {
	c.clearRows()

	for rowIdx, conf := range confs {
		name := conf.Name
		target := strings.Join(conf.Targets, ",")
		sshUser := conf.SSH.User
		sshIdentity := conf.SSH.Identity
		overrides := "N"

		if len(conf.SSH.Overrides) > 0 {
			overrides = "Y"
		}

		row := []string{name, target, sshUser, sshIdentity, overrides}

		for col, text := range row {
			if name == current && col == 0 {
				text = text + " (selected)"
			}

			cell := tview.NewTableCell(text)
			cell.SetExpansion(1)
			cell.SetAlign(tview.AlignLeft)

			if name == current {
				cell.SetTextColor(style.ColorOrange)
				cell.SetSelectable(false)
			} else {
				cell.SetSelectable(true)
			}

			c.root.SetCell(rowIdx+2, col, cell)
		}
	}
}

func (c *ConfigContext) Primitive() tview.Primitive {
	return c.root
}

func (c *ConfigContext) clearRows() {
	count := c.root.GetRowCount()

	// skip header rows
	for i := 2; i < count; i++ {
		c.root.RemoveRow(i)
	}
}
