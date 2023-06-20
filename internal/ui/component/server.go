package component

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

type ServerTable struct {
	ctx           context.Context
	cancel        context.CancelFunc
	table         *tview.Table
	columnHeaders []string
}

func NewServerTable(OnSSH func(ip string)) *ServerTable {
	columnHeaders := []string{"HOSTNAME", "IP", "ID", "OS", "SSH", "STATUS"}

	ctx, cancel := context.WithCancel(context.Background())

	table := createTable("servers", columnHeaders)

	table.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if evt.Key() == key.KeyCtrlS {
			row, _ := table.GetSelection()
			ip := table.GetCell(row, 1).Text
			OnSSH(ip)
			return nil
		}

		return evt
	})

	return &ServerTable{
		ctx:           ctx,
		cancel:        cancel,
		table:         table,
		columnHeaders: columnHeaders,
	}
}

func (t *ServerTable) Primitive() tview.Primitive {
	return t.table
}

func (t *ServerTable) UpdateTable(servers []*server.Server) {
	for rowIdx, svr := range servers {
		status := string(svr.Status)
		ssh := string(svr.SshStatus)
		hostname := svr.Hostname
		id := svr.ID
		ip := svr.IP
		os := svr.OS

		row := []string{hostname, ip, id, os, ssh, status}

		for col, text := range row {
			cell := tview.NewTableCell(text)
			cell.SetExpansion(1)
			cell.SetAlign(tview.AlignLeft)
			color := style.ColorWhite

			if text == "enabled" || text == "online" {
				color = style.ColorMediumGreen
			}

			if text == "disabled" || text == "offline" {
				color = style.ColorDimGrey
			}

			cell.SetTextColor(color)
			t.table.SetCell(rowIdx+2, col, cell)
		}
	}
}
