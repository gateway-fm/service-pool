package pool

import (
	"time"

	"github.com/gateway-fm/service-pool/service"
)

// Sleep is a helper function to Sleep
// with able to cancel timer
func Sleep(t time.Duration, cancelCh <-chan struct{}) {
	timer := time.NewTimer(t)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-cancelCh:
	}
}

// deleteFromSlice delete item with
// given index from provided slice
func deleteFromSlice(slice []service.IService, index int) []service.IService {
	copy(slice[index:], slice[index+1:])
	slice[len(slice)-1] = nil
	return slice[:len(slice)-1]
}
