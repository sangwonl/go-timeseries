package timeseries

import (
	"container/list"
	"time"
)

func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

type Primitive interface {
	Add(other Primitive)
	CopyFrom(other Primitive)
	Reset()
}

type Integer int

func NewInteger() Primitive                 { i := Integer(0); return &i }
func (i *Integer) Value() int               { return int(*i) }
func (i *Integer) Add(other Primitive)      { *i += *(other.(*Integer)) }
func (i *Integer) CopyFrom(other Primitive) { *i = *(other.(*Integer)) }
func (i *Integer) Reset()                   { *i = 0 }

const (
	ResolutionOneSecond  = 1 * time.Second
	ResolutionTenSeconds = 10 * time.Second
	ResolutionOneMinute  = 1 * time.Minute
	ResolutionTenMinutes = 10 * time.Minute
	ResolutionOneHour    = 1 * time.Hour
	ResolutionSixHours   = 6 * time.Hour
	ResolutionOneDay     = 24 * time.Hour
	ResolutionOneWeek    = 7 * 24 * time.Hour
	ResolutionFourWeeks  = 4 * 7 * 24 * time.Hour
)

type dataStream struct {
	primitiveFunc func() Primitive
	buckets       *list.List
	resolution    time.Duration
	beginTime     time.Time
	endTime       time.Time
}

func (ds *dataStream) initialize(p func() Primitive, resolution time.Duration) {
	ds.primitiveFunc = p
	ds.resolution = resolution
	ds.buckets = list.New()
}

func (ds *dataStream) NumBuckets() int {
	return ds.buckets.Len()
}

func (ds *dataStream) reset() {
	ds.beginTime = time.Time{}
	ds.endTime = time.Time{}

	var next *list.Element
	for e := ds.buckets.Front(); e != nil; e = next {
		next = e.Next()
		ds.buckets.Remove(e)
	}
}

type TimeSeries struct {
	primitiveFunc func() Primitive
	dataStreams   []*dataStream
	total         Primitive
}

func NewTimeSeries(p func() Primitive, resolutions []time.Duration) *TimeSeries {
	timeSeries := new(TimeSeries)
	timeSeries.initialize(p, resolutions)
	return timeSeries
}

func (ts *TimeSeries) initialize(p func() Primitive, resolutions []time.Duration) {
	ts.primitiveFunc = p
	ts.total = ts.primitiveFunc()
	ts.dataStreams = make([]*dataStream, len(resolutions))
	for i := range resolutions {
		ts.dataStreams[i] = new(dataStream)
		ts.dataStreams[i].initialize(p, resolutions[i])
	}
	ts.reset()
}

func (ts *TimeSeries) reset() {
	ts.total.Reset()
	for i := range ts.dataStreams {
		ts.dataStreams[i].reset()
	}
}

func (ts *TimeSeries) Add(d Primitive, t time.Time) {
	for _, ds := range ts.dataStreams {
		isFirstAdd := ds.buckets.Len() == 0
		if isFirstAdd {
			ds.beginTime = t
			ds.endTime = t

			first := ds.primitiveFunc()
			ds.buckets.PushBack(first)
		}

		bucketIdxFromEnd := int(t.Sub(ds.endTime) / ds.resolution)
		for i := 0; i < bucketIdxFromEnd; i++ {
			p := ds.primitiveFunc()
			ds.buckets.PushBack(p)
		}

		lastBucket := ds.buckets.Back().Value.(Primitive)
		lastBucket.Add(d)

		// update begin and end time
		ds.beginTime = minTime(ds.beginTime, t)
		ds.endTime = maxTime(ds.endTime, t)
	}
	ts.total.Add(d)
}

func (ts *TimeSeries) Total() Primitive {
	return ts.total
}

func (ts *TimeSeries) Range(resolutionIdx int, fromTime, toTime time.Time) []Primitive {
	ds := ts.dataStreams[resolutionIdx]

	beginBucketIdx := int(fromTime.Sub(ds.beginTime) / ds.resolution)
	endBucketIdx := int(toTime.Sub(ds.beginTime) / ds.resolution)
	filteredBuckets := list.New()

	iterIdx := 0
	for e := ds.buckets.Front(); e != nil; e = e.Next() {
		if beginBucketIdx <= iterIdx && iterIdx <= endBucketIdx {
			filteredBuckets.PushBack(e.Value)
		}
		iterIdx++
	}

	filtered := make([]Primitive, filteredBuckets.Len())
	insertIdx := 0
	for e := filteredBuckets.Front(); e != nil; e = e.Next() {
		filtered[insertIdx] = e.Value.(Primitive)
		insertIdx++
	}

	return filtered
}
