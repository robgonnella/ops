package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/ui/style"
)

type ConfigureForm struct {
	root             *tview.Form
	configName       *tview.InputField
	sshUserInput     *tview.InputField
	sshIdentityInput *tview.InputField
	cidrInput        *tview.InputField
	overrides        []map[string]*tview.InputField
	conf             config.Config
}

func generateBlankForm(conf config.Config) *ConfigureForm {
	overrides := []map[string]*tview.InputField{}

	configName := tview.NewInputField()
	configName.SetLabel("Config Name: ")

	sshUserInput := tview.NewInputField()
	sshUserInput.SetLabel("SSH User: ")

	sshIdentityInput := tview.NewInputField()
	sshIdentityInput.SetLabel("SSH Identity: ")

	cidrInput := tview.NewInputField()
	cidrInput.SetLabel("Comma Separated CIDRs or IPs: ")

	form := tview.NewForm()
	form.AddFormItem(configName)
	form.AddFormItem(cidrInput)
	form.AddFormItem(sshUserInput)
	form.AddFormItem(sshIdentityInput)

	form.SetTitle(conf.Name + " Configuration")
	form.SetBorder(true)
	form.SetBorderColor(style.ColorPurple)
	form.SetFieldBackgroundColor(tcell.ColorDefault)
	form.SetButtonBackgroundColor(style.ColorLightGreen)
	form.SetLabelColor(style.ColorOrange)
	form.SetButtonTextColor(style.ColorBlack)
	form.SetButtonActivatedStyle(
		style.StyleDefault.Background(style.ColorLightGreen),
	)

	return &ConfigureForm{
		root:             form,
		configName:       configName,
		sshUserInput:     sshUserInput,
		sshIdentityInput: sshIdentityInput,
		cidrInput:        cidrInput,
		overrides:        overrides,
		conf:             conf,
	}
}

func addFormButtons(form *ConfigureForm, onSubmit func(conf config.Config)) {
	form.root.AddButton("Add SSH Overrides", func() {
		target, user, identity := createOverrideInputs()

		form.overrides = append(form.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		form.root.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	})

	form.root.AddButton("New Config", func() {
		for _, o := range form.overrides {
			for range o {
				// TODO find a better way
				// overrides start at 4th index
				form.root.RemoveFormItem(4)
			}
		}

		form.configName.SetText("")
		form.cidrInput.SetText("")
		form.sshUserInput.SetText("")
		form.sshIdentityInput.SetText("")
	})

	form.root.AddButton("OK", func() {
		name := form.configName.GetText()
		cidr := form.cidrInput.GetText()
		sshUser := form.sshUserInput.GetText()
		sshIdentity := form.sshIdentityInput.GetText()

		if name == "" || cidr == "" || sshUser == "" || sshIdentity == "" {
			return
		}

		targets := strings.Split(cidr, ",")
		confOverrides := []config.SSHOverride{}

		for _, o := range form.overrides {
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

		onSubmit(conf)
	})
}

func createOverrideInputs() (*tview.InputField, *tview.InputField, *tview.InputField) {
	overrideTarget := tview.NewInputField()
	overrideTarget.SetLabel("Override Target: ")

	overrideSSHUser := tview.NewInputField()
	overrideSSHUser.SetLabel("Override SSH User: ")

	overrideSSHIdentity := tview.NewInputField()
	overrideSSHIdentity.SetLabel("Override SSH Identity: ")

	return overrideTarget, overrideSSHUser, overrideSSHIdentity
}

func NewConfigureForm(conf config.Config, onSubmit func(conf config.Config)) *ConfigureForm {
	form := generateBlankForm(conf)

	networkTargets := strings.Join(conf.Targets, ",")

	form.configName.SetText(conf.Name)
	form.sshUserInput.SetText(conf.SSH.User)
	form.sshIdentityInput.SetText(conf.SSH.Identity)
	form.cidrInput.SetText(networkTargets)

	for _, o := range conf.SSH.Overrides {
		target, user, identity := createOverrideInputs()

		form.overrides = append(form.overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		target.SetText(o.Target)
		user.SetText(o.User)
		identity.SetText(o.Identity)

		form.root.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	}

	addFormButtons(form, onSubmit)

	return form
}

func (f *ConfigureForm) Primitive() tview.Primitive {
	return f.root
}
