package ui

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/go-lanscan/pkg/network"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/ui/component"
	"github.com/robgonnella/ops/internal/ui/key"
)

// viewOption provides a way to modify our view during initialization
// this is helpful when restarting the view and focusing a specific page
type viewOption func(v *view)

func withShowError(msg string) viewOption {
	return func(v *view) {
		v.showErrorModal(msg)
	}
}

// data structure for managing our entire terminal ui application
type view struct {
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
	eventManager           event.Manager
	eventUpdateChan        chan event.Event
	serverUpdateChan       chan event.Event
	errorListener          chan event.Event
	eventListenerIDs       []int
	prevFocusedName        string
	focusedName            string
	viewNames              []string
	showingSwitchViewInput bool
	log                    logger.Logger
}

// returns a new instance of view
func newView(allConfigs []*config.Config, appCore *core.Core, eventManager event.Manager) *view {
	log := logger.New()

	v := &view{
		log:          log,
		appCore:      appCore,
		eventManager: eventManager,
	}

	v.initialize(allConfigs)

	return v
}

// initializes the terminal ui application
func (v *view) initialize(
	allConfigs []*config.Config,
	options ...viewOption,
) {
	netInfo := v.appCore.NetworkInfo()

	v.viewNames = []string{"servers", "events", "context", "configure"}
	v.showingSwitchViewInput = false

	v.app = tview.NewApplication()

	v.root = tview.NewFlex().SetDirection(tview.FlexRow)
	v.pages = tview.NewPages()

	v.header = component.NewHeader(
		netInfo,
		v.appCore.Conf(),
		v.onActionSubmit,
	)
	v.serverTable = component.NewServerTable(
		netInfo.Hostname(),
		netInfo.UserIP().String(),
		v.onSSH,
	)
	v.eventTable = component.NewEventTable()
	v.contextTable = component.NewConfigContext(
		v.appCore.Conf().ID,
		allConfigs,
		v.onContextSelect,
		v.onContextDelete,
	)

	v.configureForm = component.NewConfigureForm(
		v.appCore.Conf(),
		v.onConfigureFormUpdate,
		v.onConfigureFormCreate,
		v.onDismissConfigureForm,
	)

	v.pages.AddPage("servers", v.serverTable.Primitive(), true, false)
	v.pages.AddPage("events", v.eventTable.Primitive(), true, false)
	v.pages.AddPage("configure", v.configureForm.Primitive(), true, false)
	v.pages.AddPage("context", v.contextTable.Primitive(), true, false)

	v.root.
		AddItem(v.header.Primitive(), 15, 1, false).
		AddItem(v.pages, 0, 1, true)

	v.serverUpdateChan = make(chan event.Event)
	v.eventUpdateChan = make(chan event.Event)
	v.errorListener = make(chan event.Event)
	v.eventListenerIDs = []int{}

	v.focusedName = "servers"

	v.bindKeys()
	v.registerEventListeners()
	v.processBackgroundEventUpdates()
	v.processErrorEvents()

	go func() {
		for _, o := range options {
			o(v)
		}
	}()

	v.focus(v.focusedName)
}

