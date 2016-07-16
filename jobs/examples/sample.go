package gojobs

import (
	"math/rand"
	"time"

	"github.com/vjeantet/dashing-go"
)

type sample struct{}

func (j *sample) Work(send chan *dashing.Event, webroot string) {
	ticker := time.NewTicker(1 * time.Second)
	var lastValuation, lastKarma, currentValuation, currentKarma int
	for {
		select {
		case <-ticker.C:
			lastValuation, currentValuation = currentValuation, rand.Intn(100)
			lastKarma, currentKarma = currentKarma, rand.Intn(300)
			send <- dashing.NewEvent("valuation", map[string]interface{}{
				"current": currentValuation,
				"last":    lastValuation,
			}, "")
			send <- dashing.NewEvent("karma", map[string]interface{}{
				"current": currentKarma,
				"last":    lastKarma,
			}, "")

			send <- dashing.NewEvent("synergy", map[string]interface{}{
				"value": rand.Intn(100),
			}, "")
		}
	}
}

func init() {
	dashing.Register(&sample{})
}
