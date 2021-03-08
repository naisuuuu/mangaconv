package imgutil

import "sync"

func clamp(f float64) uint8 {
	v := int(f + 0.5)
	if v > 255 {
		return 255
	}
	if v > 0 {
		return uint8(v)
	}
	return 0
}

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
