package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestDbConfigParser(t *testing.T) {
	var conf DbConfig
	conf.User = "username"
	conf.Password = "password"
	conf.Host = "host"
	conf.Port = "5432"
	conf.Dbname = "postgres"
	defaultConnStr := "user=username password=password dbname=postgres host=host port=5432"
	connStr, err := conf.GetConnString()
	require.NoError(t, err)
	require.Equal(t, connStr, defaultConnStr)

	var conf2 DbConfig
	conf2.User = "username"
	connStr, err = conf2.GetConnString()
	require.NoError(t, err)
	require.Equal(t, connStr, "user=username")
}

func TestDbconnection(t *testing.T) {
	dbconf := DbConfig{
		User:     "postgres",
		Password: "password",
		Dbname:   "postgres",
		Host:     "db",
		Port:     "5432",
	}
	connStr, err := dbconf.GetConnString()
	require.NoError(t, err)

	pconf, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)
	p, err := pgxpool.ConnectConfig(context.Background(), pconf)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	err = p.Ping(ctx)
	require.NoError(t, err)
}
