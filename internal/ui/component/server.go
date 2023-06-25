package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

// ServerTable table displaying all servers for the active context
type ServerTable struct {
	table         *tview.Table
	columnHeaders []string
	hostIP        string
	hostHostname  string
}

// NewServerTable returns a new instance of ServerTable
func NewServerTable(hostHostname, hostIP string, OnSSH func(ip string)) *ServerTable {
	columnHeaders := []string{"HOSTNAME", "IP", "ID", "OS", "SSH", "STATUS"}

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
		table:         table,
		columnHeaders: columnHeaders,
		hostIP:        hostIP,
		hostHostname:  hostHostname,
	}
}

// Primitive returns the root primitive for ServerTable
func (t *ServerTable) Primitive() tview.Primitive {
	return t.table
}

// UpdateTable updates the table with the incoming list of servers from
// the database. We expect these servers to always be sorted so the ordering
// should remain relatively consistent.
func (t *ServerTable) UpdateTable(servers []*server.Server) {
	for rowIdx, svr := range servers {
		status := string(svr.Status)
		ssh := string(svr.SshStatus)
		hostname := svr.Hostname
		id := svr.ID
		ip := svr.IP
		os := svr.OS
		you := false

		if ip == t.hostIP {
			hostname = t.hostHostname
			you = true
		}

		row := []string{hostname, ip, id, os, ssh, status}

		for col, text := range row {
			if col == 0 && you {
				text = text + " (you)"
			}
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

			if you {
				color = style.ColorOrange
			}

			cell.SetTextColor(color)
			t.table.SetCell(rowIdx+2, col, cell)
		}
	}
}
