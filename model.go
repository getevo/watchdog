package main

type Watcher struct {
	Name     string `yaml:"name"`
	Command  string `yaml:"command"`
	Regex    string `yaml:"regex"`
	Contains string `yaml:"contains"`
	OnMatch  string `yaml:"onmatch"`
	Else     string `yaml:"else"`
	Interval string `yaml:"interval"`
	Log      bool   `yaml:"log"`
	LogPath  string `yaml:"path"`
}
