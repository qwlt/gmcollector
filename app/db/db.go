package db

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	cfg "github.com/qwlt/gmcollector/app/config"
	"github.com/spf13/viper"
)

var DB *pgxpool.Pool

type DbConfig struct {
	User     string `mapstruct:"user"`
	Password string `mapsctruct:"password"`
	Dbname   string `mapstruct:"dbname"`
	Host     string `mapsruct:"host"`
	Port     string `mapstruct:"port"`
}

func (c *DbConfig) GetConnString() (string, error) {
	allowedFields := []string{"user", "password", "host", "port", "dbname", "sslmode", "pool_max_conn"}
	var connStr strings.Builder
	val := reflect.ValueOf(c).Elem()
	for i := 0; i < val.NumField(); i++ {
		fieldNameLower := strings.ToLower(val.Type().Field(i).Name)
		fieldValue := val.Field(i).Interface()

		if contains(allowedFields, fieldNameLower) && fieldValue != "" {
			connStr.WriteString(fmt.Sprintf("%v=%v ", fieldNameLower, val.Field(i).Interface()))
		}
	}
	s := strings.Trim(connStr.String(), " ")
	return s, nil
}

func contains(sl []string, s string) bool {
	for _, v := range sl {
		if s == v {
			return true
		}
	}
	return false
}

func InitPgPool(conf *DbConfig) error {
	connStr, err := conf.GetConnString()
	if err != nil {
		log.Fatal(err)
	}
	c, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), c)
	if err != nil {
		log.Fatal(err)
	}
	DB = pool
	return nil
}

func GetDB() *pgxpool.Pool {
	if DB == nil {

		dbConf := DbConfig{}
		if viper.IsSet("db") {
			err := viper.UnmarshalKey("db", &dbConf)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(&cfg.ViperKeyNotFoundError{Key: "db", Config: viper.ConfigFileUsed()})
		}

		err := InitPgPool(&dbConf)
		if err != nil {
			log.Fatal(err)
		}
	}
	return DB
}
