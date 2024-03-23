package main

import (
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

	commands, err := c.Prompt(strings.Join(os.Args[1:], " "))
	if err != nil {
		log.Fatalf("Failed to prompt: %v", err)
	}
	if len(commands) == 0 {
		exit("Couldn't figure out what to execute...")
	}

	for _, command := range commands {
		fmt.Println(command)
	}
	prompt := promptui.Prompt{
		Label:     "Execute",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Couldn't confirm: %v", err)
	}
	if result == "y" {
		for _, command := range commands {
			cmd := exec.Command("sh", "-c", command)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to run command: %v", err)
			}
		}
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
