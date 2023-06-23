package component

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

type ConfigContext struct {
	root *tview.Table
}

func NewConfigContext(
	current int,
	confs []*config.Config,
	onSelect func(id int),
	onDelete func(name string, id int),
) *ConfigContext {
	log := logger.New()

	colHeaders := []string{"ID", "Name", "Target", "SSH-User", "SSH-Identity", "Overrides"}
	table := createTable("Context", colHeaders)

	table.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if evt.Key() == key.KeyCtrlD {
			row, _ := table.GetSelection()

			idStr := table.GetCell(row, 0).Text
			name := table.GetCell(row, 1).Text

			id, err := strconv.Atoi(idStr)

			if err != nil {
				log.Error().Err(err).Msg("failed to delete context")
				return nil
			}

			onDelete(name, id)

			return nil
		}

		if evt.Key() == key.KeyEnter {
			row, _ := table.GetSelection()
			idStr := table.GetCell(row, 0).Text
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Error().Err(err).Msg("failed to select new context")
				return nil
			}
			onSelect(id)
			return nil
		}

		return evt
	})

	c := &ConfigContext{root: table}
	c.UpdateConfigs(current, confs)

	return c
}

func (c *ConfigContext) UpdateConfigs(current int, confs []*config.Config) {
	c.clearRows()

	for rowIdx, conf := range confs {
		id := conf.ID
		idStr := strconv.Itoa(id)
		name := conf.Name
		target := strings.Join(conf.Targets, ",")
		sshUser := conf.SSH.User
		sshIdentity := conf.SSH.Identity
		overrides := "N"

		if len(conf.SSH.Overrides) > 0 {
			overrides = "Y"
		}

		row := []string{idStr, name, target, sshUser, sshIdentity, overrides}

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
