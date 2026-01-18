// Package main is the entry point for lazyactions.
package main

import (
	"fmt"
	"os"

	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/auth"
	"github.com/nnnkkk7/lazyactions/github"
	"github.com/nnnkkk7/lazyactions/repo"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Detect repository from current directory
	repoInfo, err := repo.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect repository: %w", err)
	}

	// Get authentication token (gh CLI -> GITHUB_TOKEN)
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get authentication: %w", err)
	}

	// Create GitHub client
	client := github.NewClient(token.Value(), repoInfo.Owner, repoInfo.Name)

	// Create repository struct
	repository := github.Repository{
		Owner: repoInfo.Owner,
		Name:  repoInfo.Name,
	}

	// Run TUI
	return app.Run(client, repository)
}
