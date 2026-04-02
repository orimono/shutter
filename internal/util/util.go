package util

func DrainChan(c <-chan struct{}) {
	for {
		select {
		case _, ok := <-c:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
