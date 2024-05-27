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
	"os"
	"strings"
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
	Run: func(cmd *cobra.Command, args []string) {
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Println("error unmarshalling config")
			fmt.Println(err)
			os.Exit(1)
		}

		go func() {
			if err := startServer(); err != nil {
				fmt.Println("error starting server")
				fmt.Println(err)
			}
		}()

		breakpoint := make(chan struct{})
		eventChan := make(chan fsnotify.Event)

		go watchForChanges(breakpoint, eventChan)

		for {
			<-breakpoint
			event := <-eventChan
			fmt.Println("Detected change in", event.Name)

			if strings.HasSuffix(event.Name, ".md") || strings.HasSuffix(event.Name, ".html") || strings.HasSuffix(event.Name, ".css") || strings.HasSuffix(event.Name, ".markdown") {
				if err := generateSite(cfg); err != nil {
					fmt.Println("error generating site")
					fmt.Println(err)
				}
			}
		}
	},
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

func watchForChanges(breakpoint chan struct{}, eventChan chan<- fsnotify.Event) {
	// watch for changes in input directory
	// if changes are detected, send signal to breakpoint
	// to regenerate site

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("error creating watcher")
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	debouncedChan := debounceEvents(500*time.Millisecond, watcher.Events)
	go func() {
		for {
			select {
			case event, ok := <-debouncedChan:
				if !ok {
					return
				}
				breakpoint <- struct{}{}
				eventChan <- event
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error watching directory")
				fmt.Println(err)
				return
			}
		}
	}()

	watchDirs := []string{cfg.getPostsDir(), cfg.getAssetsDir(), cfg.getThemeDir()}
	for _, dir := range watchDirs {
		if err = watcher.Add(dir); err != nil {
			fmt.Println("error watching directory")
			fmt.Println(err)
			return
		}
	}
	<-done
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
