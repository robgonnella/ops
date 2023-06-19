package ui

import (
	"context"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/event"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/server"
	"github.com/robgonnella/opi/internal/ui/component"
	"github.com/robgonnella/opi/internal/ui/key"
)

func restart() error {
	newUI := New()
	return newUI.Launch()
}

type view struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	app                  *tview.Application
	root                 *tview.Flex
	pages                *tview.Pages
	extraLegend          map[string]string
	header               *component.Header
	serverTable          *component.ServerTable
	eventTable           *component.EventTable
	configureForm        *component.ConfigureForm
	contextTable         *component.ConfigContext
	appCore              *core.Core
	serverUpdateChan     chan []*server.Server
	eventUpdateChan      chan *event.Event
	serverPollListenerId int
	eventListernId       int
	focused              tview.Primitive
	focusedName          string
	viewNames            []string
	logger               logger.Logger
}

func newView(appCore *core.Core) *view {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())

	app := tview.NewApplication()

	v := &view{
		ctx:       ctx,
		cancel:    cancel,
		appCore:   appCore,
		app:       app,
		viewNames: []string{"servers", "events", "context", "configure"},
		logger:    log,
	}

	allConfigs, _ := v.appCore.GetConfigs()

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	pages := tview.NewPages()

	header := component.NewHeader(v.onActionSubmit)
	serverTable := component.NewServerTable(v.onSSH)
	eventTable := component.NewEventTable()
	contextTable := component.NewConfigContext(
		v.appCore.Conf().Name,
		allConfigs,
		v.onContextSelect,
		v.onContextDelete,
	)

	configureForm := component.NewConfigureForm(
		v.appCore.Conf(),
		v.onConfigureFormSubmit,
	)

	pages.AddPage("servers", serverTable.Primitive(), true, false)
	pages.AddPage("events", eventTable.Primitive(), true, false)
	pages.AddPage("configure", configureForm.Primitive(), true, false)
	pages.AddPage("context", contextTable.Primitive(), true, false)

	root.
		AddItem(header.Primitive(), 12, 1, false).
		AddItem(pages, 0, 1, true)

	serverUpdateChan := make(chan []*server.Server, 100)
	eventUpdateChan := make(chan *event.Event, 100)

	serverPollListenerId := appCore.RegisterServerPollListener(serverUpdateChan)
	eventListenerId := appCore.RegisterEventListener(eventUpdateChan)

	v.root = root
	v.pages = pages
	v.header = header
	v.serverTable = serverTable
	v.eventTable = eventTable
	v.configureForm = configureForm
	v.contextTable = contextTable
	v.serverUpdateChan = serverUpdateChan
	v.eventUpdateChan = eventUpdateChan
	v.serverPollListenerId = serverPollListenerId
	v.eventListernId = eventListenerId

	v.focused = serverTable.Primitive()
	v.focusedName = "servers"

	v.focus()

	return v
}

func (v *view) onActionSubmit(text string) {
	for _, name := range v.viewNames {
		if strings.HasPrefix(name, text) {
			v.focusedName = name

			switch name {
			case "servers":
				v.focused = v.serverTable.Primitive()
			case "events":
				v.focused = v.eventTable.Primitive()
			case "configure":
				v.focused = v.configureForm.Primitive()
			case "context":
				v.focused = v.contextTable.Primitive()
			}

			break
		}
	}

	v.header.HideSwitchViewInput()
	v.focus()
}

func (v *view) onConfigureFormSubmit(conf config.Config) {
	if err := v.appCore.UpdateConfig(conf); err != nil {
		v.logger.Error().Err(err).Msg("failed to write config file")
	}

	v.stop()
	restart()
}

func (v *view) onContextSelect(name string) {
	if err := v.appCore.SetConfig(name); err != nil {
		v.logger.Error().Err(err).Msg("failed to set new config")
		return
	}

	v.stop()
	restart()
}

func (v *view) onContextDelete(name string) {
	if err := v.appCore.DeleteConfig(name); err != nil {
		v.logger.Error().Err(err).Msg("failed to delete config")
		return
	}

	v.stop()
	restart()
}

func (v *view) bindKeys() {
	v.app.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case key.KeyCtrlC:
			v.stop()
			return evt
		case key.KeyEsc:
			if v.header.IsShowingSwitchViewInput() {
				v.header.HideSwitchViewInput()
				v.focus()
				return nil
			}
		}

		if evt.Rune() == key.RuneColon {
			if v.header.IsShowingSwitchViewInput() {
				return evt
			}

			v.header.ShowSwitchViewInput()
			v.app.SetFocus(v.header.SwitchViewInput().Primitive())

			return nil
		}

		return evt
	})
}

func (v *view) focus() {
	extraLegend := map[string]string{}

	switch v.focusedName {
	case "servers":
		extraLegend["ctrl+s"] = "ssh to selected machine"
	case "context":
		extraLegend["ctrl+d"] = "delete context"
	}

	v.header.RemoveExtraLegend()
	v.header.ShowExtraLegend(extraLegend)

	v.pages.SwitchToPage(v.focusedName)
	v.app.SetFocus(v.focused)
}

func (v *view) onSSH(ip string) {
	v.stop()

	defer func() {
		if err := restart(); err != nil {
			v.logger.Error().Err(err).Msg("error restarting ui")
		}
	}()

	conf := v.appCore.Conf()
	user := conf.SSH.User
	identity := conf.SSH.Identity

	for _, o := range conf.SSH.Overrides {
		if o.Target == ip {
			if o.User != "" {
				user = o.User
			}

			if o.Identity != "" {
				identity = o.Identity
			}
		}
	}

	cmd := exec.Command("ssh", "-i", identity, user+"@"+ip)

	os.Stdout = originalStdout
	os.Stderr = originalStderr

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
}

func (v *view) stop() {
	v.appCore.RemoveServerPollListener(v.serverPollListenerId)
	v.appCore.RemoveEventListener(v.eventListernId)
	v.cancel()
	v.appCore.Stop()
	v.app.Stop()
	v.ctx = nil
	v.cancel = nil
}

func (v *view) processBackgroundServerUpdates() {
	go func() {
		for {
			select {
			case <-v.ctx.Done():
				return
			case servers := <-v.serverUpdateChan:
				v.app.QueueUpdateDraw(func() {
					sort.Slice(servers, func(i, j int) bool {
						if servers[i].Hostname == "unknown" {
							return false
						}
						if servers[j].Hostname == "unknown" {
							return true
						}
						return servers[i].Hostname < servers[j].Hostname
					})
					v.serverTable.UpdateTable(servers)
				})
			}
		}
	}()
}

func (v *view) processBackgroundEventUpdates() {
	go func() {
		for {
			select {
			case <-v.ctx.Done():
				return
			case evt := <-v.eventUpdateChan:
				v.app.QueueUpdateDraw(func() {
					v.eventTable.UpdateTable(evt)
				})
			}
		}
	}()
}

func (v *view) run() error {
	v.bindKeys()
	v.processBackgroundServerUpdates()
	v.processBackgroundEventUpdates()
	if err := v.appCore.BackgroundRestart(); err != nil {
		return err
	}
	return v.app.SetRoot(v.root, true).EnableMouse(true).Run()
}
