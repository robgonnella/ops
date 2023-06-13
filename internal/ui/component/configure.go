package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/ui/style"
)

type ConfigureForm struct {
	root *tview.Form
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
	overrides := []map[string]*tview.InputField{}

	networkTargets := strings.Join(conf.Targets, ",")

	configName := tview.NewInputField()
	configName.SetLabel("Config Name: ")
	configName.SetText(conf.Name)

	sshUserInput := tview.NewInputField()
	sshUserInput.SetLabel("SSH User: ")
	sshUserInput.SetText(conf.SSH.User)

	sshIdentityInput := tview.NewInputField()
	sshIdentityInput.SetLabel("SSH Identity: ")
	sshIdentityInput.SetText(conf.SSH.Identity)

	cidrInput := tview.NewInputField()
	cidrInput.SetLabel("Comma Separated CIDRs or IPs: ")
	cidrInput.SetText(networkTargets)

	form := tview.NewForm()
	form.SetTitle("Configure Network Scanning")
	form.AddFormItem(configName)
	form.AddFormItem(cidrInput)
	form.AddFormItem(sshUserInput)
	form.AddFormItem(sshIdentityInput)

	for _, o := range conf.SSH.Overrides {
		target, user, identity := createOverrideInputs()

		overrides = append(overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		target.SetText(o.Target)
		user.SetText(o.User)
		identity.SetText(o.Identity)

		form.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	}

	form.AddButton("Add SSH Overrides", func() {
		target, user, identity := createOverrideInputs()

		overrides = append(overrides, map[string]*tview.InputField{
			"target":   target,
			"user":     user,
			"identity": identity,
		})

		form.AddFormItem(target).AddFormItem(user).AddFormItem(identity)
	})

	form.AddButton("OK", func() {
		name := configName.GetText()
		cidr := cidrInput.GetText()
		targets := strings.Split(cidr, ",")
		confOverrides := []config.SSHOverride{}

		for _, o := range overrides {
			confOverride := config.SSHOverride{
				Target:   o["target"].GetText(),
				User:     o["user"].GetText(),
				Identity: o["identity"].GetText(),
			}

			confOverrides = append(confOverrides, confOverride)
		}

		conf.Name = name
		conf.SSH.User = sshUserInput.GetText()
		conf.SSH.Identity = sshIdentityInput.GetText()
		conf.SSH.Overrides = confOverrides
		conf.Targets = targets

		onSubmit(conf)
	})

	form.SetTitle(conf.Name + " Configuration")
	form.SetBorder(true)
	form.SetBorderColor(style.ColorPurple)
	form.SetFieldBackgroundColor(tcell.ColorDefault)
	form.SetButtonBackgroundColor(style.ColorMediumGreen)
	form.SetButtonTextColor(style.ColorWhite)

	return &ConfigureForm{root: form}
}

func (f *ConfigureForm) Primitive() tview.Primitive {
	return f.root
}
