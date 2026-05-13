package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ilova-bazis/umati/internal/schema"
)

// System terminal colors (ANSI 0-15) - adapts to user's terminal theme
const (
	// Base colors
	colorBg      = lipgloss.Color("0") // black (terminal background)
	colorBorder  = lipgloss.Color("8") // bright black
	colorSelected = lipgloss.Color("4") // blue — used as foreground for borders/headers
	colorText    = lipgloss.Color("7") // white (terminal foreground)
	colorDim     = lipgloss.Color("8") // bright black
	colorSuccess    = lipgloss.Color("2") // green
	colorErr        = lipgloss.Color("1") // red
	colorHeader     = lipgloss.Color("6") // cyan
	colorHeaderBg   = lipgloss.Color("0") // black
	colorInputFg    = lipgloss.Color("7") // white
	colorInputBg    = lipgloss.Color("0") // black
	colorInputFocus = lipgloss.Color("6") // cyan

	// Nav bar colors
	colorNavKey    = lipgloss.Color("15") // bright white — key buttons (?, q, ↑↓, enter…)
	colorNavLabel  = lipgloss.Color("6")  // cyan — description labels (nav, col, expand…)
	colorColHeader = lipgloss.Color("5")  // magenta — column headers

	// Priority colors
	colorPrioLow    = lipgloss.Color("2") // green
	colorPrioMed    = lipgloss.Color("3") // yellow
	colorPrioHigh   = lipgloss.Color("1") // red
	colorPrioUrgent = lipgloss.Color("9") // bright red

	// Status colors
	colorDraft      = lipgloss.Color("8") // bright black (dimmed)
	colorPaused     = lipgloss.Color("3") // yellow (waiting)
	colorReady      = lipgloss.Color("4") // blue (ready to go)
	colorClaimed    = lipgloss.Color("5") // magenta (assigned)
	colorInProgress = lipgloss.Color("2") // green (active work)
	colorDone       = lipgloss.Color("8") // bright black (completed)
	colorCancelled  = lipgloss.Color("0") // black (hidden)
)

var (
	// Status bar across the top of the board.
	styleHeader = lipgloss.NewStyle().
			Reverse(true).
			Padding(0, 1)

	// Inactive column header — magenta.
	styleColHeader = lipgloss.NewStyle().
			Foreground(colorColHeader)

	// Active column header — bold magenta.
	styleColHeaderActive = lipgloss.NewStyle().
				Foreground(colorColHeader).
				Bold(true)

	styleColSep = lipgloss.NewStyle().
			Foreground(colorDim)

	styleCardNormal = lipgloss.NewStyle()

	// Card selection uses reverse-video so contrast is always guaranteed.
	styleCardSelected = lipgloss.NewStyle().Reverse(true)

	styleCardID = lipgloss.NewStyle().Bold(true)

	styleCardIDSelected = lipgloss.NewStyle().
				Bold(true).
				Reverse(true)

	styleCardDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleCardDimSelected = lipgloss.NewStyle().Reverse(true)

	// Status/error messages in the board footer.
	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim)

	styleStatusMsg = lipgloss.NewStyle().Foreground(colorSuccess)

	styleStatusErr = lipgloss.NewStyle().Foreground(colorErr)

	// Detail panel — dim border keeps it from competing with card content.
	styleDetailPanel = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorDim).
				Padding(0, 1)

	// Overlay (form, filter, help) — bright border makes it clearly float above the board.
	styleOverlay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorText).
			Padding(1, 2)

	// Form field labels.
	styleLabelFg = lipgloss.NewStyle().Foreground(colorColHeader)

	styleValueFg = lipgloss.NewStyle()

	// Focused label: bold only — no color, works in any theme.
	styleFocusedLabel = lipgloss.NewStyle().Bold(true)

	// Action hint key bracket e.g. "[e]" — bold to pop against the dim description.
	styleActionKey = lipgloss.NewStyle().Bold(true)

	// Nav bar — key buttons and label text.
	styleNavKey   = lipgloss.NewStyle().Foreground(colorNavKey)
	styleNavLabel = lipgloss.NewStyle().Foreground(colorNavLabel)

	// Workspace/agent badge in the status bar.
	styleWorkspaceBadge = lipgloss.NewStyle().
				Background(lipgloss.Color("13")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
)

func priorityStyle(p schema.Priority) lipgloss.Style {
	switch p {
	case schema.PriorityLow:
		return lipgloss.NewStyle().Foreground(colorPrioLow).Bold(true)
	case schema.PriorityMedium:
		return lipgloss.NewStyle().Foreground(colorPrioMed).Bold(true)
	case schema.PriorityHigh:
		return lipgloss.NewStyle().Foreground(colorPrioHigh).Bold(true)
	case schema.PriorityUrgent:
		return lipgloss.NewStyle().Foreground(colorPrioUrgent).Bold(true)
	}
	return lipgloss.NewStyle()
}

func statusStyle(s schema.Status) lipgloss.Style {
	switch s {
	case schema.StatusDraft:
		return lipgloss.NewStyle().Foreground(colorDraft)
	case schema.StatusPaused:
		return lipgloss.NewStyle().Foreground(colorPaused)
	case schema.StatusReady:
		return lipgloss.NewStyle().Foreground(colorReady)
	case schema.StatusClaimed:
		return lipgloss.NewStyle().Foreground(colorClaimed)
	case schema.StatusInProgress:
		return lipgloss.NewStyle().Foreground(colorInProgress)
	case schema.StatusDone:
		return lipgloss.NewStyle().Foreground(colorDone)
	case schema.StatusCancelled:
		return lipgloss.NewStyle().Foreground(colorCancelled)
	}
	return lipgloss.NewStyle()
}

func priorityAbbr(p schema.Priority) string {
	switch p {
	case schema.PriorityLow:
		return "LOW"
	case schema.PriorityMedium:
		return "MED"
	case schema.PriorityHigh:
		return "HIG"
	case schema.PriorityUrgent:
		return "URG"
	}
	return "???"
}

func columnLabel(s schema.Status) string {
	switch s {
	case schema.StatusDraft:
		return "DRAFT"
	case schema.StatusPaused:
		return "PAUSED"
	case schema.StatusReady:
		return "READY"
	case schema.StatusClaimed:
		return "CLAIMED"
	case schema.StatusInProgress:
		return "IN PROGRESS"
	case schema.StatusDone:
		return "DONE"
	case schema.StatusCancelled:
		return "CANCELLED"
	}
	return string(s)
}
