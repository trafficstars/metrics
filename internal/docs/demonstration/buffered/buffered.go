package main

import (
	"fmt"
	"math/rand"

	"github.com/trafficstars/metrics"
)

func main() {
	metric := metrics.GaugeAggregativeBuffered(`value`, nil)

	rand.Seed(1)
	for i := 0; i < 4000; i++ {
		v := rand.Float64()
		metric.ConsiderValue(v)
		percentiles := metric.GetValuePointers().Total().GetPercentiles([]float64{0.1, 0.9, 0.99})
		fmt.Printf("%v\t%v\t%v\t%v\t%v\n",
			metric.GetValuePointers().Total().Count,
			v,
			*percentiles[0],
			*percentiles[1],
			*percentiles[2],
		)
	}
}
