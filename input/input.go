package input

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#34eb8f"))
	}()
)

type InputModel struct {
	ti      textinput.Model
	focused bool
	mode    string

	KeyMap           KeyMap
	InsertModePrompt string
	NormalModePrompt string
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
	insertMode string = "INSERT"
	normalMode string = "NORMAL"
)

func New() InputModel {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = promptStyle.Render("> ")

	return InputModel{
		ti:               ti,
		focused:          true,
		mode:             insertMode,
		KeyMap:           DefaultKeyMap,
		InsertModePrompt: promptStyle.Render("> "),
		NormalModePrompt: promptStyle.Render("< "),
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
		case key.Matches(msg, m.KeyMap.Insert) && m.mode == normalMode:
			m.enterInsert()
			return m, nil
		}
	}

	// handle cursor in normal mode
	if m.mode == normalMode {
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

	if m.mode == insertMode {
		m.ti, cmd = m.ti.Update(msg)
	}

	return m, cmd
}

func (m *InputModel) enterNormal() {
	m.mode = normalMode
	m.ti.Cursor.Blink = false
	m.ti.Prompt = m.NormalModePrompt
}

func (m *InputModel) enterInsert() {
	m.mode = insertMode
	m.ti.Prompt = m.InsertModePrompt
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
