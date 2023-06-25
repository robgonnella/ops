package ui

import (
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
	"github.com/robgonnella/ops/internal/util"
)

// viewOption provides a way to modify our view during initialization
// this is helpful when restarting the view and focusing a specific page
type viewOption func(v *view)

// withFocusedView sets the option to focus a specific view
// during initialization
func withFocusedView(name string) viewOption {
	return func(v *view) {
		v.focusedName = name
	}
}

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
	contextToDelete        int
	appCore                *core.Core
	serverUpdateChan       chan []*server.Server
	eventUpdateChan        chan *event.Event
	serverPollListenerID   int
	eventListenerID        int
	prevFocusedName        string
	focusedName            string
	viewNames              []string
	showingSwitchViewInput bool
	log                    logger.Logger
}

// returns a new instance of view
func newView(userHostname, userIP string, allConfigs []*config.Config, appCore *core.Core) *view {
	log := logger.New()

	v := &view{
		log:     log,
		appCore: appCore,
	}

	v.initialize(userHostname, userIP, allConfigs)

	return v
}

// initializes the terminal ui application
func (v *view) initialize(
	userHostname,
	userIP string,
	allConfigs []*config.Config,
	options ...viewOption,
) {
	v.viewNames = []string{"servers", "events", "context", "configure"}
	v.showingSwitchViewInput = false

	v.app = tview.NewApplication()

	v.root = tview.NewFlex().SetDirection(tview.FlexRow)
	v.pages = tview.NewPages()

	v.header = component.NewHeader(
		userIP,
		v.appCore.Conf().Targets,
		v.onActionSubmit,
	)
	v.serverTable = component.NewServerTable(userHostname, userIP, v.onSSH)
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
		AddItem(v.header.Primitive(), 12, 1, false).
		AddItem(v.pages, 0, 1, true)

	v.serverUpdateChan = make(chan []*server.Server)
	v.eventUpdateChan = make(chan *event.Event)
	v.focusedName = "servers"

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
		v.log.Error().Err(err).Msg("failed to save config")
		v.showErrorModal("Failed to save config")
		return
	}

	v.restart(withFocusedView("context"))
}

// creates a new config with results from config form inputs
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

	v.contextTable.UpdateConfigs(v.appCore.Conf().ID, confs)

	v.focus("context")
}

// selects a new context for network scanning
func (v *view) onContextSelect(id int) {
	if err := v.appCore.SetConfig(id); err != nil {
		v.log.Error().Err(err).Msg("failed to set new context")
		v.showErrorModal("Failed to set new context")
		return
	}

	v.restart()
}

// dismisses confirmation modal when deleting a context
func (v *view) dismissContextDelete() {
	v.contextToDelete = 0
	v.app.SetRoot(v.root, true)
}

// shows confirmation modal when attempting to delete a context
func (v *view) onContextDelete(name string, id int) {
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
	if v.contextToDelete == 0 {
		return
	}

	defer func() {
		v.contextToDelete = 0
	}()

	if err := v.appCore.DeleteConfig(v.contextToDelete); err != nil {
		v.log.Error().Err(err).Msg("failed to delete config")
		v.showErrorModal("Failed to delete context")
		return
	}

	currentConfig := v.appCore.Conf().ID

	if v.contextToDelete == currentConfig {
		// deleted current context - restart app
		v.restart(withFocusedView("context"))
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

// Attempts to ssh to the given server using current config's ssh properties.
// This requires stopping the terminal ui application so we can return
// to the normal terminal screen. We ensure our terminal app is restarted
// once the ssh command finishes aka when the user exists the ssh tunnel.
func (v *view) onSSH(ip string) {
	v.stop()

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

	restoreStdout()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		v.restart(
			withShowError("failed to ssh to " + ip + ": " + err.Error()),
		)
		return
	}

	v.restart()
}

// handle incoming server results from database polling
func (v *view) processBackgroundServerUpdates() {
	go func() {
		for {
			servers, ok := <-v.serverUpdateChan
			if !ok {
				return
			}
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
	}()
}

// handle incoming database events
func (v *view) processBackgroundEventUpdates() {
	go func() {
		for {
			evt, ok := <-v.eventUpdateChan
			if !ok {
				return
			}
			v.app.QueueUpdateDraw(func() {
				v.eventTable.UpdateTable(evt)
			})
		}
	}()
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
	v.appCore.RemoveServerPollListener(v.serverPollListenerID)
	v.appCore.RemoveEventListener(v.eventListenerID)
	v.serverPollListenerID = 0
	v.eventListenerID = 0
	v.appCore.Stop()
	v.app.Stop()
}

// restarts the entire application including re-instantiation of entire backend
func (v *view) restart(options ...viewOption) {
	v.stop()

	restoreStdout()

	userIP, cidr, err := util.GetNetworkInfo()

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to get default network info")
	}

	hostname, err := util.Hostname()

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to get hostname for current device")
	}

	appCore, err := util.CreateNewAppCore(*cidr)

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to restart app core")
	}

	v.appCore = appCore

	allConfigs, err := v.appCore.GetConfigs()

	if err != nil {
		v.log.Fatal().Err(err).Msg("failed to retrieve configs")
	}

	maskStdout()

	v.initialize(*hostname, *userIP, allConfigs, options...)

	if err := v.run(); err != nil {
		restoreStdout()
		v.log.Fatal().Err(err).Msg("failed to restart view")
	}
}

// sets up global key handlers, registers event listeners, sets up processing
// for channel updates, starts backend processes, and starts terminal ui
func (v *view) run() error {
	v.bindKeys()
	v.serverPollListenerID = v.appCore.RegisterServerPollListener(
		v.serverUpdateChan,
	)
	v.eventListenerID = v.appCore.RegisterEventListener(v.eventUpdateChan)
	v.processBackgroundServerUpdates()
	v.processBackgroundEventUpdates()
	v.appCore.StartDaemon()
	return v.app.SetRoot(v.root, true).EnableMouse(true).Run()
}
