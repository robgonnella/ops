package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/ui/component"
	"github.com/robgonnella/ops/internal/ui/key"
)

func restart(options ...ViewOption) error {
	newUI := NewUI()
	return newUI.Launch(options...)
}

type ViewOption = func(v *view)

func WithInitialFocus(name string) func(v *view) {
	return func(v *view) {
		v.focusedName = name
		v.focus(v.focusedName)
	}
}

type view struct {
	ctx                    context.Context
	cancel                 context.CancelFunc
	app                    *tview.Application
	root                   *tview.Flex
	pages                  *tview.Pages
	header                 *component.Header
	serverTable            *component.ServerTable
	eventTable             *component.EventTable
	configureForm          *component.ConfigureForm
	contextTable           *component.ConfigContext
	contextToDelete        string
	appCore                *core.Core
	serverUpdateChan       chan []*server.Server
	eventUpdateChan        chan *event.Event
	serverPollListenerId   int
	eventListernId         int
	prevFocusedName        string
	focusedName            string
	viewNames              []string
	showingSwitchViewInput bool
	log                    logger.Logger
}

func newView(userIP string, appCore *core.Core) *view {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())

	app := tview.NewApplication()

	v := &view{
		ctx:                    ctx,
		cancel:                 cancel,
		appCore:                appCore,
		app:                    app,
		viewNames:              []string{"servers", "events", "context", "configure"},
		showingSwitchViewInput: false,
		log:                    log,
	}

	allConfigs, _ := v.appCore.GetConfigs()

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	pages := tview.NewPages()

	header := component.NewHeader(
		userIP,
		appCore.Conf().Targets,
		v.onActionSubmit,
	)
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
		v.onConfigureFormUpdate,
		v.onConfigureFormCreate,
		v.onDismissConfigureForm,
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
	v.contextToDelete = ""
	v.serverUpdateChan = serverUpdateChan
	v.eventUpdateChan = eventUpdateChan
	v.serverPollListenerId = serverPollListenerId
	v.eventListernId = eventListenerId
	v.focusedName = "servers"

	v.focus("servers")

	return v
}

func (v *view) onActionSubmit(text string) {
	focusedName := ""

	for _, name := range v.viewNames {
		if strings.HasPrefix(name, text) {
			focusedName = name
			break
		}
	}

	v.focus(focusedName)
	v.showingSwitchViewInput = false
}

func (v *view) onDismissConfigureForm() {
	v.onActionSubmit(v.prevFocusedName)
}

func (v *view) onConfigureFormUpdate(conf config.Config) {
	if reflect.DeepEqual(conf, v.appCore.Conf()) {
		v.onDismissConfigureForm()
		return
	}

	if err := v.appCore.UpdateConfig(conf); err != nil {
		v.log.Error().Err(err).Msg("failed to save config")
		v.showErrorModal("Failed to save config")
		return
	}

	v.stop()
	restart(WithInitialFocus("context"))
}

func (v *view) onConfigureFormCreate(conf config.Config) {
	if err := v.appCore.CreateConfig(conf); err != nil {
		v.log.Error().Err(err).Msg("failed to create config")
		v.showErrorModal("Failed to create config")
		return
	}

	confs, err := v.appCore.GetConfigs()

	if err != nil {
		v.log.Error().Err(err).Msg("failed to get configs")
		v.showErrorModal("Failed to retrieve configs")
		return
	}

	v.contextTable.UpdateConfigs(v.appCore.Conf().Name, confs)

	v.focus("context")
}

func (v *view) onContextSelect(name string) {
	if err := v.appCore.SetConfig(name); err != nil {
		v.log.Error().Err(err).Msg("failed to set new context")
		v.showErrorModal("Failed to set new context")
		return
	}

	v.stop()
	restart(WithInitialFocus("servers"))
}

func (v *view) dismissContextDelete() {
	v.contextToDelete = ""
	v.app.SetRoot(v.root, true)
}

