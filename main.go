package main

import (
	"database/sql"
	"dbak/src/config.go"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	filePath := "./config.json"
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("O arquivo %s de configurações não existe", filePath)
	}

	var configs []config.Config
	if err = json.Unmarshal(configBytes, &configs); err != nil {
		log.Fatal("Não foi possível realizar o unmarshal das configurações")
	}

	for _, config := range configs {
		var stringConn string
		if config.Pass == "" {
			stringConn = fmt.Sprintf(
				"%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				config.User,
				config.Host,
				config.Port,
				config.DB,
			)
		} else {
			stringConn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
				config.User,
				config.Pass,
				config.Host,
				config.Port,
				config.DB,
			)
		}

		db, err := sql.Open("mysql", stringConn)
		if err != nil {
			printAndCloseDBConn(
				fmt.Sprintf("Não foi possível realizar a conexão inicial com o banco de dados: %s", config.ConnName),
				db,
			)
			continue
		}

		if err = db.Ping(); err != nil {
			printAndCloseDBConn(
				fmt.Sprintf("Não foi possível estabelecer uma conexão com o banco de dados: %s", config.ConnName),
				db,
			)
			continue
		}

		dumpFilename := fmt.Sprintf("%s-20060102T150405", config.DB)

		dumper, err := mysqldump.Register(db, config.DumpDir, dumpFilename)
		if err != nil {
			printAndCloseDBConn("Não foi possível realizar o dump do banco de dados: "+config.ConnName, db)
			dumper.Close()
			continue
		}

		resultFilename, err := dumper.Dump()
		if err != nil {
			printAndCloseDBConn("Error dumping: "+err.Error(), db)
			dumper.Close()
			return
		}
		fmt.Printf("File is saved to %s", resultFilename)

		dumper.Close()
		db.Close()
	}
}

func printAndCloseDBConn(message string, db *sql.DB) {
	fmt.Println(message)
	db.Close()
}
