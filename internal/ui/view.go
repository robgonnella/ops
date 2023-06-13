package ui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/ui/component"
	"github.com/robgonnella/opi/internal/ui/style"
)

type view struct {
	app                *app
	root               *tview.Frame
	container          *tview.Flex
	pages              *tview.Pages
	legend             *component.Legend
	serverTable        *component.ServerTable
	eventTable         *component.EventTable
	actionInput        *component.ActionInput
	configureForm      *component.ConfigureForm
	contextTable       *component.ConfigContext
	focused            tview.Primitive
	focusedName        string
	showingActionInput bool
	viewNames          []string
	logger             logger.Logger
}

func newView(app *app) *view {
	log := logger.New()

	const heading1 = " ██████╗ ██████╗ ██╗"
	const heading2 = "██╔═══██╗██╔══██╗██║"
	const heading3 = "██║   ██║██████╔╝██║"
	const heading4 = "██║   ██║██╔═══╝ ██║"
	const heading5 = "╚██████╔╝██║     ██║"
	const heading6 = " ╚═════╝ ╚═╝     ╚═╝"

	v := &view{
		app:                app,
		showingActionInput: false,
		viewNames:          []string{"servers", "events", "configure", "context"},
		logger:             log,
	}

	allConfigs, _ := v.app.appCore.GetConfigs()

	container := tview.NewFlex().SetDirection(tview.FlexRow)
	pages := tview.NewPages()

	legend := component.NewLegend(
		strings.Join(v.app.appCore.Conf().Targets,
			",",
		),
		v.viewNames,
	)
	actionInput := component.NewActionInput(v.onActionSubmit)
	serverTable := component.NewServerTable(v.onSSH)
	eventTable := component.NewEventTable()
	contextTable := component.NewConfigContext(
		v.app.appCore.Conf().Name,
		allConfigs,
		v.onContextSelect,
		v.onContextDelete,
	)

	configureForm := component.NewConfigureForm(
		v.app.appCore.Conf(),
		v.onConfigureFormSubmit,
	)

	pages.AddPage("servers", serverTable.Primitive(), true, false)
	pages.AddPage("events", eventTable.Primitive(), true, false)
	pages.AddPage("configure", configureForm.Primitive(), true, false)
	pages.AddPage("context", contextTable.Primitive(), true, false)

	container.
		AddItem(legend.Primitive(), 4, 1, false).
		AddItem(pages, 0, 1, true)

	frame := tview.NewFrame(container).
		AddText(heading1, true, tview.AlignCenter, style.ColorPurple).
		AddText(heading2, true, tview.AlignCenter, style.ColorPurple).
		AddText(heading3, true, tview.AlignCenter, style.ColorPurple).
		AddText(heading4, true, tview.AlignCenter, style.ColorPurple).
		AddText(heading5, true, tview.AlignCenter, style.ColorPurple).
		AddText(heading6, true, tview.AlignCenter, style.ColorPurple)

	v.root = frame
	v.container = container
	v.pages = pages
	v.legend = legend
	v.serverTable = serverTable
	v.eventTable = eventTable
	v.actionInput = actionInput
	v.configureForm = configureForm
	v.contextTable = contextTable

	v.focused = serverTable.Primitive()
	v.focusedName = "servers"

	v.focus()

	return v
}

func (v *view) showActionInput() {
	if v.showingActionInput {
		return
	}

	newContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	newContainer.
		AddItem(v.legend.Primitive(), 4, 1, false).
		AddItem(v.actionInput.Primitive(), 3, 1, false).
		AddItem(v.pages, 0, 1, true)

	v.container = newContainer
	v.root.SetPrimitive(v.container)
	v.app.tvApp.SetFocus(v.actionInput.Primitive())
	v.showingActionInput = !v.showingActionInput
}

func (v *view) hideActionInput() {
	if !v.showingActionInput {
		return
	}

	v.container.RemoveItem(v.actionInput.Primitive())
	v.showingActionInput = !v.showingActionInput
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

	v.focus()
	v.hideActionInput()
}

func (v *view) onConfigureFormSubmit(conf config.Config) {
	if err := v.app.appCore.UpdateConfig(conf); err != nil {
		v.logger.Error().Err(err).Msg("failed to write config file")
	}

	v.app.stop()
	restart()
}

func (v *view) onContextSelect(name string) {
	if err := v.app.appCore.SetConfig(name); err != nil {
		v.logger.Error().Err(err).Msg("failed to set new config")
		return
	}

	v.app.stop()
	restart()
}

func (v *view) onContextDelete(name string) {
	if err := v.app.appCore.DeleteConfig(name); err != nil {
		v.logger.Error().Err(err).Msg("failed to delete config")
		return
	}

	v.app.stop()
	restart()
}

func (v *view) focus() {
	v.pages.SwitchToPage(v.focusedName)
	v.app.tvApp.SetFocus(v.focused)
}

func (v *view) onSSH(ip string) {
	v.app.stop()

	defer func() {
		if err := restart(); err != nil {
			v.logger.Error().Err(err).Msg("error restarting ui")
		}
	}()

	conf := v.app.appCore.Conf()
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
