package gojobs

import (
	"math/rand"
	"time"

	"github.com/vjeantet/dashing-go"
)

type convergence struct {
	points []map[string]int
}

func (j *convergence) Work(send chan *dashing.Event, webroot string) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			j.points = j.points[1:]
			j.points = append(j.points, map[string]int{
				"x": j.points[len(j.points)-1]["x"] + 1,
				"y": rand.Intn(50),
			})
			send <- dashing.NewEvent("convergence", map[string]interface{}{
				"points": j.points,
			}, "")
		}
	}
}

func init() {
	c := &convergence{}
	for i := 0; i < 10; i++ {
		c.points = append(c.points, map[string]int{
			"x": i,
			"y": rand.Intn(50),
		})
	}
	dashing.Register(c)
}
