package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rakyll/kubehelp/client"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY is required")
	}

	if len(os.Args) < 2 {
		exit("Usage: kubehelp <prompt>")
	}

	c := client.NewClient(apiKey)

	data, err := c.Prompt(strings.Join(os.Args[1:], " "))
	if err != nil {
		log.Fatalf("Failed to prompt: %v", err)
	}
	commands := data.Commands
	if len(commands) == 0 {
		exit("Couldn't figure out what to execute...")
	}

	fmt.Println("Commands:")
	for _, command := range commands {
		fmt.Println(">>> ", command)
	}

	fmt.Printf("\nExplanation:\n%s\n\n", data.Explanation)
	prompt := promptui.Prompt{
		Label:     "Execute commands?",
		IsConfirm: true,
		Validate:  validate,
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Couldn't confirm: %v", err)
	}
	if result != "y" {
		return
	}
	for _, command := range commands {
		cmd := exec.Command("sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to run command: %v", err)
		}
	}
}

func validate(input string) error {
	input = strings.ToLower(input)
	if input == "" {
		return errors.New("Input is required. Please enter 'y' or 'n'.")
	}
	if input != "y" && input != "n" {
		return errors.New("Invalid input. Please enter 'y' or 'n'.")
	}
	return nil
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
