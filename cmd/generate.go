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
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

const (
	assetsDirName = "assets"
	postsDirName  = "posts"
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

type siteData struct {
	Config config
	Posts  []post
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

func generateSite(cfg config) error {
	siteData := siteData{Config: cfg}

	themeDir := filepath.Join("themes", siteData.Config.Theme)
	postsDir := filepath.Join(siteData.Config.ContentDir, postsDirName)
	assetsDir := filepath.Join(siteData.Config.ContentDir, assetsDirName)
	themeAssetsDir := filepath.Join(themeDir, assetsDirName)

	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		return fmt.Errorf("theme directory does not exist: %w", err)
	}

	requiredDirs := []string{
		postsDir,
		assetsDir,
		filepath.Join(siteData.Config.OutputDir, assetsDirName),
		filepath.Join(siteData.Config.OutputDir, postsDirName),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", dir, err)
			}
		}
	}

	assets, err := os.ReadDir(assetsDir)
	if err != nil {
		return err
	}
	if err := copyAssets(assetsDir, siteData.Config.OutputDir, assets); err != nil {
		return err
	}

	themeAssets, err := os.ReadDir(themeAssetsDir)
	if err != nil {
		return err
	}
	if err := copyAssets(themeAssetsDir, siteData.Config.OutputDir, themeAssets); err != nil {
		return err
	}

	if err := copyFile(filepath.Join(themeDir, "/style.css"), filepath.Join(siteData.Config.OutputDir, assetsDirName, "style.css")); err != nil {
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

		fbytes, err := os.ReadFile(filepath.Join(postsDir, file.Name()))
		if err != nil {
			return err
		}
		p, err := parseMarkdown(fbytes)
		if err != nil {
			return err
		}

		p.Link = fmt.Sprintf("%s-%s.html", p.Date, strings.ReplaceAll(p.Title, " ", "_"))

		siteData.Posts = append(siteData.Posts, p)
	}

	funcMap := template.FuncMap{
		"now": time.Now,
		"hasCover": func(p post) bool {
			return p.CoverImg != ""
		},
		"sortByDate": func(posts []post) []post {
			sort.Slice(posts, func(i, j int) bool {
				return posts[i].Date < posts[j].Date
			})
			return posts
		},
	}

	// create post html files in posts directory
	for _, p := range siteData.Posts {
		file, err := os.Create(filepath.Join(siteData.Config.OutputDir, postsDirName, p.Link))
		if err != nil {
			return err
		}

		tmpl := template.Must(template.New("postHTML").Funcs(funcMap).ParseGlob(filepath.Join(themeDir, "*.html")))

		if err := tmpl.ExecuteTemplate(file, "postHTML", struct {
			Post   post
			Config config
		}{
			Post:   p,
			Config: siteData.Config,
		}); err != nil {
			return err
		}
	}

	tmpl := template.Must(template.New("baseHTML").Funcs(funcMap).ParseGlob(filepath.Join(themeDir, "*.html")))
	// write site to output directory as index.html

	file, err := os.Create(filepath.Join(siteData.Config.OutputDir, "index.html"))
	if err != nil {
		return err
	}

	if err := tmpl.ExecuteTemplate(file, "baseHTML", siteData); err != nil {
		return err
	}

	return nil
}

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

func copyAssets(assetDir, outputDir string, assets []os.DirEntry) error {
	for _, asset := range assets {
		if asset.IsDir() {
			continue
		}

		fmt.Println("copying asset: ", asset.Name())
		if err := copyFile(filepath.Join(assetDir, asset.Name()), filepath.Join(outputDir, assetsDirName, asset.Name())); err != nil {
			return err
		}

	}
	return nil
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

func init() {
	rootCmd.AddCommand(generateCmd)
}
