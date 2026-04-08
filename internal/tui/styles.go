package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ilova-bazis/umati/internal/schema"
)

// System terminal colors (ANSI 0-15) - adapts to user's terminal theme
const (
	// Base colors
	colorBg         = lipgloss.Color("0") // black (terminal background)
	colorBorder     = lipgloss.Color("8") // bright black
	colorSelected   = lipgloss.Color("4") // blue
	colorCursor     = lipgloss.Color("4") // blue
	colorText       = lipgloss.Color("7") // white (terminal foreground)
	colorDim        = lipgloss.Color("8") // bright black
	colorSuccess    = lipgloss.Color("2") // green
	colorErr        = lipgloss.Color("1") // red
	colorHeader     = lipgloss.Color("6") // cyan
	colorHeaderBg   = lipgloss.Color("0") // black
	colorInputFg    = lipgloss.Color("7") // white
	colorInputBg    = lipgloss.Color("0") // black
	colorInputFocus = lipgloss.Color("6") // cyan

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
	styleHeader = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorText).
			Padding(0, 1)

	styleColHeader = lipgloss.NewStyle().
			Foreground(colorHeader).
			Bold(true)

	styleColHeaderActive = lipgloss.NewStyle().
				Foreground(colorSelected).
				Bold(true)

	styleColSep = lipgloss.NewStyle().
			Foreground(colorBorder)

	styleCardNormal = lipgloss.NewStyle().
			Foreground(colorText)

	styleCardSelected = lipgloss.NewStyle().
				Background(colorCursor).
				Foreground(colorText)

	styleCardID = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	styleCardIDSelected = lipgloss.NewStyle().
				Bold(true).
				Background(colorCursor).
				Foreground(colorText)

	styleCardDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleCardDimSelected = lipgloss.NewStyle().
				Background(colorCursor).
				Foreground(colorDim)

	styleStatusBar = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorDim).
			Padding(0, 1)

	styleStatusMsg = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorSuccess)

	styleStatusErr = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorErr)

	styleDetailPanel = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorSelected).
				Padding(0, 1)

	styleOverlay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSelected).
			Background(colorBg).
			Padding(1, 2)

	styleLabelFg = lipgloss.NewStyle().
			Foreground(colorDim)

	styleValueFg = lipgloss.NewStyle().
			Foreground(colorText)

	styleFocusedLabel = lipgloss.NewStyle().
				Foreground(colorSelected).
				Bold(true)
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
