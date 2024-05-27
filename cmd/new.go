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

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new site",
	Run: func(cmd *cobra.Command, args []string) {
		// create a new site
		// interview user for site details
		// create site directories
		// pull the requested theme
		// create a new config file
		deletePrompt := promptui.Prompt{
			Label:     "Do you want to delete the existing site? You cannot undo this action.",
			IsConfirm: true,
		}

		needSite := true

		if _, err := os.Stat(".ssg.yaml"); err == nil {
			fmt.Println("A site already exists in this directory")
			_, err := deletePrompt.Run()
			if err != nil {
				// user said no
				needSite = false
			} else {
				if err := os.RemoveAll(".ssg.yaml"); err != nil {
					fmt.Println("error deleting existing site")
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println("Existing site deleted")
			}

		}

		if needSite {

			prompts := map[string]promptui.Prompt{
				"title":       {Label: "What is the title of your site?"},
				"author":      {Label: "What is your name?"},
				"authorImg":   {Label: "What is the URL of your image?"},
				"description": {Label: "What is the description of your site?"},
				"github":      {Label: "What is your Github URL?"},
				"linkedin":    {Label: "What is your Linkedin URL?"},
				"email":       {Label: "What is your email address?"},
				"outputDir":   {Label: "What is the output directory?", Default: "public"},
				"contentDir":  {Label: "What is the content directory?", Default: "content"},
				"theme":       {Label: "What theme do you want to use?", Default: "default"},
			}

			answers := make(map[string]string)

			for k, v := range prompts {
				result, err := v.Run()
				if err != nil {
					fmt.Println("error getting answer")
					fmt.Println(err)
					os.Exit(1)
				}
				answers[k] = result
			}

			// write answers to config file

			configFile, err := os.Create(".ssg.yaml")
			if err != nil {
				fmt.Println("error creating config file")
				fmt.Println(err)
				os.Exit(1)
			}
			defer configFile.Close()

			for k, v := range answers {
				_, err := configFile.WriteString(fmt.Sprintf("%s: %s\n", k, v))
				if err != nil {
					fmt.Println("error writing to config file")
					fmt.Println(err)
					os.Exit(1)
				}
			}

			fmt.Println("Site created successfully")
			// display config file
			fmt.Println("Your site has been created with the following configuration:")
			for k, v := range answers {
				fmt.Printf("%s: %s\n", k, v)
			}
		}

		fmt.Println("Generating site...")

		var cfg config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Println("error unmarshalling config")
			fmt.Println(err)
			os.Exit(1)
		}

		// generate site
		if err := generateSite(cfg); err != nil {
			fmt.Println("error generating site")
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
