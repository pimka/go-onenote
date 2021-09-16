package db_test

import (
	"context"
	"github.com/pimka/go-onenote/db"
	"testing"
	"time"
)

func TestPurger(t *testing.T) {
	nh := NewMockDB()
	p := db.NewPurger(nh, time.Second, 5)

	p.Purge(context.Background())
	time.Sleep(time.Second * 10)
	p.Stop() <- struct{}{}
	time.Sleep(time.Second)
	<-p.Done()
}
