package metrics

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/trafficstars/fastmetrics/worker"
)

var (
	ErrDisabled = errors.New(`metrics are disabled`)
)

var (
	dummyWorkerCount = metricworker.NewWorkerCount(nil, ``)
	dummyWorkerGauge = metricworker.NewWorkerGauge(nil, ``)
	dummyWorkerGaugeFloat = metricworker.NewWorkerGaugeFloat(nil, ``)
	dummyWorkerGaugeFloatAggregative = metricworker.NewWorkerGaugeFloatAggregative(nil, ``)
	dummyWorkerGaugeFunc = metricworker.NewWorkerGaugeFunc(nil, ``, nil)
	dummyWorkerGaugeFloatFunc = metricworker.NewWorkerGaugeFloatFunc(nil, ``, nil)
	dummyWorkerTiming = metricworker.NewWorkerTiming(nil, ``)
)

func init() {
	// Dropping the pointers to "nil" to improve performance
	dummyWorkerCount = nil
	dummyWorkerGauge = nil
	dummyWorkerGaugeFloat = nil
	dummyWorkerGaugeFloatAggregative = nil
	dummyWorkerGaugeFunc = nil
	dummyWorkerGaugeFloatFunc = nil
	dummyWorkerTiming = nil
}

func runAndRegisterWorkerWrapper(key string, worker Worker, tags AnyTags) error {
	err := runAndRegister(key, worker, tags)
	if err != nil {
		return nil
		worker.Stop()
		log.WithFields(log.Fields{
			"metric_key": key,
			"tags":       tags,
		}).Errorf(`Cannot register a metric "%s": %v`, key, err)
	}
	return err
}

func createWorkerCount(key string, tags AnyTags) (WorkerCount, error) {
	if metrics.IsDisabled() {
		return dummyWorkerCount, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerCount(metrics.GetSender(), statsdKey)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerGauge(key string, tags AnyTags) (WorkerGauge, error) {
	if metrics.IsDisabled() {
		return dummyWorkerGauge, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerGauge(metrics.GetSender(), statsdKey)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerGaugeFloat(key string, tags AnyTags) (WorkerGaugeFloat, error) {
	if metrics.IsDisabled() {
		return dummyWorkerGaugeFloat, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerGaugeFloat(metrics.GetSender(), statsdKey)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerGaugeFloatAggregative(key string, tags AnyTags) (WorkerGaugeFloatAggregative, error) {
	if metrics.IsDisabled() {
		return dummyWorkerGaugeFloatAggregative, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerGaugeFloatAggregative(metrics.GetSender(), statsdKey)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerGaugeFunc(key string, tags AnyTags, fn func() int64) (WorkerGaugeFunc, error) {
	if metrics.IsDisabled() {
		return dummyWorkerGaugeFunc, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerGaugeFunc(metrics.GetSender(), statsdKey, fn)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerGaugeFloatFunc(key string, tags AnyTags, fn func() float64) (WorkerGaugeFloatFunc, error) {
	if metrics.IsDisabled() {
		return dummyWorkerGaugeFloatFunc, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerGaugeFloatFunc(metrics.GetSender(), statsdKey, fn)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func createWorkerTiming(key string, tags AnyTags) (WorkerTiming, error) {
	if metrics.IsDisabled() {
		return dummyWorkerTiming, ErrDisabled
	}
	keyBuf := generateStorageKey("", key, tags)
	statsdKey := keyBuf.result.String()
	keyBuf.Unlock()
	worker := metricworker.NewWorkerTiming(metrics.GetSender(), statsdKey)
	worker.SetGCEnabled(true)
	return worker, runAndRegisterWorkerWrapper(key, worker, tags)
}

func CreateOrGetWorkerCountWithError(key string, tags AnyTags) (WorkerCount, error) {
	m := Get(MetricTypeCount, key, tags)
	if m != nil {
		return m.worker.(WorkerCount), nil
	}
	return createWorkerCount(key, tags)
}

func CreateOrGetWorkerCount(key string, tags AnyTags) WorkerCount {
	worker, _ := CreateOrGetWorkerCountWithError(key, tags)
	return worker
}

func CreateOrGetWorkerGaugeWithError(key string, tags AnyTags) (WorkerGauge, error) {
	m := Get(MetricTypeGauge, key, tags)
	if m != nil {
		return m.worker.(WorkerGauge), nil
	}
	return createWorkerGauge(key, tags)
}

func CreateOrGetWorkerGauge(key string, tags AnyTags) WorkerGauge {
	worker, _ := CreateOrGetWorkerGaugeWithError(key, tags)
	return worker
}

func CreateOrGetWorkerGaugeFloatWithError(key string, tags AnyTags) (WorkerGaugeFloat, error) {
	m := Get(MetricTypeGauge, key, tags)
	if m != nil {
		return m.worker.(WorkerGaugeFloat), nil
	}
	return createWorkerGaugeFloat(key, tags)
}

func CreateOrGetWorkerGaugeFloat(key string, tags AnyTags) WorkerGaugeFloat {
	worker, _ := CreateOrGetWorkerGaugeFloatWithError(key, tags)
	return worker
}

func CreateOrGetWorkerGaugeFloatAggregativeWithError(key string, tags AnyTags) (WorkerGaugeFloatAggregative, error) {
	m := Get(MetricTypeGauge, key, tags)
	if m != nil {
		return m.worker.(WorkerGaugeFloatAggregative), nil
	}
	return createWorkerGaugeFloatAggregative(key, tags)
}

func CreateOrGetWorkerGaugeFloatAggregative(key string, tags AnyTags) WorkerGaugeFloatAggregative {
	worker, _ := CreateOrGetWorkerGaugeFloatAggregativeWithError(key, tags)
	return worker
}

func CreateOrGetWorkerGaugeFuncWithError(key string, tags AnyTags, fn func() int64) (WorkerGaugeFunc, error) {
	m := Get(MetricTypeGauge, key, tags)
	if m != nil {
		return m.worker.(WorkerGaugeFunc), nil
	}
	return createWorkerGaugeFunc(key, tags, fn)
}

func CreateOrGetWorkerGaugeFunc(key string, tags AnyTags, fn func() int64) WorkerGaugeFunc {
	worker, _ := CreateOrGetWorkerGaugeFuncWithError(key, tags, fn)
	return worker
}

func CreateOrGetWorkerGaugeFloatFuncWithError(key string, tags AnyTags, fn func() float64) (WorkerGaugeFloatFunc, error) {
	m := Get(MetricTypeGauge, key, tags)
	if m != nil {
		return m.worker.(WorkerGaugeFloatFunc), nil
	}
	return createWorkerGaugeFloatFunc(key, tags, fn)
}

func CreateOrGetWorkerGaugeFloatFunc(key string, tags AnyTags, fn func() float64) WorkerGaugeFloatFunc {
	worker, _ := CreateOrGetWorkerGaugeFloatFuncWithError(key, tags, fn)
	return worker
}

func CreateOrGetWorkerTimingWithError(key string, tags AnyTags) (WorkerTiming, error) {
	m := Get(MetricTypeTiming, key, tags)
	if m != nil {
		return m.worker.(WorkerTiming), nil
	}
	return createWorkerTiming(key, tags)
}

func CreateOrGetWorkerTiming(key string, tags AnyTags) WorkerTiming {
	worker, _ := CreateOrGetWorkerTimingWithError(key, tags)
	return worker
}
