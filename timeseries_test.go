package timeseries

import (
	"testing"
	"time"
)

type Integer int

func NewInteger() Primitive                 { i := Integer(0); return &i }
func (i *Integer) Value() int               { return int(*i) }
func (i *Integer) Add(other Primitive)      { *i += *(other.(*Integer)) }
func (i *Integer) CopyFrom(other Primitive) { *i = *(other.(*Integer)) }
func (i *Integer) Reset()                   { *i = 0 }

func assertEqual(t *testing.T, a, b int) {
	if a != b {
		t.Errorf("a(=%v) should be b(=%v)", a, b)
	}
}

func assertNotNil(t *testing.T, o interface{}) {
	if o == nil {
		t.Errorf("o(=%q) should not be nil", o)
	}
}

func asTime(s int64) time.Time {
	utcZeroTime := time.Time{}
	return utcZeroTime.Add(time.Duration(s) * time.Second)
}

func TestTimeSeriesPrimitive(t *testing.T) {
	i := Integer(1)
	assertEqual(t, i.Value(), 1)

	i2 := Integer(2)
	i.Add(&i2)
	assertEqual(t, i.Value(), 3)

	i.CopyFrom(&i2)
	assertEqual(t, i.Value(), 2)

	i.Reset()
	assertEqual(t, i.Value(), 0)
}

func TestTimeSereis(t *testing.T) {
	primitiveFunc := NewInteger
	resolutions := []time.Duration{
		ResolutionOneSecond,
		ResolutionTenSeconds,
	}

	timeSeries := NewTimeSeries(primitiveFunc, resolutions)
	assertNotNil(t, timeSeries)
	assertEqual(t, len(timeSeries.dataStreams), 2)
	assertEqual(t, timeSeries.dataStreams[0].NumBuckets(), 0)

	i := Integer(1)
	timeSeries.Add(&i, asTime(1))
	timeSeries.Add(&i, asTime(1))
	timeSeries.Add(&i, asTime(1))
	timeSeries.Add(&i, asTime(2))
	timeSeries.Add(&i, asTime(2))
	timeSeries.Add(&i, asTime(3))
	timeSeries.Add(&i, asTime(3))
	timeSeries.Add(&i, asTime(3))
	timeSeries.Add(&i, asTime(3))
	timeSeries.Add(&i, asTime(3))
	assertEqual(t, timeSeries.dataStreams[0].NumBuckets(), 3)

	total := timeSeries.Total().(*Integer)
	assertEqual(t, total.Value(), 10)

	resolutionIdx := 0
	var rangeVals []Primitive
	var bucketVal int

	rangeVals = timeSeries.Range(resolutionIdx, asTime(1), asTime(1))
	assertEqual(t, len(rangeVals), 1)

	bucketVal = rangeVals[0].(*Integer).Value()
	assertEqual(t, bucketVal, 3)

	rangeVals = timeSeries.Range(resolutionIdx, asTime(1), asTime(2))
	assertEqual(t, len(rangeVals), 2)

	bucketVal = rangeVals[0].(*Integer).Value()
	bucketVal += rangeVals[1].(*Integer).Value()
	assertEqual(t, bucketVal, 5)

	rangeVals = timeSeries.Range(resolutionIdx, asTime(2), asTime(2))
	assertEqual(t, len(rangeVals), 1)

	bucketVal = rangeVals[0].(*Integer).Value()
	assertEqual(t, bucketVal, 2)

	rangeVals = timeSeries.Range(resolutionIdx, asTime(2), asTime(3))
	assertEqual(t, len(rangeVals), 2)

	bucketVal = rangeVals[0].(*Integer).Value()
	bucketVal += rangeVals[1].(*Integer).Value()
	assertEqual(t, bucketVal, 7)

	resolutionIdx = 1
	rangeVals = timeSeries.Range(resolutionIdx, asTime(1), asTime(3))
	assertEqual(t, len(rangeVals), 1)

	bucketVal = rangeVals[0].(*Integer).Value()
	assertEqual(t, bucketVal, 10)
}
