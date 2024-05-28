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
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new site",
	Run: func(cmd *cobra.Command, args []string) {
		if err := newSite(); err != nil {
			fmt.Println("error creating new site")
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func newSite() error {
	needSite := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Do you want to delete the existing site? You cannot undo this action.").
				Value(&needSite)),
	)

	if _, err := os.Stat(".ssg.yaml"); err == nil {
		if err := form.Run(); err != nil {
			return err
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
			return err
		}

		answers := map[string]string{
			"theme":       "default",
			"contentDir":  "content",
			"outputDir":   "public",
			"title":       cfg.Title,
			"author":      cfg.Author,
			"authorImg":   cfg.AuthorImg,
			"description": cfg.Description,
			"github":      cfg.Github,
			"linkedin":    cfg.Linkedin,
			"email":       cfg.Email,
		}

		configFile, err := os.Create(".ssg.yaml")
		if err != nil {
			return err
		}
		defer configFile.Close()

		for k, v := range answers {
			_, err := configFile.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			if err != nil {
				return err
			}
		}

		for k, v := range answers {
			fmt.Printf("%s: %s\n", k, v)
		}
	}

	fmt.Println("Generating site...")

	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := generateSite(cfg); err != nil {
		return err
	}

	fmt.Println("Site generated successfully. Run `ssg watch` to start the server.")
	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)
}
