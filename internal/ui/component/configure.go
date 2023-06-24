package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/ui/style"
)

// ConfigureForm component for updating and creating configurations
type ConfigureForm struct {
	root              *tview.Form
	configName        *tview.InputField
	sshUserInput      *tview.InputField
	sshIdentityInput  *tview.InputField
	cidrInput         *tview.InputField
	overrides         []map[string]*tview.InputField
	conf              config.Config
	onUpdate          func(conf config.Config)
	onCreate          func(conf config.Config)
	onDismiss         func()
	creatingNewConfig bool
}

// adds blank form inputs and sets styling
func addBlankFormItems(
	form *tview.Form,
	confName string,
) (*tview.InputField, *tview.InputField, *tview.InputField, *tview.InputField) {
	configName := tview.NewInputField()
	configName.SetLabel("Config Name: ")

	sshUserInput := tview.NewInputField()
	sshUserInput.SetLabel("SSH User: ")

	sshIdentityInput := tview.NewInputField()
	sshIdentityInput.SetLabel("SSH Identity: ")

	cidrInput := tview.NewInputField()
	cidrInput.SetLabel("Comma Separated CIDRs or IPs: ")

	form.AddFormItem(configName)
	form.AddFormItem(cidrInput)
	form.AddFormItem(sshUserInput)
	form.AddFormItem(sshIdentityInput)

	form.SetTitle(confName + " Configuration")
	form.SetBorder(true)
	form.SetBorderColor(style.ColorPurple)
	form.SetFieldBackgroundColor(tcell.ColorDefault)
	form.SetButtonBackgroundColor(style.ColorLightGreen)
	form.SetLabelColor(style.ColorOrange)
	form.SetButtonTextColor(style.ColorBlack)
	form.SetButtonActivatedStyle(
		style.StyleDefault.Background(style.ColorLightGreen),
	)

	return configName, sshUserInput, sshIdentityInput, cidrInput
}

// every time the add ssh override button is clicked we add three new inputs
func createOverrideInputs() (*tview.InputField, *tview.InputField, *tview.InputField) {
	overrideTarget := tview.NewInputField()
	overrideTarget.SetLabel("Override Target: ")

	overrideSSHUser := tview.NewInputField()
	overrideSSHUser.SetLabel("Override SSH User: ")

	overrideSSHIdentity := tview.NewInputField()
	overrideSSHIdentity.SetLabel("Override SSH Identity: ")

	return overrideTarget, overrideSSHUser, overrideSSHIdentity
}

// NewConfigureForm returns a new instance of ConfigureForm
func NewConfigureForm(
	conf config.Config,
	onUpdate func(conf config.Config),
	onCreate func(conf config.Config),
	onDismiss func(),
) *ConfigureForm {
	form := tview.NewForm()

	configName, sshUserInput, sshIdentityInput, cidrInput := addBlankFormItems(
		form,
		conf.Name,
	)

	return &ConfigureForm{
		root:              form,
		configName:        configName,
		sshUserInput:      sshUserInput,
		sshIdentityInput:  sshIdentityInput,
		cidrInput:         cidrInput,
		overrides:         []map[string]*tview.InputField{},
		conf:              conf,
		onUpdate:          onUpdate,
		onCreate:          onCreate,
		onDismiss:         onDismiss,
		creatingNewConfig: false,
	}
}

// Primitive return the root primitive for ConfigureForm
func (f *ConfigureForm) Primitive() tview.Primitive {
	f.render()
	return f.root
}

// preloads all info based on current active configuration (context)
func (f *ConfigureForm) render() {
	f.root.Clear(true)
	f.overrides = []map[string]*tview.InputField{}

	f.configName, f.sshUserInput, f.sshIdentityInput, f.cidrInput =
		addBlankFormItems(f.root, f.conf.Name)

	networkTargets := strings.Join(f.conf.Targets, ",")

	f.configName.SetText(f.conf.Name)
	f.sshUserInput.SetText(f.conf.SSH.User)
	f.sshIdentityInput.SetText(f.conf.SSH.Identity)
	f.cidrInput.SetText(networkTargets)

	for _, o := range f.conf.SSH.Overrides {
		target, user, identity := createOverrideInputs()

		f.overrides = append(f.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		target.SetText(o.Target)
		user.SetText(o.User)
		identity.SetText(o.Identity)

		f.root.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	}

	f.addFormButtons()
}

// adds buttons to form
func (f *ConfigureForm) addFormButtons() {
	f.root.AddButton("Cancel", func() {
		if f.creatingNewConfig {
			f.creatingNewConfig = false
			f.render()
			return
		}

		f.onDismiss()
	})

	f.root.AddButton("Add SSH Override", func() {
		target, user, identity := createOverrideInputs()

		f.overrides = append(f.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		f.root.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	})

	f.root.AddButton("New", func() {
		for _, o := range f.overrides {
			for range o {
				// TODO find a better way
				// overrides start at 4th index
				f.root.RemoveFormItem(4)
			}
		}

		f.overrides = []map[string]*tview.InputField{}
		f.configName.SetText("")
		f.cidrInput.SetText("")
		f.sshUserInput.SetText("")
		f.sshIdentityInput.SetText("")
		f.creatingNewConfig = true
	})

	f.root.AddButton("Save", func() {
		name := f.configName.GetText()
		cidr := f.cidrInput.GetText()
		sshUser := f.sshUserInput.GetText()
		sshIdentity := f.sshIdentityInput.GetText()

		if name == "" || cidr == "" || sshUser == "" || sshIdentity == "" {
			f.creatingNewConfig = false
			return
		}

		targets := strings.Split(cidr, ",")
		confOverrides := []config.SSHOverride{}

		for _, o := range f.overrides {
			confOverride := config.SSHOverride{
				Target:   o["target"].GetText(),
				User:     o["user"].GetText(),
				Identity: o["identity"].GetText(),
			}

			confOverrides = append(confOverrides, confOverride)
		}

		conf := config.Config{
			Name: name,
			SSH: config.SSHConfig{
				User:      sshUser,
				Identity:  sshIdentity,
				Overrides: confOverrides,
			},
			Targets: targets,
		}

		if f.creatingNewConfig {
			f.creatingNewConfig = false
			f.onCreate(conf)
			return
		}

		conf.ID = f.conf.ID
		f.onUpdate(conf)
	})
}
