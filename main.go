package main

import (
	"fmt"
	"github.com/getevo/evo/lib/gpath"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var defaults = Watcher{
	Interval: 1 * time.Second,
	Log:      true,
}

var configDir = "/etc/watchdog/conf.d"

func main() {

	fmt.Println()

	Install()
	err := filepath.Walk(configDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			filename, _ := filepath.Abs(path)
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			watcher := defaults
			err = yaml.Unmarshal(yamlFile, &watcher)
			if err != nil {
				panic(err)
			}
			if watcher.Log && watcher.LogPath == "" {
				watcher.LogPath = path + ".log"
			}
			if watcher.Name == "" {
				watcher.Name = filepath.Base(path)
			}
			fmt.Printf("Watcher %s created \n", watcher.Name)
			createWatcher(watcher)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func Install() {
	if !gpath.IsDirExist(configDir) {
		gpath.MakePath(configDir)
	}
}

func createWatcher(config Watcher) {
	l, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	var regexes = []*regexp.Regexp{}
	for _, item := range config.Regex {
		regexes = append(regexes, regexp.MustCompile(item))
	}
	go func() {
		for {
			for _, command := range config.Command {
				out, err := exec.Command("bash", "-c", command).Output()
				if err != nil {
					if config.Log {
						l.WriteString(string(out))
					}
					log.Println(err)
					continue
				}

				fmt.Println(string(out))
				for _, regex := range regexes {
					if regex.Match(out) && config.OnMatch != "" {
						exec.Command("bash", "-c", config.OnMatch).Output()
						continue
					} else {
						exec.Command("bash", "-c", config.Else).Output()
						continue
					}
				}

				for _, item := range config.Contains {
					if strings.Contains(string(out), item) && config.OnMatch != "" {
						exec.Command("bash", "-c", config.OnMatch).Output()
						continue
					} else {
						exec.Command("bash", "-c", config.Else).Output()
						continue
					}
				}

				time.Sleep(config.Interval)

			}
		}
	}()
}
