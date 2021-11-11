package writebuffer

import (
	"log"
	"sync"
	"testing"

	"github.com/qwlt/gmcollector/app/models"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestPoolProcessDataUnderBuffLimit(t *testing.T) {
	var conf = WBufferConfig{BufMaxSize: 10, WriteTimeout: 10}
	var pool WriteBuffer
	pool.Conf = conf
	pool.Buff = make([]models.Model, 10)
	pool.Storage = &MockStorage{}
	pool.dataChan = make(chan models.Model)
	pool.stopChan = make(chan int64)
	defer pool.Shutdown()
	go pool.RunDataHandler()
	var wg sync.WaitGroup

	for i := 0; i == 10; i++ {
		wg.Add(1)
		go DatapointWriter(&pool, float64(i), &wg)
	}
	wg.Wait()
	if len(pool.Buff) != 10 {
		t.Fatalf("Buffer len should be 10,got %v", len(pool.Buff))
	}
	pool.AddDatapoint(models.Measurement{})
	if len(pool.Buff) != 1 {
		t.Fatalf("Buffer len should be 1,got %v", len(pool.Buff))
	}

}

func TestPoolBufferOverflow(t *testing.T) {
	var conf = WBufferConfig{BufMaxSize: 10, WriteTimeout: 10}
	var pool WriteBuffer
	pool.Conf = conf
	pool.Buff = make([]models.Model, 10)
	pool.Storage = &MockStorage{}
	pool.dataChan = make(chan models.Model)
	pool.stopChan = make(chan int64)
	defer pool.Shutdown()
	go pool.RunDataHandler()
	for i := 0; i < pool.Conf.BufMaxSize+1; i++ {
		pool.AddDatapoint(models.Measurement{Value: float64(i)})
	}
	if len(pool.Buff) != 1 {
		log.Fatalf("Buffer len must be equal to 1, but its equal = %v", len(pool.Buff))
	}

}

func DatapointWriter(p *WriteBuffer, v float64, wg *sync.WaitGroup) {
	p.AddDatapoint(models.Measurement{Value: v})
	wg.Done()
}
