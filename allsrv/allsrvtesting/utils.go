package allsrvtesting

import (
	"strconv"
	"time"
)

func IDGen(start, incr int) func() string {
	return func() string {
		id := strconv.Itoa(start)
		start += incr
		return id
	}
}

func NowFn(start time.Time, incr time.Duration) func() time.Time {
	return func() time.Time {
		t := start
		start = start.Add(incr)
		return t
	}
}

func Ptr[T any](v T) *T {
	return &v
}
