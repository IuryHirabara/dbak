package config

type Config struct {
	ConnName string `json:"connName"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	DB       string `json:"db"`
	DBType   string `json:"dbType"`
	DumpDir  string `json:"dumpDir"`
}
