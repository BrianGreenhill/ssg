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
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type config struct {
	Theme       string
	ContentDir  string
	OutputDir   string
	Title       string
	Author      string
	AuthorImg   string
	Description string
	Github      string
	Linkedin    string
	Email       string
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a static site from markdown files",
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Println("error unmarshalling config")
			fmt.Println(err)
			os.Exit(1)
		}
		if err := generateSite(cfg); err != nil {
			fmt.Println("error generating site")
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func copyFile(src, dst string) error {
	// check if file to copy exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("file to copy does not exist: %w", err)
	}

	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening file to copy: %w", err)
	}
	defer r.Close()

	// if the destination file already exists, remove it
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("error removing destination file: %w", err)
		}
	}

	// create destination file
	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer w.Close()

	if _, err := w.ReadFrom(r); err != nil {
		return err
	}

	return nil
}

func (c *config) getThemeDir() string {
	return filepath.Join("themes", c.Theme)
}

func (c *config) getPostsDir() string {
	return filepath.Join(c.ContentDir, "posts")
}

func (c *config) getAssetsDir() string {
	return filepath.Join(c.ContentDir, "assets")
}

func generateSite(cfg config) error {
	s := site{
		Config: cfg,
	}

	themeDir := cfg.getThemeDir()
	postsDir := cfg.getPostsDir()
	assetsDir := cfg.getAssetsDir()
	themeStyle := themeDir + "/style.css"

	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		return fmt.Errorf("theme directory does not exist: %w", err)
	}

	if _, err := os.Stat(cfg.ContentDir); os.IsNotExist(err) {
		return fmt.Errorf("input folder does not exist: %w", err)
	}
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		if err := os.Mkdir(cfg.OutputDir, 0755); err != nil {
			return err
		}
		if err := os.Mkdir(cfg.OutputDir+"/assets", 0755); err != nil {
			return err
		}
		if err := os.Mkdir(cfg.OutputDir+"/posts", 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		if err := os.Mkdir(postsDir, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		if err := os.Mkdir(assetsDir, 0755); err != nil {
			return err
		}
	}

	assets, err := os.ReadDir(assetsDir)
	if err != nil {
		return err
	}
	for _, asset := range assets {
		if asset.IsDir() {
			continue
		}
		fmt.Println("copying asset: ", asset.Name())
		if err := copyFile(assetsDir+"/"+asset.Name(), cfg.OutputDir+"/assets/"+asset.Name()); err != nil {
			return err
		}
	}

	// move style.css from template to output directory
	if err := copyFile(themeStyle, cfg.OutputDir+"/assets/style.css"); err != nil {
		return err
	}

	mdFiles, err := os.ReadDir(postsDir)
	if err != nil {
		return err
	}
	if len(mdFiles) == 0 {
		fmt.Printf("warning: no markdown files found in %s folder\n", postsDir)
	}

	for _, file := range mdFiles {
		if !strings.HasSuffix(file.Name(), ".md") && !strings.HasSuffix(file.Name(), ".markdown") {
			continue
		}

		// build post object from markdown file
		// read file contents into memory
		fbytes, err := os.ReadFile(postsDir + "/" + file.Name())
		if err != nil {
			return err
		}
		p, err := parseMarkdown(fbytes)
		if err != nil {
			return err
		}

		p.Link = p.getLink()

		// add post to site struct
		s.Posts = append(s.Posts, p)
	}

	funcMap := template.FuncMap{
		"now": time.Now,
		"hasCover": func(p post) bool {
			return p.CoverImg != ""
		},
		"sortByDate": func(posts []post) []post {
			// sort posts by date
			// newest first
			for i := 0; i < len(posts); i++ {
				for j := i + 1; j < len(posts); j++ {
					if posts[i].Date < posts[j].Date {
						posts[i], posts[j] = posts[j], posts[i]
					}
				}
			}
			return posts
		},
	}

	// create post html files in posts directory
	for _, p := range s.Posts {
		file, err := os.Create(cfg.OutputDir + "/posts/" + p.getFileName())
		if err != nil {
			return err
		}

		tmpl := template.Must(template.New("postHTML").Funcs(funcMap).ParseGlob(filepath.Join(themeDir, "*.html")))

		if err := tmpl.ExecuteTemplate(file, "postHTML", struct {
			Post   post
			Config config
		}{
			Post:   p,
			Config: cfg,
		}); err != nil {
			return err
		}
	}

	tmpl := template.Must(template.New("baseHTML").Funcs(funcMap).ParseGlob(filepath.Join(themeDir, "*.html")))
	// write site to output directory as index.html

	file, err := os.Create(cfg.OutputDir + "/index.html")
	if err != nil {
		return err
	}

	if err := tmpl.ExecuteTemplate(file, "baseHTML", s); err != nil {
		return err
	}

	return nil
}

type site struct {
	Config config
	Posts  []post
}

type post struct {
	Title       string        `yaml:"title"`
	Author      string        `yaml:"author"`
	Description string        `yaml:"description"`
	AuthorImg   string        `yaml:"author_image"`
	CoverImg    string        `yaml:"cover_image"`
	Date        string        `yaml:"date"`
	Link        string        `yaml:"link"`
	Content     template.HTML `yaml:"-"`
}

func (p *post) getFileName() string {
	return p.Date + "-" + strings.ReplaceAll(p.Title, " ", "_") + ".html"
}

// parseMarkdown reads a markdown file and returns a post struct
// containing the metadata and content
// metadata is expected to be in the format:
// ---
// title: "title"
// date: "YYYY-MM-DD"
// author: "author"
// description: "description"
// author_image: "http://example.com/image.jpg"
// cover_image: "http://example.com/image.jpg"
// ---
// content
func parseMarkdown(content []byte) (post, error) {
	ctx := parser.NewContext()
	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	md.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	d := frontmatter.Get(ctx)
	if d == nil {
		return post{}, fmt.Errorf("no frontmatter found")
	}

	var p post
	if err := d.Decode(&p); err != nil {
		return post{}, err
	}

	// the rest of the file is the content
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return post{}, err
	}
	p.Content = template.HTML(buf.String())

	return p, nil
}

func (p *post) getLink() string {
	return p.Date + "-" + strings.ReplaceAll(p.Title, " ", "_") + ".html"
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
