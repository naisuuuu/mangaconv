package imgutil

import "sync"

func concurrentIterate(limit int, fn func(int)) {
	var wg sync.WaitGroup
	for j := 0; j < limit; j++ {
		wg.Add(1)
		go func(k int) {
			fn(k)
			wg.Done()
		}(j)
	}
	wg.Wait()
}
