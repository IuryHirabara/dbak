package main

import (
	"database/sql"
	"dbak/src/configs"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	args := os.Args
	if len(args) < 2 || args[1] != "run" {
		showMenu()
		return
	}

	err := configs.Load()
	if err != nil {
		log.Fatalln(err)
	}

	if err = checkIfDirExists(configs.Config.DumpDir); err != nil {
		log.Fatalln(err)
	}

	connections := configs.Config.Connections

	var (
		wg    sync.WaitGroup
		limit int
	)
	for _, conn := range connections {
		limit += len(conn.Databases)
	}
	wg.Add(limit)

	dbsToExclude := strings.Split(
		configs.GetFlagValue(configs.Flags, "-ed", ""),
		",",
	)
	connsToExclude := strings.Split(
		configs.GetFlagValue(configs.Flags, "-ec", ""),
		",",
	)
	for _, conn := range connections {
		if isToExclude(connsToExclude, conn.ConnName) {
			markAsDone(len(conn.Databases), &wg)
			continue
		}

		baseStrConn := createBaseStrConn(&conn)

		for _, dbName := range conn.Databases {
			if isToExclude(dbsToExclude, dbName) {
				wg.Done()
				continue
			}

			strConn := fmt.Sprintf(baseStrConn, dbName)
			db, err := createConn(strConn)
			if err != nil {
				wg.Done()

				fmt.Println(err)
				continue
			}

			go dump(db, conn.ConnName, dbName, &wg)
		}

	}

	wg.Wait()
}

func showMenu() {
	fmt.Println(`Dbak is a tool to dump MySQL databases.

Usage:
	
        dbak run [arguments]

The arguments are:

        -cf        change config file to another. must be absolute path
        -dp        change dump dir to another. must be absolute path
        -ed        ignore databases by database name, separate by ','
        -ec        ignore connections by name, separate by ','
	
Example:
	        
        dbak run -cf=/home/user/myconf.json -ed=db1,db2

Config file example can be found on https://github.com/IuryHirabara/dbak/blob/main/config.example.json

To run Dbak, the current directory must have a config.json at root or the config file must be provided with '-cf' flag.
	`)
}

func checkIfDirExists(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("invalid dir: %s is not a valid dir", dir)
	}

	return nil
}

func createBaseStrConn(conn *configs.Connection) (baseStrConn string) {
	if conn.Pass == "" {
		baseStrConn = "%s@tcp(%s:%s)"
		baseStrConn = fmt.Sprintf(baseStrConn, conn.User, conn.Host, conn.Port)
	} else {
		baseStrConn = "%s:%s@tcp(%s:%s)"
		baseStrConn = fmt.Sprintf(baseStrConn, conn.User, conn.Pass, conn.Host, conn.Port)
	}

	return baseStrConn + "/%s?charset=utf8&parseTime=True&loc=Local"
}

func isToExclude(valuesToExclude []string, value string) bool {
	for _, valueToExclude := range valuesToExclude {
		if value == valueToExclude {
			return true
		}
	}

	return false
}

func markAsDone(qty int, wg *sync.WaitGroup) {
	for i := 0; i < qty; i++ {
		wg.Done()
	}
}

func createConn(strConn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", strConn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

func dump(db *sql.DB, connName, dbName string, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Dump of database '%s' begin\n", dbName)

	fileFormat := fmt.Sprintf("%s-%s-20060102-150405", connName, dbName)
	dumper, err := mysqldump.Register(db, configs.Config.DumpDir, fileFormat)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dumper.Close()

	resultFilename, err := dumper.Dump()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Dump of database '%s' end: %s\n", dbName, resultFilename)
}
