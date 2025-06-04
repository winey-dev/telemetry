package influxdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/winey-dev/telemetry/dto"
	"github.com/winey-dev/telemetry/metric"
)

const (
	period = "REALTIME"
)

type Bucket struct {
	items map[string][]*write.Point // Map of Period + Category to Metrics
}

func NewBucket() *Bucket {
	return &Bucket{
		items: make(map[string][]*write.Point),
	}
}

func (b *Bucket) Add(metric metric.Metric, now time.Time) {
	var value dto.Metric
	metric.Write(&value)

	key := strings.Join([]string{period, value.Category}, "_")

	point := toWritePoint(&value, now)
	b.items[key] = append(b.items[key], point)
}

func toWritePoint(m *dto.Metric, now time.Time) *write.Point {
	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["item_name"] = m.ItemName
	for i, tagName := range m.TagNames {
		tags[tagName] = m.TagValues[i]
	}
	fields["value"] = m.Value

	point := write.NewPoint(
		m.SubCategory,
		tags,
		fields,
		now,
	)
	return point
}

func (b *Bucket) Summary(now time.Time) {
	var builder strings.Builder

	builder.WriteString("InfluxDB Bucket Summary:\n")
	builder.WriteString(fmt.Sprintf("- Time: %s\n", now.Format(time.RFC3339)))
	for key, points := range b.items {
		builder.WriteString(fmt.Sprintf("- Bucket: %s, Points: %d\n", key, len(points)))
		for _, point := range points {
			builder.WriteString(fmt.Sprintf("  - Point: %s\n", pointToString(point)))
		}
	}
	fmt.Println(builder.String())
}

func pointToString(point *write.Point) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Measurement: %s, ", point.Name()))
	for _, tag := range point.TagList() {
		builder.WriteString(fmt.Sprintf("%s=%s, ", tag.Key, tag.Value))
	}
	for _, field := range point.FieldList() {
		builder.WriteString(fmt.Sprintf("%s=%v, ", field.Key, field.Value))
	}
	builder.WriteString(fmt.Sprintf("Time: %s", point.Time().Format(time.RFC3339)))
	return builder.String()
}
