package checker

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/asciimoo/coa/event"
	"github.com/asciimoo/coa/notification"

	"github.com/fsnotify/fsnotify"
)

type Checker struct {
	Paths []string
	Exclude string `yaml:",omitempty"`
	Command string
	Name string
	ProjectName string
	running bool
	stop chan bool
}

func (c *Checker) Start(cwd string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Inotify error:", err.Error())
		return
	}
	defer watcher.Close()

	dirsToWatch := make(map[string]bool)

	for i, p := range c.Paths {
		if !strings.HasPrefix(p, "/") {
			p = path.Join(cwd, p)
			c.Paths[i] = p
		}

		if pinfo, err := os.Stat(filepath.Dir(p)); err == nil && pinfo.IsDir() {
			dirsToWatch[filepath.Dir(p)] = true
		}

		files, err := filepath.Glob(p)
		if err != nil {
			log.Println("Invalid glob expression:", p, err.Error())
			return
		}
		for _, f := range files {
			finfo, err := os.Stat(f)
			if err != nil {
				log.Println("Cannot stat file:", err.Error(), f)
				return
			}
			if finfo.IsDir() {
				dirsToWatch[f] = true
			} else {
				dirsToWatch[filepath.Dir(f)] = true
			}
		}
	}

	if len(dirsToWatch) == 0 {
		log.Println("No files to watch in checker", c.Name, c.ProjectName)
		return
	}

	for d, _ := range dirsToWatch {
		if err := watcher.Add(d); err != nil {
			log.Println("Inotify error:", err.Error(), d)
			return
		}
	}

	activeFiles := make(map[string]bool)
	startCheck := make(chan bool)
	c.stop = make(chan bool)

	for {
		select {
		case ev := <-watcher.Events:
			// inotify event handler
			if ev.Op&fsnotify.Remove == fsnotify.Remove {
				watcher.Remove(ev.Name)
				break
			}

			if ev.Op&fsnotify.Write != fsnotify.Write && ev.Op&fsnotify.Create != fsnotify.Create {
				break
			}

			validFile := false
			for _, pat := range c.Paths {
				if match, err := filepath.Match(pat, ev.Name); err == nil && match {
					validFile = true
					break
				}
			}
			if !validFile {
				break
			}

			if ev.Op&fsnotify.Create == fsnotify.Create {
				if err := watcher.Add(ev.Name); err != nil {
					log.Println("Inotify error:", err.Error(), ev.Name)
					return
				}
			}

			activeFiles[ev.Name] = true
			if !c.running {
				c.running = true
				go func() {
					time.Sleep(time.Microsecond * 200)
					startCheck <- true
				}()
			}
		case err := <-watcher.Errors:
			log.Println("Inotify error:", err)
		case <- startCheck:
			c.running = true
			cmd := exec.Command("sh", "-c", c.Command)
			cmd.Dir = cwd
			go func(cmd *exec.Cmd, activeFiles *map[string]bool, c *Checker) {
				// TODO make fileArgs available from the check command
				// fileArgs := *activeFiles
				*activeFiles = make(map[string]bool)
				out, err := cmd.CombinedOutput()
				if err != nil {
					notification.Send(c.CreateEvent(event.Fail, "Check command returned with error: " + string(out)))
				} else {
					if cmd.ProcessState.Success() {
						notification.Send(c.CreateEvent(event.Pass, string(out)))
					} else {
						notification.Send(c.CreateEvent(event.Fail, string(out)))
					}
				}
				c.running = false
				if len(*activeFiles) > 0 {
					go func() { startCheck <- true }()
				}
			}(cmd, &activeFiles, c)
		case <- c.stop:
			c.running = false
			return
		}
	}
}

func (c *Checker) CreateEvent(t event.EventType, message string) *event.Event {
	return &event.Event{
		Type: t,
		Message: message,
		ProjectName: c.ProjectName,
		CheckerName: c.Name,
	}
}

func (c *Checker) Stop() {
	c.stop <- true
}
