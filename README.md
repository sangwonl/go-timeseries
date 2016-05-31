## go-timeseries 
Golang Timeseries Data Bucket

## Get Package
```
go get github.com/sangwonl/go-timeseries
```

## Import Package
```
import "github.com/sangwonl/go-timeseries"
```

## Create a TimeSeries Struct
```
ts := timeseries.NewTimeSeries(<PrimitiveClass>, <ResolutionList>)
```

## Add Data with Time
```
i := Integer(1)
ts.Add(&i, time.Now())
```

## Extract Data in Time Range
```
rangeVals = ts.Range(resolutionIdx, beginTime, endTime)
firstBucketVal = rangeVals[0].(*Integer).Value()
secondBucketVal = rangeVals[1].(*Integer).Value()
```

## Example
```
import (
    "time"
    "github.com/sangwonl/go-timeseries"
)

type Integer int

func NewInteger() timeseries.Primitive                 { i := Integer(0); return &i }
func (i *Integer) Value() int                          { return int(*i) }
func (i *Integer) Add(other timeseries.Primitive)      { *i += *(other.(*Integer)) }
func (i *Integer) CopyFrom(other timeseries.Primitive) { *i = *(other.(*Integer)) }
func (i *Integer) Reset()                              { *i = 0 }

func TestTimeSeries() {
    primitiveCreator := NewInteger
    resolutions := []time.Duration{
        timeseries.ResolutionOneSecond,
        timeseries.ResolutionTenSeconds,
    }

    ts := timeseries.NewTimeSeries(primitiveCreator, resolutions)
    assertNotNil(t, ts)
    assertEqual(t, len(ts.dataStreams), 2)
    assertEqual(t, ts.dataStreams[0].NumBuckets(), 0)

    i := Integer(1)
    ts.Add(&i, asTime(1))
    ts.Add(&i, asTime(1))
    ts.Add(&i, asTime(1))
    ts.Add(&i, asTime(2))
    ts.Add(&i, asTime(2))
    ts.Add(&i, asTime(3))
    ts.Add(&i, asTime(3))
    ts.Add(&i, asTime(3))
    ts.Add(&i, asTime(3))
    ts.Add(&i, asTime(3))
    assertEqual(t, ts.dataStreams[0].NumBuckets(), 3)

    total := ts.Total().(*Integer)
    assertEqual(t, total.Value(), 10)

    resolutionIdx := 0
    var rangeVals []timeseries.Primitive
    var bucketVal int

    rangeVals = ts.Range(resolutionIdx, asTime(1), asTime(1))
    assertEqual(t, len(rangeVals), 1)

    bucketVal = rangeVals[0].(*Integer).Value()
    assertEqual(t, bucketVal, 3)

    rangeVals = ts.Range(resolutionIdx, asTime(1), asTime(2))
    assertEqual(t, len(rangeVals), 2)

    bucketVal = rangeVals[0].(*Integer).Value()
    bucketVal += rangeVals[1].(*Integer).Value()
    assertEqual(t, bucketVal, 5)

    rangeVals = ts.Range(resolutionIdx, asTime(2), asTime(2))
    assertEqual(t, len(rangeVals), 1)

    bucketVal = rangeVals[0].(*Integer).Value()
    assertEqual(t, bucketVal, 2)

    rangeVals = ts.Range(resolutionIdx, asTime(2), asTime(3))
    assertEqual(t, len(rangeVals), 2)

    bucketVal = rangeVals[0].(*Integer).Value()
    bucketVal += rangeVals[1].(*Integer).Value()
    assertEqual(t, bucketVal, 7)

    resolutionIdx = 1
    rangeVals = ts.Range(resolutionIdx, asTime(1), asTime(3))
    assertEqual(t, len(rangeVals), 1)

    bucketVal = rangeVals[0].(*Integer).Value()
    assertEqual(t, bucketVal, 10)
}
```