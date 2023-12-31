package component

import (
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/ui/style"
)

// EventTable table for viewing all incoming events in realtime
type EventTable struct {
	table         *tview.Table
	columnHeaders []string
	count         uint
	maxEvents     uint
}

// NewEventTable returns a new instance of EventTable
func NewEventTable() *EventTable {
	columnHeaders := []string{
		"EVENT TYPE",
		"HOSTNAME",
		"IP",
		"ID",
		"OS",
		"VENDOR",
		"SSH",
		"STATUS",
	}

	return &EventTable{
		table:         createTable("events", columnHeaders),
		columnHeaders: columnHeaders,
		count:         0,
		maxEvents:     50,
	}
}

// Primitive returns the root primitive for EventTable
func (t *EventTable) Primitive() tview.Primitive {
	return t.table
}

// UpdateTable adds a new event to the table and removes oldest row if we've
// reached configured maximum for events to display
func (t *EventTable) UpdateTable(evt event.Event) {
	t.count++
	evtType := string(evt.Type)

	payload, ok := evt.Payload.(discovery.DiscoveryResult)

	if !ok {
		return
	}

	status := string(payload.Status)
	ssh := string(payload.Port.Status)
	hostname := payload.Hostname
	id := payload.ID
	ip := payload.IP
	os := payload.OS
	vendor := payload.Vendor

	row := []string{evtType, hostname, ip, id, os, vendor, ssh, status}
	rowIdx := t.table.GetRowCount()

	for col, text := range row {
		cell := tview.NewTableCell(text)
		cell.SetExpansion(1)
		cell.SetAlign(tview.AlignLeft)
		cell.SetTextColor(style.ColorWhite)
		t.table.SetCell(rowIdx, col, cell)
	}

	// remove oldest row if max reached
	if t.count > t.maxEvents {
		t.table.RemoveRow(2)
	}
}
