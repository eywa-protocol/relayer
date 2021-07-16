package schedule

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeStream(t *testing.T) {

	// skip on short test
	if testing.Short() {
		t.SkipNow()
	}

	// how many time stream nodes to start
	nodesCount := 3
	// interval between scheduled rounds
	roundInterval := 2 * time.Minute
	// stop collect test data on roundsCount
	roundsCount := 3

	wg := new(sync.WaitGroup)

	startTime, err := TimeStringToTimeUTC("00:01")
	assert.NoError(t, err)
	startTime = startTime.Add(-24 * time.Hour)

	rounds := make(map[int][]string)
	mx := new(sync.Mutex)

	ctx, cancel := context.WithCancel(context.Background())

	var currentRound int

	startedNodesCount := 0

	nodesStarted := func() bool {
		mx.Lock()
		defer mx.Unlock()
		return startedNodesCount >= nodesCount
	}

	saveRoundTopic := func(n, round int, topic string) bool {
		mx.Lock()
		defer mx.Unlock()
		if round > currentRound {
			currentRound = round
			fmt.Printf("Round%d:\n", currentRound)
		}
		if _, ok := rounds[currentRound]; !ok {
			rounds[currentRound] = make([]string, 0, nodesCount)
		}
		rounds[currentRound] = append(rounds[currentRound], topic)
		fmt.Printf("node%d topic: %s\n", n, topic)
		return currentRound >= roundsCount
	}

	for i := 1; i < nodesCount+1; i++ {
		fmt.Printf("start node%d at %s\n", i, time.Now().UTC().String())
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			round := 0
			for t := range TimeStream(ctx, startTime, roundInterval) {
				if nodesStarted() {
					round++
					topic := TimeToTopicName("uptime", t)
					if saveRoundTopic(n, round, topic) {
						return
					}
				}
			}
		}(i)
		mx.Lock()
		startedNodesCount++
		mx.Unlock()
		// for emulate run nodes in different time
		time.Sleep(15 * time.Second)
	}

	wg.Wait()
	cancel()

	for _, round := range rounds {
		// check that all nodes
		assert.Equal(t, nodesCount, len(round))
		for i := 0; i < len(round)-1; i++ {
			// check that nodes has same time topic in same round
			assert.Equal(t, round[i], round[i+1])
		}
	}

}

func TestTimeStringToTimeUTC(t *testing.T) {
	tt, err := TimeStringToTimeUTC("23:00")
	assert.NoError(t, err)
	assert.Equal(t, tt, time.Now().UTC().Truncate(24*time.Hour).Add(23*time.Hour))
}
