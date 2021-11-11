package writebuffer

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	cfg "github.com/qwlt/gmcollector/app/config"
	db "github.com/qwlt/gmcollector/app/db"
	m "github.com/qwlt/gmcollector/app/models"
	"github.com/spf13/viper"
)

var WB *WriteBuffer

var TimeoutError = errors.New("Buffer write timed out")

type StorageInterface interface {
	Write(data []m.Model) error
}

type WriteBuffer struct {
	mu       sync.Mutex
	Buff     []m.Model
	dataChan chan m.Model
	stopChan chan int64
	Storage  StorageInterface
	Conf     WBufferConfig
}

// BufMaxSize - max amount of records inside a buffer before it will be flushed to permanent storage
// WriteTimeout - max duration after last buffer flush, after it expires buffer will be forced to flush
// TableName - identifier in permanent storage which is used to save record(real tablename inside SQL storages)
type WBufferConfig struct {
	BufMaxSize   int    `mapstructure:"bufMaxSize"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	TableName    string `mapstructure:"tableName"`
}

func (w *WriteBuffer) AddDatapoint(datapoint m.Model) error {
	select {
	case w.dataChan <- datapoint:
		return nil
	case <-time.After(time.Second * 1):

		return TimeoutError
	}

}

func (w *WriteBuffer) FlushBuffer() error {
	// log.Printf("Flushing at %v", time.Now())
	// start := time.Now()
	// log.Println("Flushing buffer")
	err := w.Storage.Write(w.Buff)
	if err != nil {
		return fmt.Errorf("cant write buffer")
	}
	w.Buff = nil
	// log.Printf("Flushing done in %v", time.Since(start))
	return nil

}

// RunDataHandler run as goroutine and collect values into write buffer
// flush buffer after overflow or after timeout
func (w *WriteBuffer) RunDataHandler() {
	log.Println("Running data handler")
	ticker := time.NewTicker(time.Duration(w.Conf.WriteTimeout) * time.Second)
OuterLoop:
	for {
		select {
		case <-w.stopChan:
			log.Println("write buffer recieve stop chan")
			break OuterLoop

		case m := <-w.dataChan:
			// p.mu.Lock()
			if len(w.Buff) == w.Conf.BufMaxSize {
				w.FlushBuffer()
			}
			// p.mu.Unlock()
			w.Buff = append(w.Buff, m)

		case <-ticker.C:
			// w.mu.Lock()
			w.FlushBuffer()
			// w.mu.Unlock()
		}

	}

}

func (w *WriteBuffer) Shutdown() {
	w.stopChan <- 1
}

func CreateWriteBuffer(config *WBufferConfig) (*WriteBuffer, error) {
	// TODO try different buffer sizes
	buf := WriteBuffer{}
	buf.Conf = *config
	buf.Buff = make([]m.Model, 0, buf.Conf.BufMaxSize)
	buf.dataChan = make(chan m.Model, buf.Conf.BufMaxSize)
	buf.stopChan = make(chan int64)

	ConnPool := db.GetDB()
	var tablename string
	if config.TableName == "" {
		tablename = "measurements"
	} else {
		tablename = config.TableName
	}

	writerConf := &PGWriterConfig{Pool: ConnPool, TableName: tablename}
	buf.Storage = NewPGWriter(writerConf)
	return &buf, nil
}

func GetBuffer() (*WriteBuffer, error) {
	if WB == nil {
		poolConf := WBufferConfig{}
		if viper.IsSet("pool") {
			err := viper.UnmarshalKey("pool", &poolConf)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(&cfg.ViperKeyNotFoundError{Key: "pool", Config: viper.ConfigFileUsed()})
		}

		pool, err := CreateWriteBuffer(&poolConf)
		if err != nil {
			return nil, err
		}
		WB = pool
	}
	return WB, nil

}
