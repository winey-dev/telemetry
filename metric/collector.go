package metric

type Collector interface {
	Describe(ch chan<- *Desc)
	Collect(ch chan<- Metric)
}

type selfCollector struct {
	self Metric
}

func (c *selfCollector) init(m Metric) {
	c.self = m
}
func (c *selfCollector) Describe(ch chan<- *Desc) {
	ch <- c.self.Desc()
}
func (c *selfCollector) Collect(ch chan<- Metric) {
	ch <- c.self
}