// change view based on result from switch view input
func (v *view) onActionSubmit(text string) {
	if text == "q" || text == "quit" {
		v.stop()
		return
	}

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

// dismisses configuration form - focuses previously focused view
func (v *view) onDismissConfigureForm() {
	v.onActionSubmit(v.prevFocusedName)
}

// updates current config with result from config form inputs
func (v *view) onConfigureFormUpdate(conf config.Config) {
	if reflect.DeepEqual(conf, v.appCore.Conf()) {
		v.onDismissConfigureForm()
		return
	}

	if err := v.appCore.UpdateConfig(conf); err != nil {
		v.eventManager.ReportError(err)
		return
	}

	confs, err := v.appCore.GetConfigs()

	if err != nil {
		v.eventManager.ReportError(err)
		return
	}

	v.contextTable.UpdateConfigs(v.appCore.Conf().ID, confs)
	v.configureForm.UpdateConfig(v.appCore.Conf())
	v.header.UpdateConfAndNetworkInfo(v.appCore.Conf(), v.appCore.NetworkInfo())

	v.onActionSubmit(v.prevFocusedName)
}

// creates a new config with results from config form inputs
func (v *view) onConfigureFormCreate(conf config.Config) {
	if err := v.appCore.CreateConfig(conf); err != nil {
		v.eventManager.ReportError(err)
		return
	}

	confs, err := v.appCore.GetConfigs()

	if err != nil {
		v.eventManager.ReportError(err)
		return
	}

	v.contextTable.UpdateConfigs(v.appCore.Conf().ID, confs)
	v.configureForm.UpdateConfig(v.appCore.Conf())
	v.header.UpdateConfAndNetworkInfo(v.appCore.Conf(), v.appCore.NetworkInfo())

	v.focus("context")
}

// selects a new context for network scanning
func (v *view) onContextSelect(id string) {
	if id == v.appCore.Conf().ID {
		v.focus("servers")
		return
	}

	if err := v.appCore.SetConfig(id); err != nil {
		v.eventManager.ReportError(err)
		return
	}

	confs, err := v.appCore.GetConfigs()

	if err != nil {
		v.eventManager.ReportError(err)
		return
	}

	v.contextTable.UpdateConfigs(v.appCore.Conf().ID, confs)
	v.configureForm.UpdateConfig(v.appCore.Conf())
	v.header.UpdateConfAndNetworkInfo(v.appCore.Conf(), v.appCore.NetworkInfo())

	v.focus("servers")
}

// dismisses confirmation modal when deleting a context
func (v *view) dismissContextDelete() {
	v.contextToDelete = ""
	v.app.SetRoot(v.root, true)
}

// shows confirmation modal when attempting to delete a context
func (v *view) onContextDelete(name string, id string) {
	v.contextToDelete = id
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

// deletes a network scanning context (configuration)
func (v *view) deleteContext() {
	if v.contextToDelete == "" {
		return
	}

	defer func() {
		v.contextToDelete = ""
	}()

	if err := v.appCore.DeleteConfig(v.contextToDelete); err != nil {
		v.eventManager.ReportError(err)
		return
	}

	currentConfig := v.appCore.Conf().ID

	if v.contextToDelete == currentConfig {
		v.showErrorModal("cannot delete current active context")
	} else {
		confs, err := v.appCore.GetConfigs()

		if err != nil {
			v.eventManager.ReportError(err)
			return
		}

		v.contextTable.UpdateConfigs(currentConfig, confs)
		v.dismissContextDelete()
	}
}

// displays an error modal
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

// dismisses an error modal
func (v *view) dismissErrorModal() {
	v.app.SetRoot(v.root, true)
}

func (v *view) showFatalErrorModal(errMsg string) {
	dismiss := func() {
		v.dismissFatalErrorModal()
	}

	buttons := []component.ModalButton{
		{
			Label:   "Exit",
			OnClick: dismiss,
		},
	}

	errorModal := component.NewModal(
		errMsg,
		buttons,
	)

	v.app.SetRoot(errorModal.Primitive(), false)
}

func (v *view) dismissFatalErrorModal() {
	v.restart()
}

// binds global key handlers
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

// focuses a given view by name and updates the legend to display the correct
// key mappings for that view
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
			v.showErrorModal("Failed to retrieve configurations")
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

// Attempts to ssh to the given server using current config's ssh properties.
// This requires stopping the terminal ui application so we can return
// to the normal terminal screen. We ensure our terminal app is restarted
// once the ssh command finishes aka when the user exists the ssh tunnel.
func (v *view) onSSH(ip string) {
	v.stop()

	conf := v.appCore.Conf()
	user := conf.SSH.User
	identity := conf.SSH.Identity
	port := conf.SSH.Port

	for _, o := range conf.SSH.Overrides {
		if o.Target == ip {
			if o.User != "" {
				user = o.User
			}

			if o.Identity != "" {
				identity = o.Identity
			}

			if o.Port != "" {
				port = o.Port
			}
		}
	}

	cmd := exec.Command(
		"ssh",
		"-i",
		identity,
		"-p",
		port,
		"-l",
		user,
		ip,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		v.restart(withShowError("failed to ssh to " + ip + ": " + err.Error()))
		return
	}

	v.restart()
}

// maps names to primitives for focusing
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

// completely stops the tui app and all backend processes.
// this requires a full restart including re-instantiation of entire backend
func (v *view) stop() {
	for _, id := range v.eventListenerIDs {
		v.eventManager.RemoveListener(id)
	}
	v.eventListenerIDs = []int{}
	v.app.Stop()
	restoreStdout()
	go func() {
		if err := v.appCore.Stop(); err != nil {
			v.eventManager.ReportFatalError(err)
		}
	}()
}

// restarts the entire application including re-instantiation of entire backend
func (v *view) restart(options ...viewOption) {
	v.stop()

	restoreStdout()

	conf := v.appCore.Conf()

	netInfo, err := network.NewDefaultNetwork()

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to get default network info")
	}

	appCore, err := core.CreateNewAppCore(netInfo, v.eventManager, false)

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to restart app core")
	}

	if err := appCore.SetConfig(conf.ID); err != nil {
		v.log.Fatal().Err(err).Msg("failed to set config on restart")
	}

	v.appCore = appCore

	allConfigs, err := v.appCore.GetConfigs()

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to retrieve configs")
	}

	maskStdout()

	v.initialize(allConfigs, options...)

	if err := v.run(); err != nil {
		restoreStdout()
		v.log.Fatal().Err(err).Msg("failed to restart view")
	}
}

