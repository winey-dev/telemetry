package metric

import "sync"

type metricMap struct {
	mtx       sync.RWMutex
	metrics   map[uint64][]metricWithTagValues
	desc      *Desc
	newMetric func(tagValues ...string) Metric
}

type metricWithTagValues struct {
	values []string
	metric Metric
}

func (m *metricMap) Describe(ch chan<- *Desc) {
	ch <- m.desc
}

func (m *metricMap) Collect(ch chan<- Metric) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, metrics := range m.metrics {
		for _, metric := range metrics {
			ch <- metric.metric
		}
	}
}

func (m *metricMap) Reset() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for h := range m.metrics {
		delete(m.metrics, h)
	}
}

func (m *metricMap) deleteWithTagValues(h uint64, tagValues []string) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	metrics, ok := m.metrics[h]
	if !ok {
		return false
	}
	i := findMetricWithTagValues(metrics, tagValues)
	if i >= len(metrics) {
		return false
	}

	if len(metrics) > 1 {
		old := metrics
		m.metrics[h] = append(metrics[:i], metrics[i+1:]...)
		old[len(old)-1] = metricWithTagValues{}
	} else {
		delete(m.metrics, h)
	}

	return false
}

func (m *metricMap) getOrCreateWithTagValues(hash uint64, tagValues []string) Metric {
	m.mtx.RLock()
	metric, ok := m.getMetricWithTagValues(hash, tagValues)
	m.mtx.RUnlock()
	if ok {
		return metric
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	metric = m.newMetric(tagValues...)
	m.metrics[hash] = append(m.metrics[hash], metricWithTagValues{
		values: tagValues,
		metric: metric,
	})
	return metric
}

func (m *metricMap) getMetricWithTagValues(h uint64, tagValues []string) (Metric, bool) {
	metrics, ok := m.metrics[h]
	if ok {
		if i := findMetricWithTagValues(metrics, tagValues); i < len(metrics) {
			return metrics[i].metric, true
		}
	}

	return nil, false
}

func findMetricWithTagValues(metrics []metricWithTagValues, tagValues []string) int {
	for i, metric := range metrics {
		if matchLabelValues(metric.values, tagValues) {
			return i
		}
	}
	return -1
}

func matchLabelValues(values, tagValues []string) bool {
	if len(values) != len(tagValues) {
		return false
	}
	for i, value := range values {
		if value != tagValues[i] {
			return false
		}
	}
	return true
}
