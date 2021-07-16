package schedule

import (
	"context"
	"strings"
	"time"
)

// TimeStream return time stream with synchronized utc time after startTime
// use it for schedule time events in p2p network with second precision
func TimeStream(ctx context.Context, startTime time.Time, interval time.Duration) <-chan time.Time {

	timeStream := make(chan time.Time, 1)

	if startTime.IsZero() {
		startTime = time.Now().UTC().Truncate(interval)
	}

	diff := startTime.Sub(time.Now().UTC())
	if diff < 0 {
		total := diff - interval
		times := total / interval * -1

		startTime = startTime.Add(times * interval)
	}

	go func() {

		// run the first event after start time
		t := <-time.After(startTime.Sub(time.Now().UTC()))
		timeStream <- t.UTC()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				timeStream <- t.UTC()
			case <-ctx.Done():
				close(timeStream)
				return
			}
		}
	}()

	return timeStream
}

func TimeToTopicName(prefix string, t time.Time) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(":")
	b.WriteString(t.Format("2006-01-02 15:04"))
	return b.String()
}

func TimeStringToTimeUTC(timeStr string) (time.Time, error) {
	var b strings.Builder
	b.WriteString(time.Now().UTC().Format("2006-01-02 "))
	b.WriteString(timeStr)
	if t, err := time.Parse("2006-01-02 15:04", b.String()); err != nil {
		return time.Time{}, err
	} else {
		return t, nil
	}
}
