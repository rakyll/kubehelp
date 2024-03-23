package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rakyll/kubehelp/client"
	"github.com/rakyll/kubehelp/history"
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
	historyStore, err := history.NewStore("")
	if err != nil {
		log.Fatalf("Failed to create history store: %v", err)
	}

	commands, err := c.Prompt(strings.Join(os.Args[1:], " "))
	if err != nil {
		log.Fatalf("Failed to prompt: %v", err)
	}
	if len(commands) == 0 {
		exit("Couldn't figure out what kubectl commands to execute...")
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
			if err := historyStore.Append(command); err != nil {
				log.Printf("Failed to append to history: %v", err)
			}
		}
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
