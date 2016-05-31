## go-bucket
Golang Timeseries Data Bucket

## Example
```
type Integer int

func NewInteger() Primitive                 { i := Integer(0); return &i }
func (i *Integer) Value() int               { return int(*i) }
func (i *Integer) Add(other Primitive)      { *i += *(other.(*Integer)) }
func (i *Integer) CopyFrom(other Primitive) { *i = *(other.(*Integer)) }
func (i *Integer) Reset()                   { *i = 0 }

func TestTimeSeries() {
    primitiveCreator := NewInteger
    resolutions := []time.Duration{
        ResolutionOneSecond,
        ResolutionTenSeconds,
    }

    timeSeries := NewTimeSeries(primitiveCreator, resolutions)
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
}
```