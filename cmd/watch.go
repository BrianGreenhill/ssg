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
	"net/http"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.Unmarshal(&cfg); err != nil {
			return fmt.Errorf("unable to decode into struct, %v", err)
		}

		// generate the site initially
		if err := generateSite(cfg); err != nil {
			return fmt.Errorf("error generating site: %v", err)
		}

		go func() {
			if err := startServer(); err != nil {
				fmt.Printf("error starting server: %v", err)
			}
		}()

		return watchForChanges()
	},
}

func watchForChanges() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}
	defer watcher.Close()

	themeDir := filepath.Join("themes", cfg.Theme)
	content := filepath.Join(cfg.ContentDir, postsDirName)
	assets := filepath.Join(cfg.ContentDir, assetsDirName)
	for _, dir := range []string{themeDir, content, assets} {
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("error watching directory: %v", err)
		}

	}

	debouncedChan := debounceEvents(500*time.Millisecond, watcher.Events)
	for {
		select {
		case event, ok := <-debouncedChan:
			if !ok {
				return nil
			}
			fmt.Printf("Detected change in %s\n", event.Name)
			if shouldRegenerate(event.Name) {
				if err := generateSite(cfg); err != nil {
					fmt.Printf("error generating site: %v\n", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("error watching directory: %v", err)
		}
	}
}

func shouldRegenerate(filename string) bool {
	switch filepath.Ext(filename) {
	case ".md", ".html", ".css", ".markdown":
		return true
	default:
		return false
	}
}

func debounceEvents(interval time.Duration, eventChan <-chan fsnotify.Event) <-chan fsnotify.Event {
	debouncedChan := make(chan fsnotify.Event)
	go func() {
		var lastEvent fsnotify.Event
		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					return
				}
				// check if event is a file change
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					lastEvent = event
				}
			case <-time.After(interval):
				if lastEvent.Name != "" {
					debouncedChan <- lastEvent
					lastEvent = fsnotify.Event{}
				}
			}
		}
	}()
	return debouncedChan
}

func startServer() error {
	mux := http.NewServeMux()

	// serve files form configured output directory
	dir := viper.GetString("outputDir")
	mux.Handle("/", http.FileServer(http.Dir(dir)))
	fmt.Println("Serving files from", dir)
	fmt.Println("Starting server on :8080")
	fmt.Println()
	fmt.Println("Visit http://localhost:8080 to view your site")
	return http.ListenAndServe(":8080", mux)
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
