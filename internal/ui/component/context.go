package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

// ConfigContext table selecting and deleting contexts (configurations)
type ConfigContext struct {
	root *tview.Table
}

// NewConfigContext returns a new instance of NewConfigContext
func NewConfigContext(
	current string,
	confs []*config.Config,
	onSelect func(id string),
	onDelete func(name string, id string),
) *ConfigContext {
	colHeaders := []string{"ID", "Name", "CIDR", "SSH-User", "SSH-Identity", "Overrides"}
	table := createTable("Context", colHeaders)

	table.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if evt.Rune() == key.Rune_d {
			row, _ := table.GetSelection()

			id := table.GetCell(row, 0).Text
			name := table.GetCell(row, 1).Text

			onDelete(name, id)

			return nil
		}

		if evt.Key() == key.KeyEnter {
			row, _ := table.GetSelection()
			id := table.GetCell(row, 0).Text
			onSelect(id)
			return nil
		}

		return evt
	})

	c := &ConfigContext{root: table}
	c.UpdateConfigs(current, confs)

	return c
}

// UpdateConfigs updates the table with a new list of contexts
func (c *ConfigContext) UpdateConfigs(current string, confs []*config.Config) {
	c.clearRows()

	for rowIdx, conf := range confs {
		id := conf.ID
		name := conf.Name
		iface := conf.Interface
		sshUser := conf.SSH.User
		sshIdentity := conf.SSH.Identity
		overrides := "N"

		if len(conf.SSH.Overrides) > 0 {
			overrides = "Y"
		}

		row := []string{id, name, iface, sshUser, sshIdentity, overrides}

		for col, text := range row {
			if id == current && col == 1 {
				text = text + " (selected)"
			}

			cell := tview.NewTableCell(text)
			cell.SetExpansion(1)
			cell.SetAlign(tview.AlignLeft)

			if id == current {
				cell.SetTextColor(style.ColorOrange)
			}

			c.root.SetCell(rowIdx+2, col, cell)
		}
	}
}

// Primitive returns the root primitive for ConfigContext
func (c *ConfigContext) Primitive() tview.Primitive {
	return c.root
}

// removes all row from table
func (c *ConfigContext) clearRows() {
	count := c.root.GetRowCount()

	// skip header rows
	for i := 2; i < count; i++ {
		c.root.RemoveRow(i)
	}
}
