package wsmock

import "log"

func init() {
	log.Println("using wsmock")
}

func last[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}
