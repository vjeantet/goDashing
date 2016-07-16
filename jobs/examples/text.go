package gojobs

import (
	"time"

	"github.com/vjeantet/goDash"
)

type sampleTXT struct{}

func (j *sampleTXT) Work(send chan *dashing.Event, webroot string) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			send <- dashing.NewEvent("sampleTXT", map[string]interface{}{
				"text": "<u>currentValuation</u>",
			}, "")

		}
	}
}

func init() {
	dashing.Register(&sampleTXT{})
}
