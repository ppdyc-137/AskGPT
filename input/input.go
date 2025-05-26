package input

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModel struct {
	ti      textinput.Model
	focused bool
	mode    string
	KeyMap  KeyMap
}

type KeyMap struct {
	Left   key.Binding
	Right  key.Binding
	Escape key.Binding
	Insert key.Binding
}

var DefaultKeyMap = KeyMap{
	Left:   key.NewBinding(key.WithKeys("h")),
	Right:  key.NewBinding(key.WithKeys("l")),
	Escape: key.NewBinding(key.WithKeys("esc")),
	Insert: key.NewBinding(key.WithKeys("i")),
}

var (
	insert string = "INSERT"
	normal string = "NORMAL"
)

func New() InputModel {
	ti := textinput.New()
	ti.Focus()

	return InputModel{
		ti:      ti,
		focused: true,
		mode:    insert,
		KeyMap:  DefaultKeyMap,
	}
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	var cmd tea.Cmd

	// handle mode switching
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Escape):
			m.enterNormal()
		case key.Matches(msg, m.KeyMap.Insert) && m.mode == normal:
			m.enterInsert()
		}
	}

	// handle cursor in normal mode
	if m.mode == normal {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.KeyMap.Left):
				m.ti.SetCursor(m.ti.Position() - 1)
			case key.Matches(msg, m.KeyMap.Right):
				m.ti.SetCursor(m.ti.Position() + 1)
			}
		}
	}

	if m.mode == insert {
		m.ti, cmd = m.ti.Update(msg)
	}

	return m, cmd
}

func (m *InputModel) enterNormal() {
	m.mode = normal
	m.ti.Cursor.Blink = false
	// m.ti.Blur()
}

func (m *InputModel) enterInsert() {
	m.mode = insert
	// m.ti.Focus()
}

func (m InputModel) Value() string {
	return m.ti.Value()
}

func (m *InputModel) Reset() {
	m.ti.Reset()
}

func (m InputModel) View() string {
	return m.ti.View()
}

func (m InputModel) Focused() bool {
	return m.focused
}

func (m *InputModel) Focus() {
	m.focused = true
	m.ti.Focus()
}

func (m *InputModel) Blur() {
	m.focused = false
	m.ti.Blur()
}
