package main

import (
	"askgpt/input"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	stateStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return titleStyle.BorderStyle(b)
	}()

	focusStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#71eb34"))
	}()
)

type keyMap struct {
	Switch key.Binding
	Quit   key.Binding
	Send   key.Binding
}

var keys = keyMap{
	Switch: key.NewBinding(key.WithKeys("ctrl+j", "ctrl+k")),
	Quit:   key.NewBinding(key.WithKeys("ctrl+c")),
	Send:   key.NewBinding(key.WithKeys("enter")),
}

type model struct {
	content       string
	ready         bool
	viewport      viewport.Model
	textInput     input.InputModel
	textInputMode bool
	render        *glamour.TermRenderer
	gpt           gptModel
	isAnswering   bool
}

func (m *model) renderContent() {
	if content, err := m.render.Render(m.content); err == nil {
		m.viewport.SetContent(content)
		m.viewport.GotoBottom()
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.gpt.newConversation(), m.gpt.waitForResponse())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	// handle mouse events in the viewport
	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Switch):
			if m.textInput.Focused() {
				m.textInput.Blur()
			} else {
				m.textInput.Focus()
			}
		case key.Matches(msg, keys.Send):
			if m.textInput.Focused() && !m.isAnswering {
				question := m.textInput.Value()

				m.content += fmt.Sprintf("You ->\n%s\n\n", question)
				m.content += fmt.Sprintf("GPT ->\n")
				m.renderContent()

				m.textInput.Reset()
				m.isAnswering = true

				return m, m.gpt.newQuestion(question)
			}
		default:
			if !m.textInput.Focused() {
				m.viewport, cmd = m.viewport.Update(msg)
			}
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		textinputHeight := lipgloss.Height(m.textInput.View())
		verticalMarginHeight := headerHeight + footerHeight + textinputHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	case respMsg:
		if msg.finished {
			m.isAnswering = false
			m.content += "\n\n"
		} else {
			m.content += msg.content
		}
		m.renderContent()
		return m, m.gpt.waitForResponse()
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView(), m.textInput.View())
}

func (m model) headerView() string {
	title := titleStyle.Render("AskGPT")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	if !m.textInput.Focused() {
		line = focusStyle.Render(line)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	var line string
	if m.isAnswering {
		state := stateStyle.Render("Answering")
		line = strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)-lipgloss.Width(state)))
		line = lipgloss.JoinHorizontal(lipgloss.Center, state, line)
	} else {
		line = strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	}
	if m.textInput.Focused() {
		line = focusStyle.Render(line)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func main() {
	render, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(180))
	if err != nil {
		fmt.Println(err)
		return
	}

	baseurl := "https://dashscope.aliyuncs.com/compatible-mode/v1"
	apikey, ok := os.LookupEnv("API_KEY")
	if !ok || apikey == "" {
		fmt.Println("no api key")
		os.Exit(1)
	}

	p := tea.NewProgram(
		model{
			render: render,
			gpt: gptModel{
				client: newGPTClient(baseurl, apikey),
				model:  "deepseek-v3",
				askCh:  make(chan string),
				respCh: make(chan respMsg),
			},
			textInput: input.New(),
		},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
