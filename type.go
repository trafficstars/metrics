package metrics

type Type int

const (
	TypeCount = iota
	TypeGaugeInt64
	TypeGaugeInt64Func
	TypeGaugeFloat64
	TypeGaugeFloat64Func
	TypeGaugeAggregativeFlow
	TypeGaugeAggregativeBuffered
	TypeGaugeAggregativeSimple
	TypeTimingFlow
	TypeTimingBuffered
	TypeTimingSimple
)

var (
	// It's here to be static (to do not do memory allocation every time)
	typeStrings = map[Type]string{
		TypeCount:                    `count`,
		TypeGaugeInt64:               `gauge_int64`,
		TypeGaugeInt64Func:           `gauge_int64_func`,
		TypeGaugeFloat64:             `gauge_float64`,
		TypeGaugeFloat64Func:         `gauge_float64_func`,
		TypeGaugeAggregativeFlow:     `gauge_aggregative_flow`,
		TypeGaugeAggregativeBuffered: `gauge_aggregative_buffered`,
		TypeGaugeAggregativeSimple:   `gauge_aggregative_simple`,
		TypeTimingFlow:               `timing_flow`,
		TypeTimingBuffered:           `timing_buffered`,
		TypeTimingSimple:             `timing_simple`,
	}
)

func (t Type) String() string {
	return typeStrings[t]
}