func (v *view) onContextDelete(name string) {
	v.contextToDelete = name
	buttons := []component.ModalButton{
		{
			Label:   "OK",
			OnClick: v.deleteContext,
		},
		{
			Label:   "Dismiss",
			OnClick: v.dismissContextDelete,
		},
	}
	contextDeleteConfirm := component.NewModal(
		fmt.Sprintf("Delete %s configuration?", name),
		buttons,
	)
	v.app.SetRoot(contextDeleteConfirm.Primitive(), false)
}

func (v *view) deleteContext() {
	if v.contextToDelete == "" {
		return
	}

	defer func() {
		v.contextToDelete = ""
	}()

	if err := v.appCore.DeleteConfig(v.contextToDelete); err != nil {
		v.log.Error().Err(err).Msg("failed to delete config")
		v.showErrorModal("Failed to delete context")
		return
	}

	currentConfig := v.appCore.Conf().Name

	if v.contextToDelete == currentConfig {
		// deleted current context - restart app
		v.stop()
		restart(WithInitialFocus("context"))
	} else {
		confs, err := v.appCore.GetConfigs()

		if err != nil {
			v.log.Error().Err(err).Msg("failed to get configs")
			v.showErrorModal("Failed to retrieve configs")
			return
		}

		v.contextTable.UpdateConfigs(currentConfig, confs)
		v.dismissContextDelete()
	}
}

func (v *view) showErrorModal(message string) {
	buttons := []component.ModalButton{
		{
			Label:   "Dismiss",
			OnClick: v.dismissErrorModal,
		},
	}
	errorModal := component.NewModal(
		message,
		buttons,
	)
	v.app.SetRoot(errorModal.Primitive(), false)
}

func (v *view) dismissErrorModal() {
	v.app.SetRoot(v.root, true)
}

func (v *view) bindKeys() {
	v.app.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case key.KeyCtrlC:
			v.stop()
			return evt
		case key.KeyEsc:
			if v.showingSwitchViewInput {
				v.focus(v.focusedName)
				v.showingSwitchViewInput = false
				return nil
			}
		}

		if evt.Rune() == key.RuneColon {
			if v.showingSwitchViewInput {
				return evt
			}

			v.app.SetFocus(v.header.SwitchViewInput().Primitive())
			v.showingSwitchViewInput = true

			return nil
		}

		return evt
	})
}

func (v *view) focus(name string) {
	p := v.getFocusNamePrimitive(name)

	if p == nil {
		return
	}

	switch name {
	case "servers":
		v.header.RemoveAllExtraLegendKeys()
		v.header.AddLegendKey("ctrl+s", "ssh to selected machine")
	case "context":
		confs, err := v.appCore.GetConfigs()

		if err != nil {
			v.log.Error().Err(err).Msg("")
			v.showErrorModal("Failed to retrieve configurations from database")
			return
		}

		if len(confs) > 1 {
			v.header.RemoveAllExtraLegendKeys()
			v.header.AddLegendKey("ctrl+d", "delete context")
			v.header.AddLegendKey("enter", "select new context")
		}
	default:
		v.header.RemoveAllExtraLegendKeys()
	}

	v.prevFocusedName = v.focusedName
	v.focusedName = name

	v.pages.SwitchToPage(name)
	v.app.SetFocus(p)
}

func (v *view) onSSH(ip string) {
	v.stop()

	defer func() {
		if err := restart(); err != nil {
			v.log.Error().Err(err).Msg("error restarting ui")
			os.Exit(1)
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

func (v *view) getFocusNamePrimitive(name string) tview.Primitive {
	switch name {
	case "servers":
		return v.serverTable.Primitive()
	case "events":
		return v.eventTable.Primitive()
	case "context":
		return v.contextTable.Primitive()
	case "configure":
		return v.configureForm.Primitive()
	default:
		return nil
	}
}

func (v *view) run() error {
	v.bindKeys()
	v.processBackgroundServerUpdates()
	v.processBackgroundEventUpdates()
	v.appCore.StartDaemon()
	return v.app.SetRoot(v.root, true).EnableMouse(true).Run()
}
