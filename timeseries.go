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
	Ts() time.Time
	SetTs(t time.Time)
}

type Integer struct {
	val int
	ts  time.Time
}

func NewInteger() Primitive                 { i := Integer{0, time.Time{}}; return &i }
func (i *Integer) Value() int               { return i.val }
func (i *Integer) SetValue(v int)           { i.val = v }
func (i *Integer) Add(other Primitive)      { i.val += other.(*Integer).val }
func (i *Integer) CopyFrom(other Primitive) { i.val = other.(*Integer).val }
func (i *Integer) Reset()                   { i.val = 0 }
func (i *Integer) Ts() time.Time            { return i.ts }
func (i *Integer) SetTs(t time.Time)        { i.ts = t }

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
			first.SetTs(ds.beginTime)
			ds.buckets.PushBack(first)
		}

		bucketIdxAtEnd := int(ds.endTime.Sub(ds.beginTime) / ds.resolution)
		bucketIdxFromBegin := int(t.Sub(ds.beginTime) / ds.resolution)
		for i := 0; i < bucketIdxFromBegin-bucketIdxAtEnd; i++ {
			bucketTimeDelta := time.Duration(bucketIdxAtEnd+i+1) * ds.resolution
			bucketTime := ds.beginTime.Add(bucketTimeDelta)
			p := ds.primitiveFunc()
			p.SetTs(bucketTime)
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

func (ts *TimeSeries) All(resolutionIdx int) []Primitive {
	ds := ts.dataStreams[resolutionIdx]
	return primitivesAsArray(ds.buckets)
}

func (ts *TimeSeries) Range(resolutionIdx int, fromTime, toTime time.Time) []Primitive {
	ds := ts.dataStreams[resolutionIdx]
	filtered := filterBucket(ds, fromTime, toTime)
	return primitivesAsArray(filtered)
}

func filterBucket(ds *dataStream, fromTime, toTime time.Time) *list.List {
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
	return filteredBuckets
}

func primitivesAsArray(buckets *list.List) []Primitive {
	primitives := make([]Primitive, buckets.Len())
	insertIdx := 0
	for e := buckets.Front(); e != nil; e = e.Next() {
		primitives[insertIdx] = e.Value.(Primitive)
		insertIdx++
	}
	return primitives
}
