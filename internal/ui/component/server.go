package component

import (
	"bytes"
	"net"
	"slices"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/ui/key"
	"github.com/robgonnella/ops/internal/ui/style"
)

// ServerTable table displaying all servers for the active context
type ServerTable struct {
	table         *tview.Table
	columnHeaders []string
	hostIP        string
	hostHostname  string
	rows          [][]string
	mux           sync.RWMutex
}

// NewServerTable returns a new instance of ServerTable
func NewServerTable(hostHostname, hostIP string, OnSSH func(ip string)) *ServerTable {
	columnHeaders := []string{"HOSTNAME", "IP", "ID", "OS", "VENDOR", "SSH", "STATUS"}

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
		rows:          [][]string{},
		mux:           sync.RWMutex{},
	}
}

// Primitive returns the root primitive for ServerTable
func (t *ServerTable) Primitive() tview.Primitive {
	return t.table
}

// UpdateTable updates the table with the incoming list of servers from
// the database. We expect these servers to always be sorted so the ordering
// should remain relatively consistent.
func (t *ServerTable) UpdateTable(evt *event.Event) {
	payload, ok := evt.Payload.(*discovery.DiscoveryResult)

	if !ok {
		return
	}

	status := "offline"

	if payload.Status == discovery.ServerOnline {
		status = "online"
	}

	ssh := "disabled"

	if payload.Port.Status == discovery.PortOpen {
		ssh = "enabled"
	}

	hostname := payload.Hostname
	id := payload.ID
	ip := payload.IP
	os := payload.OS
	vendor := payload.Vendor

	row := []string{hostname, ip, id, os, vendor, ssh, status}

	t.mux.Lock()
	defer t.mux.Unlock()

	idx := slices.IndexFunc(t.rows, func(r []string) bool {
		return r[2] == id
	})

	exists := idx > -1
	isARP := evt.Type == discovery.DiscoveryArpUpdateEvent
	isSYN := evt.Type == discovery.DiscoverySynUpdateEvent

	if exists && isARP {
		// we already have this entry no need to do anything else
		return
	}

	if !exists && isARP {
		t.rows = append(t.rows, row)
	} else if exists && isSYN {
		r := t.rows[idx]
		// keep previous vendor as syn results don't have vendor
		row[4] = r[4]
		t.rows[idx] = row
	} else {
		// this should never happen
		return
	}

	slices.SortFunc(t.rows, func(r1, r2 []string) int {
		if isSYN && r1[5] == "enabled" && r2[5] == "disabled" {
			return -1
		}

		if isSYN && r1[5] == "disabled" && r2[5] == "enabled" {
			return 1
		}

		ip1 := net.ParseIP(r1[1])
		ip2 := net.ParseIP(r2[1])

		return bytes.Compare(ip1, ip2)
	})

	t.table.Clear()
	setTableHeaders(t.table, t.columnHeaders)

	for rowIdx, row := range t.rows {
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
