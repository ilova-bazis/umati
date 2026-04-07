package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// interactivePrompt handles interactive task creation prompts
func interactivePrompt() (*struct {
	title       string
	description string
	priority    string
	status      string
	parent      string
	agent       string
}, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("")
	fmt.Println("umati interactive task creation")
	fmt.Println("================================")
	fmt.Println("")

	// Title (required)
	title, err := promptString(reader, "Title", true, "")
	if err != nil {
		return nil, err
	}

	// Description (optional)
	description, err := promptString(reader, "Description", false, "")
	if err != nil {
		return nil, err
	}

	// Priority (optional, default: medium)
	priorityOptions := []string{"low", "medium", "high", "urgent"}
	priorityIdx, err := promptSelect(reader, "Priority", priorityOptions, 1) // 1 = medium
	if err != nil {
		return nil, err
	}
	priority := priorityOptions[priorityIdx]

	// Status (optional, default: draft)
	statusOptions := []string{"draft", "paused", "ready"}
	statusIdx, err := promptSelect(reader, "Status", statusOptions, 0) // 0 = draft
	if err != nil {
		return nil, err
	}
	status := statusOptions[statusIdx]

	// Parent (optional, default: none)
	parent, err := promptString(reader, "Parent task ID (or 'none')", false, "none")
	if err != nil {
		return nil, err
	}
	if parent == "none" {
		parent = ""
	}

	// Agent (required)
	agentOptions := []string{"human", "claude", "codex", "opencode"}
	agentIdx, err := promptSelect(reader, "Agent", agentOptions, -1) // -1 = no default
	if err != nil {
		return nil, err
	}
	agent := agentOptions[agentIdx]

	// Build result
	result := &struct {
		title       string
		description string
		priority    string
		status      string
		parent      string
		agent       string
	}{
		title:       title,
		description: description,
		priority:    priority,
		status:      status,
		parent:      parent,
		agent:       agent,
	}

	// Show summary and confirm
	if !promptConfirm(reader, result) {
		return nil, fmt.Errorf("cancelled by user")
	}

	return result, nil
}

// promptString prompts for a string value
func promptString(reader *bufio.Reader, label string, required bool, defaultValue string) (string, error) {
	for {
		// Build prompt
		prompt := label
		if required {
			prompt += " [required]"
		} else {
			prompt += " [optional]"
		}
		if defaultValue != "" {
			prompt += fmt.Sprintf(" [default: %s]", defaultValue)
		}
		prompt += "\n> "

		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		// Trim whitespace
		input = strings.TrimSpace(input)

		// Use default if empty
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}

		// Validate required
		if required && input == "" {
			fmt.Println("Error: This field is required. Please enter a value.")
			continue
		}

		return input, nil
	}
}

// promptSelect prompts for a selection from numbered options
func promptSelect(reader *bufio.Reader, label string, options []string, defaultIndex int) (int, error) {
	for {
		fmt.Printf("\n%s:\n", label)
		for i, opt := range options {
			marker := ""
			if i == defaultIndex {
				marker = " [default]"
			}
			fmt.Printf("  %d) %s%s\n", i+1, opt, marker)
		}

		if defaultIndex >= 0 {
			fmt.Printf("Select (1-%d or Enter for default): ", len(options))
		} else {
			fmt.Printf("Select (1-%d): ", len(options))
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return -1, err
		}

		input = strings.TrimSpace(input)

		// Use default if empty
		if input == "" && defaultIndex >= 0 {
			return defaultIndex, nil
		}

		// Parse selection
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			fmt.Printf("Error: Please enter a number between 1 and %d.\n", len(options))
			continue
		}

		return choice - 1, nil // Convert to 0-based index
	}
}

// promptConfirm shows a summary and asks for confirmation
func promptConfirm(reader *bufio.Reader, opts *struct {
	title       string
	description string
	priority    string
	status      string
	parent      string
	agent       string
}) bool {
	fmt.Println("\nCreate this task?")
	fmt.Println("==================")
	fmt.Printf("  Title:       %s\n", opts.title)

	desc := opts.description
	if desc == "" {
		desc = "(none)"
	}
	fmt.Printf("  Description: %s\n", desc)

	fmt.Printf("  Priority:    %s\n", opts.priority)
	fmt.Printf("  Status:      %s\n", opts.status)

	parent := opts.parent
	if parent == "" {
		parent = "none"
	}
	fmt.Printf("  Parent:      %s\n", parent)

	fmt.Printf("  Agent:       %s\n", opts.agent)
	fmt.Println()

	for {
		fmt.Print("Confirm (y/n): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}
