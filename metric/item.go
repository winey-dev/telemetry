package metric

import (
	"math"
	"sync/atomic"

	"github.com/winey-dev/telemetry/dto"
)

type Item interface {
	Metric
	Collector

	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
	Min(float64)
	Max(float64)
	//Avg(float64)

	IsError() bool
	Error() error
}

type ItemOpts Opts

func NewItem(opts ItemOpts) *item {
	if opts.Category == "" || opts.SubCategory == "" || opts.ItemName == "" {
		panic(ErrRequiredFields.Error())
	}
	if !opts.ConstraintTags.IsEmpty() && !opts.ConstraintTags.IsValid() {
		panic(ErrInvalidTagValues.Error())
	}
	desc := NewDesc(opts.Category, opts.SubCategory, opts.ItemName, opts.Description, opts.ConstraintTags)
	result := &item{desc: desc}
	result.init(result)
	return result
}

type item struct {
	valBits uint64
	//valInt  uint64

	selfCollector
	desc      *Desc
	tagValues []string
	err       error
}

// implement Metric interface
func (i *item) Desc() *Desc {
	return i.desc
}

func (i *item) Write(out *dto.Metric) error {
	if i.IsError() {
		return i.Error()
	}

	// Write 호출 시점에 valBits를 원자적으로 읽고 0으로 초기화 시킨다.
	valBits := atomic.SwapUint64(&i.valBits, 0)
	val := math.Float64frombits(valBits)
	//val := math.Float64frombits(atomic.LoadUint64(&i.valBits))

	out.Category = i.desc.Category
	out.SubCategory = i.desc.SubCategory
	out.ItemName = i.desc.ItemName
	out.Description = i.desc.Description
	out.TagNames = append(i.desc.ConstraintTags.TagNames, i.desc.TagNames...)
	out.TagValues = append(i.desc.ConstraintTags.TagValues, i.tagValues...)
	out.Value = val
	return nil
}

// implement Item interface
func (i *item) Set(value float64) {
	atomic.StoreUint64(&i.valBits, math.Float64bits(value))
}

func (i *item) Inc() {
	i.Add(1)
}

func (i *item) Dec() {
	i.Add(-1)
}
func (i *item) Add(value float64) {
	for {
		oldBits := atomic.LoadUint64(&i.valBits)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + value)
		if atomic.CompareAndSwapUint64(&i.valBits, oldBits, newBits) {
			return
		}
	}
}

func (i *item) Sub(value float64) {
	i.Add(value * -1)
}

func (i *item) Min(value float64) {
	for {
		oldBits := atomic.LoadUint64(&i.valBits)
		oldValue := math.Float64frombits(oldBits)
		if oldValue <= value {
			return
		}
		newBits := math.Float64bits(value)
		if atomic.CompareAndSwapUint64(&i.valBits, oldBits, newBits) {
			return
		}
	}
}

func (i *item) Max(value float64) {
	for {
		oldBits := atomic.LoadUint64(&i.valBits)
		oldValue := math.Float64frombits(oldBits)
		if oldValue >= value {
			return
		}
		newBits := math.Float64bits(value)
		if atomic.CompareAndSwapUint64(&i.valBits, oldBits, newBits) {
			return
		}
	}
}

func (i *item) IsError() bool {
	return i.err != nil
}

func (i *item) Error() error {
	return i.err
}

type ItemVec struct {
	*MetricVec
}

// 동적 태그 밸류를 갖는 아이템 벡터를 생성하기 위한 생성자
// 고정 태그 + 동적 태그 밸류를 안전하기 관리하기 위해 Hash 및 Map을 사용
func NewItemVec(opts ItemOpts, tagNames ...string) *ItemVec {
	if len(tagNames) == 0 {
		panic("tagNames must not be empty")
	}
	desc := NewDesc(opts.Category, opts.SubCategory, opts.ItemName, opts.Description, opts.ConstraintTags, tagNames...)
	return &ItemVec{
		MetricVec: NewMetricVec(desc, func(tagValues ...string) Metric {
			if len(tagValues) != len(desc.TagNames) {
				// panic을 사용하지 않고 errorMetric 구조를 반환하도록 변경
				panic("tagValues length does not match tagNames length")
			}
			result := &item{desc: desc, tagValues: tagValues}
			result.init(result)
			return result
		}),
	}
}

func (v *ItemVec) WithTagValues(tagValues ...string) Item {
	metric, err := v.MetricVec.WithTagValues(tagValues...)
	if err != nil {
		return &item{err: err}
	}
	return metric.(Item)
}
