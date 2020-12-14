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
	Interval: "1s",
	Log:      true,
}

var configDir = "/etc/watchdog/conf.d"

func main() {

	fmt.Println("Watchdog Started")
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			Install()
		} else {
			configDir = os.Args[1]
		}
	}

	err := filepath.Walk(configDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(info.Name(), ".yml") {
				return nil
			}
			filename, _ := filepath.Abs(path)
			yamlFile, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Println(filename)
				fmt.Println(string(yamlFile))
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

	for {
		time.Sleep(1 * time.Hour)
	}
}

func Install() {
	if !gpath.IsDirExist(configDir) {
		gpath.MakePath(configDir)
	}

	res, err := exec.Command("bash", "-c", "cp "+strings.TrimRight(gpath.WorkingDir(), "/")+"/"+os.Args[0]+" /usr/bin/watchdog").Output()
	if err != nil {
		fmt.Println("Unable to install watchdog")
		fmt.Println("cp " + strings.TrimRight(gpath.WorkingDir(), "/") + "/" + os.Args[0])
		log.Fatal(err)
	}
	fmt.Println(string(res))

	f, err := os.Create("/etc/systemd/system/watchdog.service")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(`[Unit]
Description=Watchdog

[Service]
Type=simple
Restart=always
RestartSec=5s
ExecStart=/usr/bin/watchdog

[Install]
WantedBy=multi-user.target`)

	if err != nil {
		fmt.Println("Unable to install watchdog")
		log.Fatal(err)
	}

	res, err = exec.Command("bash", "-c", "systemctl daemon-reload").Output()
	if err != nil {
		fmt.Println("Unable to install watchdog")
		fmt.Println(string(res))
		log.Fatal(err)
	}
	fmt.Println(string(res))
	res, err = exec.Command("bash", "-c", "systemctl enable watchdog.service").Output()
	if err != nil {
		fmt.Println("Unable to install watchdog")
		fmt.Println(string(res))
		log.Fatal(err)
	}
	fmt.Println(string(res))
	res, err = exec.Command("bash", "-c", "systemctl start watchdog.service").Output()
	if err != nil {
		fmt.Println("Unable to install watchdog")
		fmt.Println(string(res))
		log.Fatal(err)
	}
	fmt.Println(string(res))
	fmt.Println("Installation done")
	os.Exit(0)
}

func createWatcher(config Watcher) {
	l, err := os.OpenFile(config.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	var regex = regexp.MustCompile(config.Regex)

	duration, err := time.ParseDuration(config.Interval)
	if err != nil {
		panic(err)
	}
	go func() {
		for {

			out, err := exec.Command("bash", "-c", config.Command).Output()
			if err != nil {
				if config.Log {
					l.WriteString(string(out))
				}
				log.Println(err)
				time.Sleep(duration)
				continue
			}
			fmt.Println(string(out))
			if config.Regex != "" {
				if regex.Match(out) && config.OnMatch != "" {
					exec.Command("bash", "-c", config.OnMatch).Output()
					time.Sleep(duration)
					continue
				} else {
					exec.Command("bash", "-c", config.Else).Output()
					time.Sleep(duration)
					continue
				}
			}

			if config.Contains != "" {
				if strings.Contains(string(out), config.Contains) && config.OnMatch != "" {
					exec.Command("bash", "-c", config.OnMatch).Output()
					time.Sleep(duration)
					continue
				} else {
					exec.Command("bash", "-c", config.Else).Output()
					time.Sleep(duration)
					continue
				}
			}

			time.Sleep(duration)

		}
	}()
}
