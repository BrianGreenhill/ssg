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
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// read markdown from folder
		// check if markdown folder exists
		if _, err := os.Stat("markdown"); os.IsNotExist(err) {
			fmt.Println("markdown folder does not exist")
			return
		}
		// check if html folder exists
		if _, err := os.Stat("html"); os.IsNotExist(err) {
			fmt.Println("html folder does not exist")
			return
		}
		// check if markdown folder contains markdown files
		mdFiles, err := os.ReadDir("markdown")
		if err != nil {
			fmt.Println("error reading markdown folder")
			return
		}
		if len(mdFiles) == 0 {
			fmt.Println("no markdown files found")
			return
		}

		toGenerate := []string{}
		for _, file := range mdFiles {
			if file.IsDir() {
				continue
			}
			if file.Name()[len(file.Name())-3:] != ".md" {
				continue
			}
			toGenerate = append(toGenerate, file.Name())
		}

		// parse markdown to html
		// TODO: speed this up using goroutines
		for _, file := range toGenerate {
			fmt.Println("parsing", file)
			content, err := os.ReadFile("markdown/" + file)
			if err != nil {
				fmt.Println("error reading file", file)
				fmt.Println(err)
				return
			}
			post := parseMarkdown(content)
			post.generateHTML()

			// write html to file
			err = os.WriteFile("html/"+strings.TrimSuffix(file, ".md")+".html", []byte(post.HTML), 0644)
			if err != nil {
				fmt.Println("error writing file", file)
				fmt.Println(err)
				return
			}
		}
	},
}

type post struct {
	Title   string
	Date    string
	Content string
	HTML    string
}

func (p *post) generateHTML() {
	// convert markdown to html
	p.Content = string(blackfriday.Run([]byte(p.Content)))
	// for now, just wrap the content in a div
	p.HTML = "<html><head><title>" + p.Title + "</title></head><body><div>" + p.Date + "</div><div>" + p.Content + "</div></body></html>"
}

// parseMarkdown reads a markdown file and returns a post struct
// containing the metadata and content
// metadata is expected to be in the format:
// ---
// title: "title"
// date: "YYYY-MM-DD"
// ---
// content
func parseMarkdown(content []byte) post {
	p := post{}
	contentStr := string(content)
	// read the file line by line
	// if the line is "---" then we are at the metadata
	// if the line is "---" and we have already read metadata, then we are at the content
	readMetadata := false
	contentLine := 0
	for i, line := range strings.Split(contentStr, "\n") {
		if line == "---" {
			if readMetadata {
				contentLine = i
				break
			}
			readMetadata = true
			continue
		}
		if readMetadata {
			// split the line by ":"
			metaArr := strings.Split(line, ":")
			if len(metaArr) != 2 {
				continue
			}
			// trim the whitespace from the parts
			metaArr[0] = strings.TrimSpace(metaArr[0])
			metaArr[1] = strings.TrimSpace(metaArr[1])

			switch metaArr[0] {
			case "title":
				p.Title = metaArr[1]
			case "date":
				p.Date = metaArr[1]
			}
		}
	}

	// the rest of the file is the content
	p.Content = strings.Join(strings.Split(contentStr, "\n")[contentLine+1:], "\n")

	return p
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
