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
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := generateSite(); err != nil {
			fmt.Println("error generating site")
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func generateSite() error {
	// read markdown from folder
	// check if markdown folder exists
	if _, err := os.Stat("markdown"); os.IsNotExist(err) {
		return fmt.Errorf("markdown folder does not exist: %w", err)
	}
	// check if html folder exists
	if _, err := os.Stat("html"); os.IsNotExist(err) {
		// create html folder
		if err := os.Mkdir("html", 0755); err != nil {
			return err
		}
	}
	// check if markdown folder contains markdown files
	mdFiles, err := os.ReadDir("markdown")
	if err != nil {
		return err
	}
	if len(mdFiles) == 0 {
		return errors.New("no markdown files found in markdown folder")
	}

	toGenerate := []string{}
	for _, file := range mdFiles {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".md" && filepath.Ext(file.Name()) != ".markdown" {
			continue
		}
		toGenerate = append(toGenerate, file.Name())
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("post").ParseFiles(wd + "/templates/post.html"))

	// parse markdown to html
	// TODO: speed this up using goroutines
	for _, file := range toGenerate {
		// TODO: read file line by line
		content, err := os.ReadFile("markdown/" + file)
		if err != nil {
			return err
		}
		post, err := parseMarkdown(content)
		if err != nil {
			return err
		}
		str, err := post.generateHTML(tmpl)
		if err != nil {
			return err
		}

		// write html to file
		err = os.WriteFile("html/"+strings.TrimSuffix(file, ".md")+".html", []byte(str), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

type post struct {
	Title    string
	Date     string
	Markdown string
	Content  template.HTML
}

func (p *post) generateHTML(tmpl *template.Template) (string, error) {
	// convert markdown to html

	content := string(blackfriday.Run([]byte(p.Markdown)))
	p.Content = template.HTML(content)

	// execute template into string
	var strBuffer bytes.Buffer
	if err := tmpl.Execute(&strBuffer, p); err != nil {
		return "", err
	}

	result := strBuffer.String()

	return result, nil
}

func removeQuotes(str string) string {
	return strings.ReplaceAll(str, "\"", "")
}

// parseMarkdown reads a markdown file and returns a post struct
// containing the metadata and content
// metadata is expected to be in the format:
// ---
// title: "title"
// date: "YYYY-MM-DD"
// ---
// content
func parseMarkdown(content []byte) (post, error) {
	p := post{}
	lines := strings.Split(string(content), "\n")
	readMetadata := false
	contentLine := 0
	for i, line := range lines {
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
				return post{}, errors.New("metadata line does not contain a colon")
			}
			// trim the whitespace from the parts
			metaArr[0] = strings.TrimSpace(metaArr[0])
			metaArr[1] = strings.TrimSpace(metaArr[1])

			switch metaArr[0] {
			case "title":
				p.Title = removeQuotes(metaArr[1])
			case "date":
				p.Date = removeQuotes(metaArr[1])
			}
		}
	}

	if !readMetadata {
		return post{}, errors.New("metadata not found")
	}

	// the rest of the file is the content
	p.Markdown = strings.Join(lines[contentLine+1:], "\n")

	return p, nil
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
