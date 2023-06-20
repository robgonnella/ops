package component

import (
	"context"
	"strconv"

	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/ui/style"
)

type EventTable struct {
	ctx           context.Context
	cancel        context.CancelFunc
	table         *tview.Table
	columnHeaders []string
	count         uint
	maxEvents     uint
}

func NewEventTable() *EventTable {
	columnHeaders := []string{
		"NO",
		"EVENT TYPE",
		"HOSTNAME",
		"IP",
		"ID",
		"OS",
		"SSH",
		"STATUS",
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EventTable{
		ctx:           ctx,
		cancel:        cancel,
		table:         createTable("events", columnHeaders),
		columnHeaders: columnHeaders,
		count:         0,
		maxEvents:     50,
	}
}

func (t *EventTable) Primitive() tview.Primitive {
	return t.table
}

func (t *EventTable) UpdateTable(evt *event.Event) {
	t.count++
	evtType := string(evt.Type)
	payload := evt.Payload.(*server.Server)

	status := string(payload.Status)
	ssh := string(payload.SshStatus)
	hostname := payload.Hostname
	id := payload.ID
	ip := payload.IP
	os := payload.OS

	countStr := strconv.Itoa(int(t.count))

	row := []string{countStr, evtType, hostname, ip, id, os, ssh, status}
	rowIdx := t.table.GetRowCount()

	for col, text := range row {
		cell := tview.NewTableCell(text)
		cell.SetExpansion(1)
		cell.SetAlign(tview.AlignLeft)
		cell.SetTextColor(style.ColorWhite)
		t.table.SetCell(rowIdx, col, cell)
	}

	if t.count > t.maxEvents {
		t.table.RemoveRow(2)
	}
}
