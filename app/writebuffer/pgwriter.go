package writebuffer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/qwlt/gmcollector/app/models"
)

// PGWriter - wrapper around postgres connections pool
type PGWriter struct {
	ConnPool  *pgxpool.Pool
	TableName string
}

// Write - inserts batch of data into database or in case of any errors
// reject batch entierly
func (pg *PGWriter) Write(data []models.Model) error {


	if len(data) == 0 {
		return nil
	}
	flatData := make([]interface{}, 0, 2048)
	for i := range data {

		flatData = append(flatData, data[i].Flatten()...)
	}
	query := BuildQueryString(pg.TableName, len(data), len(data[1].Flatten()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	tx, err := pg.ConnPool.Begin(ctx)
	if err != nil {
		return err
	}

	com, err := tx.Exec(ctx, query, flatData...)
	if err != nil {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		txErr := tx.Rollback(rollbackCtx)
		log.Println(txErr)
		return err
	}
	if com.RowsAffected() == int64(len(data)) {
		commitCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := tx.Commit(commitCtx)
		if err != nil {
			return err
		}
		return nil
	} else {
		panic(fmt.Sprintf("Number of affected rows %v != len(data) %v", com.RowsAffected(), len(data)))
	}

}

type PGWriterConfig struct {
	TableName string
	Pool      *pgxpool.Pool
}

func NewPGWriter(conf *PGWriterConfig) *PGWriter {
	return &PGWriter{ConnPool: conf.Pool, TableName: conf.TableName}

}

// BuildQueryString - generates single  `INSERT` SQL query for given slice of datapoints
// considering number of fields in a datapoint model
func BuildQueryString(tablename string, numRecords, numColumns int) string {
	insert := fmt.Sprintf("INSERT INTO %v (uid, datetime, value) VALUES ", tablename)
	argCounter := 0
	var sb strings.Builder
	// 2 parentheses, $n + coma per column
	charsPerEntry := 2 + 3*numColumns
	sb.Grow(len(insert) + charsPerEntry*numRecords)
	sb.WriteString(insert)
	for i := 0; i < numRecords; i++ {
		sb.WriteString("(")
		for j := 0; j < numColumns; j++ {
			argCounter++
			sb.WriteString("$")
			sb.WriteString(strconv.Itoa(argCounter))
			if j != numColumns-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(")")
		if i != numRecords-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString(";")
	return sb.String()

}
