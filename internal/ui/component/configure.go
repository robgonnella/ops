package component

import (
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
	sshPortInput      *tview.InputField
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
) (*tview.InputField, *tview.InputField, *tview.InputField, *tview.InputField, *tview.InputField) {
	configName := tview.NewInputField()
	configName.SetLabel("Config Name: ")

	sshUserInput := tview.NewInputField()
	sshUserInput.SetLabel("SSH User: ")

	sshIdentityInput := tview.NewInputField()
	sshIdentityInput.SetLabel("SSH Identity: ")

	sshPortInput := tview.NewInputField()
	sshPortInput.SetLabel("SSH Port: ")

	cidrInput := tview.NewInputField()
	cidrInput.SetLabel("Network CIDR: ")

	form.AddFormItem(configName)
	form.AddFormItem(cidrInput)
	form.AddFormItem(sshUserInput)
	form.AddFormItem(sshIdentityInput)
	form.AddFormItem(sshPortInput)

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

	return configName, sshUserInput, sshIdentityInput, sshPortInput, cidrInput
}

// every time the add ssh override button is clicked we add three new inputs
func createOverrideInputs(conf config.Config) (*tview.InputField, *tview.InputField, *tview.InputField, *tview.InputField) {
	overrideTarget := tview.NewInputField()
	overrideTarget.SetLabel("Override Target IP: ")

	overrideSSHUser := tview.NewInputField()
	overrideSSHUser.SetLabel("Override SSH User: ")
	overrideSSHUser.SetText(conf.SSH.User)

	overrideSSHIdentity := tview.NewInputField()
	overrideSSHIdentity.SetLabel("Override SSH Identity: ")
	overrideSSHIdentity.SetText(conf.SSH.Identity)

	overrideSSHPort := tview.NewInputField()
	overrideSSHPort.SetLabel("Override SSH Port: ")
	overrideSSHPort.SetText(conf.SSH.Port)

	return overrideTarget, overrideSSHUser, overrideSSHIdentity, overrideSSHPort
}

// NewConfigureForm returns a new instance of ConfigureForm
func NewConfigureForm(
	conf config.Config,
	onUpdate func(conf config.Config),
	onCreate func(conf config.Config),
	onDismiss func(),
) *ConfigureForm {
	form := tview.NewForm()

	configName, sshUserInput, sshIdentityInput, sshPortInput, cidrInput := addBlankFormItems(
		form,
		conf.Name,
	)

	return &ConfigureForm{
		root:              form,
		configName:        configName,
		sshUserInput:      sshUserInput,
		sshIdentityInput:  sshIdentityInput,
		sshPortInput:      sshPortInput,
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

	f.configName, f.sshUserInput, f.sshIdentityInput, f.sshPortInput, f.cidrInput =
		addBlankFormItems(f.root, f.conf.Name)

	networkTargets := f.conf.CIDR

	f.configName.SetText(f.conf.Name)
	f.sshUserInput.SetText(f.conf.SSH.User)
	f.sshIdentityInput.SetText(f.conf.SSH.Identity)
	f.sshPortInput.SetText(f.conf.SSH.Port)
	f.cidrInput.SetText(networkTargets)

	for _, o := range f.conf.SSH.Overrides {
		target, user, identity, port := createOverrideInputs(f.conf)

		f.overrides = append(f.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
			"port":     port,
		})

		target.SetText(o.Target)
		user.SetText(o.User)
		identity.SetText(o.Identity)
		port.SetText(o.Port)

		f.root.
			AddFormItem(target).
			AddFormItem(user).
			AddFormItem(identity).
			AddFormItem(port)
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
		target, user, identity, port := createOverrideInputs(f.conf)

		f.overrides = append(f.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
			"port":     port,
		})

		f.root.
			AddFormItem(target).
			AddFormItem(user).
			AddFormItem(identity).
			AddFormItem(port)
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
		f.sshPortInput.SetText("")
		f.creatingNewConfig = true
	})

	f.root.AddButton("Save", func() {
		name := f.configName.GetText()
		cidr := f.cidrInput.GetText()
		sshUser := f.sshUserInput.GetText()
		sshIdentity := f.sshIdentityInput.GetText()
		sshPort := f.sshPortInput.GetText()

		if name == "" || cidr == "" || sshUser == "" || sshIdentity == "" || sshPort == "" {
			f.creatingNewConfig = false
			return
		}

		confOverrides := []config.SSHOverride{}

		for _, o := range f.overrides {
			confOverride := config.SSHOverride{
				Target:   o["target"].GetText(),
				User:     o["user"].GetText(),
				Identity: o["identity"].GetText(),
				Port:     o["port"].GetText(),
			}

			confOverrides = append(confOverrides, confOverride)
		}

		conf := config.Config{
			Name: name,
			SSH: config.SSHConfig{
				User:      sshUser,
				Identity:  sshIdentity,
				Port:      sshPort,
				Overrides: confOverrides,
			},
			CIDR: cidr,
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
