package gojobs

import (
	"math/rand"
	"time"

	"github.com/vjeantet/goDashing"
)

type sparklinechart struct{}

func (j *sparklinechart) Work(send chan *dashing.Event, webroot string) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:

			send <- dashing.NewEvent("water_main_city", map[string]interface{}{
				"points": []map[string]interface{}{
					{
						"x": 1,
						"y": rand.Intn(70),
					}, {
						"x": 2,
						"y": rand.Intn(70),
					}, {
						"x": 3,
						"y": rand.Intn(70),
					}, {
						"x": 4,
						"y": rand.Intn(70),
					}, {
						"x": 5,
						"y": rand.Intn(70),
					}, {
						"x": 6,
						"y": rand.Intn(70),
					}, {
						"x": 7,
						"y": rand.Intn(70),
					}, {
						"x": 8,
						"y": rand.Intn(70),
					}, {
						"x": 9,
						"y": rand.Intn(70),
					}, {
						"x": 10,
						"y": rand.Intn(70),
					},
				},
			}, "")

		}
	}
}

func init() {
	dashing.Register(&sparklinechart{})
}

// curl -d '{
//   "auth_token": "YOUR_AUTH_TOKEN",
//   "points":
//     [
//       { "x": "1",  "y": "10" },
//       { "x": "2",  "y": "20" },
//       { "x": "3",  "y": "70" },
//       { "x": "4",  "y": "60" },
//       { "x": "5",  "y": "10" },
//       { "x": "6",  "y": "80" },
//       { "x": "7",  "y": "90" },
//       { "x": "8",  "y": "40" },
//       { "x": "9",  "y": "30" },
//       { "x": "10", "y": "10" }
//     ]
//   }' \
// http://127.0.0.1:8080/widgets/water_main_city
