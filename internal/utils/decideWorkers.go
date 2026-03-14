package utils

func DecideWorkers(size int64) int {

	switch {
	case size < 10*1024*1024:
		return 2
	case size < 100*1024*1024:
		return 4
	case size < 1*1024*1024*1024:
		return 8
	case size < 10*1024*1024*1024:
		return 16
	default:
		return 32
	}
}
