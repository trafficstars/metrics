package main

import (
	"fmt"
	"math/rand"

	"github.com/trafficstars/metrics"
)

func main() {
	metric := metrics.GaugeAggregativeFlow(`value`, nil)

	rand.Seed(1)
	for i := 0; i < 4000; i++ {
		v := rand.Float64()
		metric.ConsiderValue(v)
		fmt.Printf("%v\t%v\t%v\t%v\n",
			v,
			*metric.GetValuePointers().Total().GetPercentile(0.1),
			*metric.GetValuePointers().Total().GetPercentile(0.9),
			*metric.GetValuePointers().Total().GetPercentile(0.99),
		)
	}
}
