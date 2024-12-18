package configs

import (
	"encoding/json"
	"os"
	"strings"
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
	Flags  map[int]map[string]string
)

func Load() error {
	Flags = getFlags()

	filename := GetFlagValue(Flags, "-cf", "./config.json")
	fileInBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileInBytes, &Config)
	if err != nil {
		return err
	}

	dumpDir := GetFlagValue(Flags, "-dp", "")
	if dumpDir != "" {
		Config.DumpDir = dumpDir
	}

	return nil
}

func getFlags() (flags map[int]map[string]string) {
	flags = map[int]map[string]string{}
	args := os.Args[2:]
	for i, arg := range args {
		flagValue := strings.Split(arg, "=")
		if len(flagValue) < 2 {
			continue
		}

		flag := flagValue[0]
		value := flagValue[1]

		if strings.Trim(flag, " ") == "" || strings.Trim(value, "") == "" {
			continue
		}

		flags[i] = map[string]string{
			"flag":  flag,
			"value": value,
		}
	}

	return flags
}

func GetFlagValue(flags map[int]map[string]string, flag, defaultValue string) string {
	for _, arg := range flags {
		switch arg["flag"] {
		case flag:
			return arg["value"]
		}
	}

	return defaultValue
}
