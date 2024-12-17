package configs

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Connections []Connection `json:"connections"`
	DumpDir     string       `json:"dumpDir"`
}

type Connection struct {
	ConnName  string   `json:"connName"`
	Host      string   `json:"host"`
	Port      string   `json:"port"`
	User      string   `json:"user"`
	Pass      string   `json:"pass"`
	DBType    string   `json:"dbType"`
	Databases []string `json:"databases"`
}

var (
	Config Configuration
)

func Load() error {
	filename := "./config.json"
	fileInBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileInBytes, &Config)
	if err != nil {
		return err
	}

	return nil
}