// registers event listeners
func (v *view) registerEventListeners() {
	v.eventListenerIDs = append(
		v.eventListenerIDs,
		v.eventManager.RegisterListener(discovery.ArpUpdateEvent, v.eventUpdateChan),
		v.eventManager.RegisterListener(discovery.SynUpdateEvent, v.eventUpdateChan),
	)
	v.eventListenerIDs = append(
		v.eventListenerIDs,
		v.eventManager.RegisterListener(discovery.ArpUpdateEvent, v.serverUpdateChan),
		v.eventManager.RegisterListener(discovery.SynUpdateEvent, v.serverUpdateChan),
	)
	v.eventListenerIDs = append(
		v.eventListenerIDs,
		v.eventManager.RegisterListener(event.ErrorEventType, v.errorListener),
		v.eventManager.RegisterListener(event.FatalErrorEventType, v.errorListener),
	)
}

// handle incoming server events
func (v *view) processBackgroundEventUpdates() {
	go func() {
		for {
			select {
			case evt, ok := <-v.eventUpdateChan:
				if !ok {
					return
				}
				v.app.QueueUpdateDraw(func() {
					v.eventTable.UpdateTable(evt)
				})
			case evt, ok := <-v.serverUpdateChan:
				if !ok {
					return
				}
				v.app.QueueUpdateDraw(func() {
					v.serverTable.UpdateTable(evt)
				})
			}
		}
	}()
}

// handle incoming error events
func (v *view) processErrorEvents() {
	go func() {
		for evt := range v.errorListener {
			switch evt.Type {
			case event.FatalErrorEventType:
				v.showFatalErrorModal(evt.Payload.(error).Error())
				v.app.Draw()
			case event.ErrorEventType:
				v.showErrorModal(evt.Payload.(error).Error())
				v.app.Draw()
			}
		}
	}()
}

// start application
func (v *view) run() error {
	v.appCore.StartDaemon()
	return v.app.SetRoot(v.root, true).EnableMouse(true).Run()
}
