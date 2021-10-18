package base

import "time"

// for time gauge <5ms,10ms,100ms,500ms,1s,5s,10s,30s,>30s
const (
	time5ms   = "<5ms"
	time10ms  = "10ms"
	time100ms = "100ms"
	time500ms = "500ms"
	time1s    = "1s"
	time5s    = "5s"
	time10s   = "10s"
	time30s   = "30s"
	timeG30s  = ">30s"
)

type Duration time.Duration

func (d Duration) String() string {
	switch {
	case time.Duration(d) < 5*time.Millisecond:
		return time5ms

	case time.Duration(d) >= 5*time.Millisecond && time.Duration(d) < 10*time.Millisecond:
		return time10ms

	case time.Duration(d) >= 10*time.Millisecond && time.Duration(d) < 100*time.Millisecond:
		return time100ms

	case time.Duration(d) >= 100*time.Millisecond && time.Duration(d) < 500*time.Millisecond:
		return time500ms

	case time.Duration(d) >= 500*time.Millisecond && time.Duration(d) < time.Second:
		return time1s

	case time.Duration(d) >= time.Second && time.Duration(d) < 5*time.Second:
		return time5s

	case time.Duration(d) >= 5*time.Second && time.Duration(d) < 10*time.Second:
		return time10s

	case time.Duration(d) >= 10*time.Second && time.Duration(d) < 30*time.Second:
		return time30s

	default:
		return timeG30s
	}
}
