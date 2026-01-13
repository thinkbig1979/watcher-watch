package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProcessWatchers struct {
	PID     int
	Name    string
	Watches int
}

type model struct {
	table        table.Model
	maxWatches   int
	maxInstances int
	totalWatches int
	err          error
}

type refreshMsg struct {
	processes    []ProcessWatchers
	maxWatches   int
	maxInstances int
	err          error
}

// Package-level styles (created once, reused on every render)
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func getInotifyLimits() (maxWatches, maxInstances int, err error) {
	data, err := os.ReadFile("/proc/sys/fs/inotify/max_user_watches")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read max_user_watches: %w", err)
	}
	maxWatches, err = strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse max_user_watches: %w", err)
	}

	data, err = os.ReadFile("/proc/sys/fs/inotify/max_user_instances")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read max_user_instances: %w", err)
	}
	maxInstances, err = strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse max_user_instances: %w", err)
	}

	return maxWatches, maxInstances, nil
}

func getProcessName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func countInotifyWatches() ([]ProcessWatchers, error) {
	processWatches := make(map[int]int)

	procs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		return nil, err
	}

	for _, proc := range procs {
		pid, err := strconv.Atoi(filepath.Base(proc))
		if err != nil {
			continue
		}

		fdPath := filepath.Join(proc, "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}

			if strings.HasPrefix(link, "anon_inode:inotify") {
				// Count the watches for this inotify instance
				fdinfoPath := filepath.Join(proc, "fdinfo", fd.Name())
				data, err := os.ReadFile(fdinfoPath)
				if err != nil {
					continue
				}

				watches := strings.Count(string(data), "inotify wd:")
				processWatches[pid] += watches
			}
		}
	}

	result := make([]ProcessWatchers, 0, len(processWatches))
	for pid, watches := range processWatches {
		if watches > 0 {
			result = append(result, ProcessWatchers{
				PID:     pid,
				Name:    getProcessName(pid),
				Watches: watches,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Watches > result[j].Watches
	})

	return result, nil
}

func refreshData() tea.Msg {
	maxWatches, maxInstances, limitsErr := getInotifyLimits()
	processes, err := countInotifyWatches()
	if limitsErr != nil && err == nil {
		err = limitsErr
	}
	return refreshMsg{
		processes:    processes,
		maxWatches:   maxWatches,
		maxInstances: maxInstances,
		err:          err,
	}
}

// buildRows creates table rows from process data and returns total watch count
func buildRows(processes []ProcessWatchers, maxWatches int) ([]table.Row, int) {
	rows := make([]table.Row, 0, len(processes))
	totalWatches := 0

	for _, p := range processes {
		totalWatches += p.Watches
		var pct float64
		if maxWatches > 0 {
			pct = float64(p.Watches) / float64(maxWatches) * 100
		}
		rows = append(rows, table.Row{
			strconv.Itoa(p.PID),
			p.Name,
			strconv.Itoa(p.Watches),
			fmt.Sprintf("%.2f%%", pct),
		})
	}

	return rows, totalWatches
}

func initialModel() model {
	maxWatches, maxInstances, limitsErr := getInotifyLimits()
	processes, err := countInotifyWatches()
	if limitsErr != nil && err == nil {
		err = limitsErr
	}

	columns := []table.Column{
		{Title: "PID", Width: 10},
		{Title: "Process", Width: 30},
		{Title: "Watches", Width: 12},
		{Title: "% of Max", Width: 10},
	}

	rows, totalWatches := buildRows(processes, maxWatches)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return model{
		table:        t,
		maxWatches:   maxWatches,
		maxInstances: maxInstances,
		totalWatches: totalWatches,
		err:          err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, refreshData
		}
	case refreshMsg:
		m.maxWatches = msg.maxWatches
		m.maxInstances = msg.maxInstances
		m.err = msg.err
		rows, totalWatches := buildRows(msg.processes, m.maxWatches)
		m.totalWatches = totalWatches
		m.table.SetRows(rows)
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress r to retry or q to quit.", m.err)
	}

	var usagePct float64
	if m.maxWatches > 0 {
		usagePct = float64(m.totalWatches) / float64(m.maxWatches) * 100
	}
	usageColor := "82" // green
	if usagePct > 80 {
		usageColor = "196" // red
	} else if usagePct > 50 {
		usageColor = "214" // orange
	}

	usageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(usageColor)).
		Bold(true)

	header := titleStyle.Render("🔍 Inotify Watcher Usage")
	info := infoStyle.Render(fmt.Sprintf(
		"Max Watches: %d | Max Instances: %d | Total Used: %s",
		m.maxWatches,
		m.maxInstances,
		usageStyle.Render(fmt.Sprintf("%d (%.2f%%)", m.totalWatches, usagePct)),
	))

	help := infoStyle.Render("\n↑/↓: Navigate | r: Refresh | q: Quit")

	tableView := m.table.View()
	if m.totalWatches == 0 {
		tableView = infoStyle.Render("No inotify watchers found.")
	}

	return fmt.Sprintf("%s\n%s\n\n%s%s", header, info, tableView, help)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
