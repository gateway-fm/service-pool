package pool

import (
	"time"

	"github.com/gateway-fm/prover-pool-lib/service"
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
	//https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
	temp := make([]service.IService, 0)
	temp = append(temp, slice[:index]...)
	return append(temp, slice[index+1:]...)
}
