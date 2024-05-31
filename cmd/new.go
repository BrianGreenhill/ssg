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
	"log"
	"os"
	"os/exec"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	themeRepo = "https://github.com/briangreenhill/ssg"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new site",
	Run: func(cmd *cobra.Command, args []string) {
		if err := createSite(); err != nil {
			log.Fatal("error creating site ", err)
		}
	},
}

func createSite() error {
	needSite := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Do you want to delete the existing site? You cannot undo this action.").
				Value(&needSite)),
	)

	if _, err := os.Stat(".ssg.yaml"); err == nil {
		if err := form.Run(); err != nil {
			return fmt.Errorf("error running form: %w", err)
		}
	}

	if needSite {
		var cfg config
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter the theme").
					Placeholder("default").
					Value(&cfg.Theme),
				huh.NewInput().
					Title("Enter the site title").
					Placeholder("My Site").
					Value(&cfg.Title),
				huh.NewInput().
					Title("Enter a description of the site").
					Placeholder("A site about things").
					Value(&cfg.Description),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Enter the author's name").
					Placeholder("John Doe").
					Value(&cfg.Author),
				huh.NewInput().
					Title("Enter an image URL of the author").
					Placeholder("https://example.com/image.jpg").
					Value(&cfg.AuthorImg),
				huh.NewInput().
					Title("Enter the GitHub URL").
					Placeholder("https://github.com/mona").
					Value(&cfg.Github),
				huh.NewInput().
					Title("Enter the LinkedIn URL").
					Placeholder("https://linkedin.com/in/mona").
					Value(&cfg.Linkedin),
				huh.NewInput().
					Title("Enter the email address").
					Placeholder("user@email.com").
					Value(&cfg.Email),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("error running form: %w", err)
		}

		if cfg.Theme == "" {
			viper.Set("theme", "default")
			cfg.Theme = "default"
		}
		if cfg.ContentDir == "" {
			viper.Set("contentDir", "content")
			cfg.ContentDir = "content"
		}
		if cfg.OutputDir == "" {
			viper.Set("outputDir", "public")
			cfg.OutputDir = "public"
		}
		if cfg.Title == "" {
			viper.Set("title", "Finn the Human")
			cfg.Title = "Finn the Human"
		}
		if cfg.Author == "" {
			viper.Set("author", "Finn the Human")
			cfg.Author = "Finn the Human"
		}
		if cfg.AuthorImg == "" {
			viper.Set("authorImg", "https://octodex.github.com/images/adventure-cat.png")
			cfg.AuthorImg = "https://octodex.github.com/images/adventure-cat.png"
		}
		if cfg.Description == "" {
			viper.Set("description", "Mathematical!")
			cfg.Description = "Mathematical!"
		}
		if cfg.Github == "" {
			viper.Set("github", "https://github.com/mona")
			cfg.Github = "https://github.com/mona"
		}
		if cfg.Linkedin == "" {
			viper.Set("linkedin", "https://linkedin.com/in/mona")
			cfg.Linkedin = "https://linkedin.com/in/mona"
		}

		viper.Set("theme", cfg.Theme)
		viper.Set("contentDir", cfg.ContentDir)
		viper.Set("outputDir", cfg.OutputDir)
		viper.Set("title", cfg.Title)
		viper.Set("author", cfg.Author)
		viper.Set("authorImg", cfg.AuthorImg)
		viper.Set("description", cfg.Description)
		viper.Set("github", cfg.Github)
		viper.Set("linkedin", cfg.Linkedin)
		viper.Set("email", cfg.Email)

		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("error writing config: %w", err)
		}

		fmt.Println("Selecting theme and downloading...")
		if err := selectTheme(cfg.Theme); err != nil {
			return fmt.Errorf("error selecting theme: %w", err)
		}
		fmt.Println("downloaded theme", cfg.Theme)
	}

	fmt.Println("Generating site...")

	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("unable to decode into struct: %w", err)
	}

	if err := generateSite(&cfg); err != nil {
		return err
	}

	fmt.Println("Site generated successfully. Run `ssg watch` to start the server.")
	return nil
}

func selectTheme(theme string) error {
	if theme == "" {
		return fmt.Errorf("theme cannot be empty")
	}

	themes := []string{"default"}

	if !slices.Contains(themes, theme) {
		return fmt.Errorf("theme %s not found", theme)
	}

	if err := downloadTheme(theme); err != nil {
		return fmt.Errorf("error downloading theme: %w", err)
	}

	return nil
}

func downloadTheme(theme string) error {
	directory := "tmp"

	// make tmp directory
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", directory, err)
	}

	// download theme from github to tmp directory
	cmd := exec.Command("git", "clone", themeRepo, fmt.Sprintf("tmp"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning theme: %w", err)
	}

	// copy theme to the current directory
	themeDir := fmt.Sprintf("tmp/themes/%s", theme)
	if err := copyDir(themeDir, "themes"); err != nil {
		return fmt.Errorf("error copying theme: %w", err)
	}

	// remove tmp directory
	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("error removing directory %s: %w", directory, err)
	}

	return nil
}

func copyDir(src, dest string) error {
	// copy src directory to destination

	// check if src directory exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %w", err)
	}

	// check if dest directory exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dest, err)
		}
	}

	// copy files from src to dest
	cmd := exec.Command("cp", "-r", src, dest)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error copying directory: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)
}
