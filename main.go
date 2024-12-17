package main

import (
	"database/sql"
	"dbak/src/configs"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
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

	for _, conn := range connections {
		baseStrConn := createBaseStrConn(&conn)

		for _, dbName := range conn.Databases {
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

	fmt.Printf("Dump do banco '%s' iniciado\n", dbName)

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

	fmt.Printf("Dump do banco '%s' concluído: %s\n", dbName, resultFilename)
}
