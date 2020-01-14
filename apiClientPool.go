package main

import (
	"fmt"
	"sync"

	sfoxapi "github.com/ldcicconi/sfox-api-lib"
)

type SFOXAPIClientPool struct {
	ready []*sfoxapi.SFOXAPI
	lock  sync.Mutex
}

var ErrNoClientAvailable = fmt.Errorf("no client available in pool")

func NewSFOXAPIClientPool(apiKeys []string, numOfConnections int) *SFOXAPIClientPool {
	ready := []*sfoxapi.SFOXAPI{}
	apiErrorMonitor := sfoxapi.NewMonitor()
	for i := 0; i < 20; i++ {
		ready = append(ready, sfoxapi.NewSFOXAPI(apiKeys[i%len(apiKeys)], apiErrorMonitor))
	}
	apiErrorMonitor.Start()
	return &SFOXAPIClientPool{
		ready: ready,
	}
}

// NOTE: This could be a blocking call (if there is no api waiting in the ReadyQueue)
func (pool *SFOXAPIClientPool) GetAPIClient() (*sfoxapi.SFOXAPI, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if len(pool.ready) == 0 {
		return nil, ErrNoClientAvailable
	}
	var newReadyQueue []*sfoxapi.SFOXAPI
	client, newReadyQueue := pool.ready[len(pool.ready)-1], pool.ready[:len(pool.ready)-1]
	pool.ready = newReadyQueue
	return client, nil
}

func (pool *SFOXAPIClientPool) ReturnAPIClient(c *sfoxapi.SFOXAPI) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.ready = append(pool.ready, c)
	return
}
