package gojobs

import (
	"math/rand"
	"time"

	"github.com/vjeantet/dashing-go"
)

type linechart struct{}

func (j *linechart) Work(send chan *dashing.Event, webroot string) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:

			send <- dashing.NewEvent("linechart", map[string]interface{}{
				"labels": []string{"January", "February", "March", "April", "May", "June", "July"},
				"datasets": []map[string]interface{}{
					{
						"label":           "My First dataset",
						"fillColor":       "rgba(220,220,220,0.5)",
						"strokeColor":     "rgba(220,220,220,0.8)",
						"highlightFill":   "rgba(220,220,220,0.75)",
						"highlightStroke": "rgba(220,220,220,1)",
						"data":            []int{rand.Intn(60), rand.Intn(42), rand.Intn(82), rand.Intn(13), rand.Intn(57), 5, 57},
					}, {
						"label":           "My Second dataset",
						"fillColor":       "rgba(151,187,205,0.5)",
						"strokeColor":     "rgba(151,187,205,0.8)",
						"highlightFill":   "rgba(151,187,205,0.75)",
						"highlightStroke": "rgba(151,187,205,1)",
						"data":            []int{60, rand.Intn(80), 62, rand.Intn(63), 67, rand.Intn(50), 57},
					},
				},
				"options": map[string]string{"pointLabelFontColor": "#fff"},
			}, "")

		}
	}
}

func init() {
	dashing.Register(&linechart{})
}
