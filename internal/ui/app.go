package ui

import (
	"context"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/event"
	"github.com/robgonnella/opi/internal/server"
	"github.com/robgonnella/opi/internal/ui/key"
)

type app struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	appCore              *core.Core
	tvApp                *tview.Application
	view                 *view
	serverUpdateChan     chan []*server.Server
	eventUpdateChan      chan *event.Event
	serverPollListenerId int
	eventListernId       int
}

func newApp(appCore *core.Core) *app {
	ctx, cancel := context.WithCancel(context.Background())

	serverUpdateChan := make(chan []*server.Server, 100)
	eventUpdateChan := make(chan *event.Event, 100)

	serverPollListenerId := appCore.RegisterServerPollListener(serverUpdateChan)
	eventListenerId := appCore.RegisterEventListener(eventUpdateChan)

	tvApp := tview.NewApplication()

	uiApp := &app{
		ctx:                  ctx,
		cancel:               cancel,
		appCore:              appCore,
		tvApp:                tvApp,
		serverPollListenerId: serverPollListenerId,
		eventListernId:       eventListenerId,
		serverUpdateChan:     serverUpdateChan,
		eventUpdateChan:      eventUpdateChan,
	}

	view := newView(uiApp)

	uiApp.view = view

	return uiApp

}

func (a *app) bindKeys() {
	a.tvApp.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case tcell.KeyCtrlC:
			a.stop()
			return evt
		case tcell.KeyEsc:
			if a.view.showingActionInput {
				a.view.hideActionInput()
				a.view.focus()
				return nil
			}
		}

		if evt.Rune() == key.RuneColon {
			returnEvt := a.view.showingActionInput
			a.view.showActionInput()

			if returnEvt {
				return evt
			}

			return nil
		}

		return evt
	})
}

func (a *app) processBackgroundServerUpdates() {
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				return
			case servers := <-a.serverUpdateChan:
				a.tvApp.QueueUpdateDraw(func() {
					sort.Slice(servers, func(i, j int) bool {
						if servers[i].Hostname == "unknown" {
							return false
						}
						if servers[j].Hostname == "unknown" {
							return true
						}
						return servers[i].Hostname < servers[j].Hostname
					})
					a.view.serverTable.UpdateTable(servers)
				})
			}
		}
	}()
}

func (a *app) processBackgroundEventUpdates() {
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				return
			case evt := <-a.eventUpdateChan:
				a.tvApp.QueueUpdateDraw(func() {
					a.view.eventTable.UpdateTable(evt)
				})
			}
		}
	}()
}

func (a *app) stop() {
	a.appCore.RemoveServerPollListener(a.serverPollListenerId)
	a.appCore.RemoveEventListener(a.eventListernId)
	a.cancel()
	a.appCore.Stop()
	a.tvApp.Stop()
	a.ctx = nil
	a.cancel = nil
}

func restart(appCore *core.Core) error {
	newUI := New(appCore)
	return newUI.Launch()
}

func (a *app) run() error {
	a.bindKeys()
	a.processBackgroundServerUpdates()
	a.processBackgroundEventUpdates()
	if err := a.appCore.BackgroundRestart(); err != nil {
		return err
	}
	return a.tvApp.SetRoot(a.view.root, true).EnableMouse(true).Run()
}
