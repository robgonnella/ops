package component

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// AddButtons()) or it will never disappear.
//
// See https://github.com/rivo/tview/wiki/Modal for an example.
type TViewModal struct {
	*tview.Box

	// The frame embedded in the modal.
	frame *tview.Frame

	// The form embedded in the modal's frame.
	form *tview.Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)
}

// NewModal returns a new modal message window.
func NewTViewModal() *TViewModal {
	m := &TViewModal{
		Box:       tview.NewBox(),
		textColor: tview.Styles.PrimaryTextColor,
	}
	m.form = tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.frame = tview.NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *TViewModal) SetBackgroundColor(color tcell.Color) *TViewModal {
	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *TViewModal) SetTextColor(color tcell.Color) *TViewModal {
	m.textColor = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *TViewModal) SetButtonBackgroundColor(color tcell.Color) *TViewModal {
	m.form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *TViewModal) SetButtonTextColor(color tcell.Color) *TViewModal {
	m.form.SetButtonTextColor(color)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an empty string.
func (m *TViewModal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *TViewModal {
	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks but color tag states will not transfer to following lines. Note that
// words are wrapped, too, based on the final size of the window.
func (m *TViewModal) SetText(text string) *TViewModal {
	m.text = text
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *TViewModal) AddButtons(labels []string) *TViewModal {
	for index, label := range labels {
		func(i int, l string) {
			m.form.AddButton(label, func() {
				if m.done != nil {
					m.done(i, l)
				}
			})
			button := m.form.GetButton(m.form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

func (m *TViewModal) SetButtonActivatedStyle(style tcell.Style) {
	m.form.SetButtonActivatedStyle(style)
}

// ClearButtons removes all buttons from the window.
func (m *TViewModal) ClearButtons() *TViewModal {
	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *TViewModal) SetFocus(index int) *TViewModal {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *TViewModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *TViewModal) HasFocus() bool {
	return m.form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *TViewModal) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	buttonCount := m.form.GetButtonCount()

	for i := 0; i < buttonCount-1; i++ {
		button := m.form.GetButton(i)
		buttonsWidth += tview.TaggedStringWidth(button.GetLabel()) + 4 + 2
	}

	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	var lines []string
	for _, line := range strings.Split(m.text, "\n") {
		if len(line) == 0 {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, tview.WordWrap(line, width)...)
	}
	//lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, tview.AlignCenter, m.textColor)
	}

	// Set the modal's position and size.
	height := len(lines) + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

// MouseHandler returns the mouse handler for this primitive.
func (m *TViewModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftDown && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *TViewModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
