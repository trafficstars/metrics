package metrics

type Type int

const (
	TypeTiming = iota
	TypeGaugeInt64
	TypeGaugeInt64Func
	TypeGaugeFloat64
	TypeGaugeFloat64Func
	TypeGaugeAggregative
	TypeCount
)

var (
	// It's here to be static (to do not do memory allocation every time)
	typeStrings = map[Type]string{
		TypeTiming:           `timing`,
		TypeGaugeInt64:       `gauge_int64`,
		TypeGaugeInt64Func:   `gauge_int64_func`,
		TypeGaugeFloat64:     `gauge_float64`,
		TypeGaugeFloat64Func: `gauge_float64_func`,
		TypeGaugeAggregative: `gauge_aggregative`,
		TypeCount:            `count`,
	}
)

func (t Type) String() string {
	return typeStrings[t]
}
