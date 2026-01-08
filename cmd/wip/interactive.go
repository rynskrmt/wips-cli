package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rynskrmt/wips-cli/internal/usecase"
)

func runInteractive(u usecase.NoteUsecase) error {
	// Initialize CWD
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 3. Start Loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("wips-cli interactive mode.")
	fmt.Println("Type your memo and press Enter to record.")
	fmt.Println("Type \":help\" for available commands.")
	fmt.Println("Press Ctrl+C or Enter on an empty line to exit.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break // EOF or Error
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break // Empty line to exit
		}

		// Handle commands
		if strings.HasPrefix(line, ":") {
			cmd := strings.ToLower(line)
			switch cmd {
			case ":help":
				fmt.Println("Available commands:")
				fmt.Println("  :help        Show this help message")
				fmt.Println("  :exit, :quit Exit interactive mode")
				fmt.Println("  <text>       Record note")
				fmt.Println("\nTip: Run `wip --help` after exiting for other commands (edit, summary, etc.)")
				continue
			case ":exit", ":quit":
				fmt.Println("Bye!")
				return nil
			}
			// If not matched, treat as normal text (or warn? let's treat as text for now to allow notes starting with :)
			// But usually REPLs are strict. Let's record it but maybe print a hint if it looks like a typo?
			// For now, just record it.
		}

		event, err := u.RecordNote(line, cwd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if event != nil {
			ts := event.TS.Format("15:04")

			// Format: [HH:mm] message (ID: <id>)
			c := color.New(color.FgCyan)
			c.Printf("[%s] %s ", ts, event.Content)

			// Light gray for ID. Fatih color doesn't have explicit "Gray" but Faint often works.
			// Using FgHiBlack as dark gray substitute/subtle color.
			idColor := color.New(color.FgHiBlack)
			idColor.Printf("(ID: %s)\n", event.ID)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading standard input: %w", err)
	}

	fmt.Println("Bye!")
	return nil
}
