package writebuffer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var P *pgxpool.Pool

func setup(b *testing.B) {

	setupTable := `
	CREATE TABLE test_measurements(uid UUID ,time TIMESTAMPTZ,value real,metadata JSONB);
	CREATE INDEX pk_time_index ON test_measurements(uid,time);
	`

	connStr := "postgresql://postgres:postgres@db:5432/postgres"
	pconf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}
	P, err = pgxpool.ConnectConfig(context.Background(), pconf)
	if err != nil {
		panic("Cant initialize pgxpool")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := P.Acquire(ctx)
	if err != nil {
		panic(fmt.Sprintf("Cant acquire connection from pool with err = %v", err))
	}
	conn.Exec(ctx, setupTable)

}

func teardown(b *testing.B) {

	removeTable := `
	DROP TABLE IF EXISTS test_measurements;
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn, err := P.Acquire(ctx)
	if err != nil {
		panic(fmt.Sprintf("Cant acquire connection from pool with err = %v in teardown", err))
	}
	_, err = conn.Exec(ctx, removeTable)
	if err != nil {
		b.Log(err)
	}
	conn.Release()
}

func countRows(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn, err := P.Acquire(ctx)
	if err != nil {
		panic(fmt.Sprintf("Cant acquire connection from pool with err = %v in teardown", err))
	}

	var num int
	record := conn.QueryRow(ctx, "SELECT COUNT(uid) from test_measurements;")
	conn.Release()
	err = record.Scan(&num)
	if err != nil {
		b.Fatal(err)
	}
	b.Log(fmt.Sprintf("Number of rows inserted is %v", num))
}

type M struct {
	Uid      uuid.UUID
	Time     time.Time
	Value    float32
	Metadata map[string]interface{}
}

func BenchmarkDbWriteThroughput(b *testing.B) {

	numRecords := 1024
	setup(b)
	b.Run("SequentialRowInsert", func(b *testing.B) {
		SequentialRowInsert(numRecords, b)
	})
	countRows(b)
	teardown(b)
	setup(b)
	b.Run("BatchRowInsert", func(b *testing.B) {
		BatchRowInsert(numRecords, b)
	})
	countRows(b)
	teardown(b)
	setup(b)
	b.Run("SingleQueryInsert", func(b *testing.B) {
		SingleQueryInsert(numRecords, b)
	})
	countRows(b)
	teardown(b)

}

func SequentialRowInsert(numRecords int, b *testing.B) {
	insertStmt := `INSERT INTO test_measurements VALUES ($1,$2,$3,$4);`
	data := make([]M, numRecords)
	for i := 0; i < numRecords; i++ {
		data = append(data, M{Uid: uuid.New(), Time: time.Now(), Value: rand.Float32(), Metadata: map[string]interface{}{fmt.Sprintf("%v", i): i}})
	}
	for i := 0; i < b.N; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		conn, _ := P.Acquire(ctx)

		for _, r := range data {
			conn.Exec(ctx, insertStmt, r.Uid, r.Time, r.Value, r.Metadata)
		}
		conn.Release()

	}
}

func BatchRowInsert(numRecords int, b *testing.B) {
	insertStmt := `INSERT INTO test_measurements VALUES ($1,$2,$3,$4);`
	data := make([]M, 0, 1024)

	for i := 0; i < numRecords; i++ {
		data = append(data, M{Uid: uuid.New(), Time: time.Now(), Value: rand.Float32(), Metadata: map[string]interface{}{fmt.Sprintf("%v", i): i}})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := &pgx.Batch{}

		for i := range data {
			var r M
			r = data[i]
			batch.Queue(insertStmt, r.Uid, r.Time, r.Value, r.Metadata)
		}
		batch.Queue("select count(*) from test_measurements;")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		br := P.SendBatch(ctx, batch)
		for i := 0; i < len(data); i++ {
			_, err := br.Exec()
			if err != nil {
				b.Fatal(err)
			}

		}
		var inserted int
		br.QueryRow().Scan(&inserted)
		br.Close()
	}
}

func SingleQueryInsert(numRecords int, b *testing.B) {

	data := make([]M, 0, numRecords)
	for i := 0; i < numRecords; i++ {
		data = append(data, M{Uid: uuid.New(), Time: time.Now(), Value: rand.Float32(), Metadata: map[string]interface{}{fmt.Sprintf("%v", i): i}})
	}

	valueArgs := make([]interface{}, 0, len(data)*3)
	for i := range data {
		// var r M
		r := data[i]
		valueArgs = append(valueArgs, r.Uid, r.Time, r.Value, r.Metadata)
	}
	sqlQuery := BuildQueryString("test_measurements", len(data), 4)
	b.ResetTimer()
	var count int
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := P.Exec(ctx, sqlQuery, valueArgs...)
		if err != nil {
			b.Fatal(err)
		}
		err = P.QueryRow(context.Background(), "select count(*) from test_measurements;").Scan(&count)
		if err != nil {
			b.Fatal(err)
		}
		b.Logf("Inserted %v", count)
	}

}
