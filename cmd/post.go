/*
Copyright © 2024 Brian Greenhill <brian@briangreenhill.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Create a new post",
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config
		if err := viper.Unmarshal(&cfg); err != nil {
			slog.Error(
				"error unmarshalling config",
				slog.Any("error.message", err.Error()))
			os.Exit(1)
		}

		if err := createPost(&cfg); err != nil {
			slog.Error(
				"error creating post",
				slog.Any("error.message", err.Error()))
		}
	},
}

func createPost(cfg *config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	var p post
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("title").Placeholder("Title").Value(&p.Title),
		huh.NewInput().Title("date").Placeholder("Date").Value(&p.Date),
		huh.NewInput().Title("author").Placeholder("Author").Value(&p.Author),
		huh.NewInput().Title("authorimage").Placeholder("Author Image").Value(&p.AuthorImg),
		huh.NewInput().Title("description").Placeholder("Description").Value(&p.Description),
	))

	if err := form.Run(); err != nil {
		return fmt.Errorf("error running form: %w", err)
	}

	if p.Date == "" {
		p.Date = time.Now().Format("2006-01-02")
	}

	if p.Author == "" {
		p.Author = cfg.Author
	}

	if p.AuthorImg == "" {
		p.AuthorImg = cfg.AuthorImg
	}

	if p.Description == "" {
		p.Description = "Replace this with a short description of the post"
	}

	// create frontmatter from form values
	frontmatter := fmt.Sprintf(`---
title: %s
date: %s
author: %s
author_image: %s
description: %s
---

Your post content goes here!

`, p.Title, p.Date, p.Author, p.AuthorImg, p.Description)

	var shouldCreate = true
	filename := fmt.Sprintf("%s/%s-%s.md", filepath.Join(cfg.ContentDir, postsDirName), p.Date, strings.ReplaceAll(p.Title, " ", "_"))
	if _, err := os.Stat(filename); err == nil {
		if err := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Title("Post already exists. Continue?").Value(&shouldCreate),
		)).Run(); err != nil {
			return fmt.Errorf("error running form: %w", err)
		}
	}

	if shouldCreate == false {
		fmt.Println("Post creation cancelled")
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating post file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(frontmatter)
	if err != nil {
		return fmt.Errorf("error writing frontmatter to post file: %w", err)
	}

	fmt.Printf("Created %s at %s\n", p.Title, filename)

	return nil
}

func init() {
	rootCmd.AddCommand(postCmd)
}
