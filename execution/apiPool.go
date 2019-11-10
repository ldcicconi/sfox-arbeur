package execution

import (
	sfoxapi "github.com/ldcicconi/sfox-api-lib"
)

type SFOXAPIClientPool struct {
	ReadyQueue chan *sfoxapi.SFOXAPI
}

func NewSFOXAPIClientPool(apiKeys []string) *SFOXAPIClientPool {
	ready := make(chan *sfoxapi.SFOXAPI, 10)
	for i := 0; i < 10; i++ {
		ready <- sfoxapi.NewSFOXAPI(apiKeys[i%len(apiKeys)])
	}
	return &SFOXAPIClientPool{
		ReadyQueue: ready,
	}
}

// NOTE: This could be a blocking call (if there is no api waiting in the ReadyQueue)
func (pool *SFOXAPIClientPool) GetAPIClient() *sfoxapi.SFOXAPI {
	ret := <-pool.ReadyQueue
	return ret
}

func (pool *SFOXAPIClientPool) ReturnAPIClient(c *sfoxapi.SFOXAPI) {
	pool.ReadyQueue <- c
	return
}
