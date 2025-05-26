package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModel struct {
	ti      textinput.Model
	focused bool
	mode    string
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
		case msg.Type == tea.KeyEsc:
			m.enterNormal()
		case msg.String() == "i" && m.mode == normal:
			m.enterInsert()
		}
	}

	// handle cursor in normal mode
	if m.mode == normal {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "h":
				m.ti.SetCursor(m.ti.Position() - 1)
			case "l":
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
