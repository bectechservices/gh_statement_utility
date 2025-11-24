package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	b64 "encoding/base64"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v5"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectMSSQL1() error {
	var err error
	env := envy.Get("GO_ENV", "development")
	dbUrl := envy.Get("GH_DB_URL", "")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Println(err)
	}

	encryptedPassword, err := b64.StdEncoding.DecodeString(string(ReadPasswordFromFile()))

	if err != nil {
		log.Println(err)
	}
	password := DecryptPassword(encryptedPassword)

	GormDB, err = gorm.Open(sqlserver.Open(strings.ReplaceAll(dbUrl, "__password__", string(password))), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	/// Logger: logger.Default.LogMode(logger.Info)
	if err != nil {
		log.Println(err)
	}
	return err
}

// ConnectMSSQL initializes the MS SQL Server connection
func ConnectMSSQL() error {
	server := "127.0.0.1"
	port := 1433
	user := "sladm"
	password := "password"
	database := "SLDATA"

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, password, server, port, database,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return fmt.Errorf("failed to open MSSQL connection: %w", err)
	}

	// Verify connection
	if err = db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("failed to ping MSSQL: %w", err)
	}

	//DB :=

	// DB = db // store globally for future queries

	log.Println("Connected to MSSQL database successfully")
	return nil
}
