package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// AskConfirmation asks for user confirmation
func AskConfirmation(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	return false
}

// PromptInput prompts for user input with a message
func PromptInput(message string) (string, error) {
	fmt.Printf("%s: ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", scanner.Err()
}

// PromptSelect prompts user to select from options
func PromptSelect(message string, options []string) (int, error) {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	for {
		fmt.Printf("Select (1-%d): ", len(options))
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err == nil && choice >= 1 && choice <= len(options) {
				return choice - 1, nil
			}
		}
		fmt.Printf("Invalid choice. Please select 1-%d.\n", len(options))
	}
}
