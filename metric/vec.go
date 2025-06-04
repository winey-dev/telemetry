package metric

type MetricVec struct {
	*metricMap

	hashAdd     func(h uint64, s string) uint64
	hashAddByte func(h uint64, b byte) uint64
}

func NewMetricVec(desc *Desc, newMetric func(tagValues ...string) Metric) *MetricVec {
	return &MetricVec{
		metricMap: &metricMap{
			metrics:   make(map[uint64][]metricWithTagValues),
			desc:      desc,
			newMetric: newMetric,
		},
		hashAdd:     hashAdd,
		hashAddByte: hashAddByte,
	}
}

func (m *MetricVec) Describe(ch chan<- *Desc) { m.metricMap.Describe(ch) }

func (m *MetricVec) Collect(ch chan<- Metric) { m.metricMap.Collect(ch) }

// Reset deletes all metrics in this vector.
func (m *MetricVec) Reset() { m.metricMap.Reset() }

func (m *MetricVec) DeletTagValues(tagValues ...string) bool {
	if len(tagValues) != len(m.desc.TagNames) {
		return false
	}

	h, err := m.hashTagValues(tagValues)
	if err != nil {
		return false
	}
	return m.deleteWithTagValues(h, tagValues)
}

func (m *MetricVec) WithTagValues(tagValues ...string) (Metric, error) {
	if len(tagValues) != len(m.desc.TagNames) {
		return nil, ErrInvalidTagValues
	}

	h, err := m.hashTagValues(tagValues)
	if err != nil {
		return nil, err
	}

	return m.getOrCreateWithTagValues(h, tagValues), nil
}

func (m *MetricVec) hashTagValues(tagValues []string) (uint64, error) {
	var h = hashNew()
	for i := 0; i < len(m.desc.TagNames); i++ {
		h = m.hashAdd(h, tagValues[i])
		h = m.hashAddByte(h, seperatorByte)
	}
	return h, nil
}
