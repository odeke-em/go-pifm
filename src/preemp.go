package pifm

type PreEmp struct {
	fmConstant float32
	dataOld    float32
	next       *SampleSink
}

func newPreEmp(rate float32, next *SampleSink) *PreEmp {
	return &PreEmp{
		dataOld:    0.0,
		next:       next,
		fmConstant: rate * 75e-6,
	}
}

func (pe *PreEmp) Consume(data []float32) {
	for i, n := 0, len(data); i < n; i++ {
		cur = data[i]

		sample := cur + (pe.dataOld-cur)/(1-pe.fmConstant) // fir of 1 + s tau

		next.Consume([]float32{sample})

		pe.dataOld = cur
	}
}

const (
	Quality  = 5  // comp.complexity goes up linearly with this
	SQuality = 10 // start time quality (defines max phase error of filter vs ram used & cache thrashing)
	BufSize  = 1000
)

type QualityFloatArray [Quality]float32

type Resamp struct {
	dataOld     [Quality]float32
	sincLUT     [SQuality][Quality]float32
	ration      float32
	outTimeLeft float32
	outBuffer   [BufSize]float32
	outBufPtr   int
	next        *SampleSink
}

func newResamp(rateIn, rateOut float32, next *SampleSink) *Resamp {
	rs := Resamp{
		outTimeLeft: 1.0,
		outBufPtr:   0,
		ratio:       rateIn / rateOut,
		next:        next,
	}

	rs.init()

	return &rs
}

func (rs *Resamp) init() {
	for i := 0; i < Quality; i++ { // Sample
		for j := 0; j < SQuality; j++ { // starttime
			x := math.Pi * (float32(j)/SQuality + (Quality - 1 - i) - (Quality-1)/2.0)

			v := 1.0 // sin(0)/(0) == 1, says "their" limits theory
			if x != 0 {
				v = math.Sin(x) / x
			}

			rs.sincLUT[j][i] = v
		}
	}
}

func (rs *Resamp) Consume(data []float32) {
	for i, n := 0, len(data); i < n; i++ {
		// shift old data along

		for j, jEnd := 0, Quality-1; j < jEnd; j++ {
			rs.dataOld[j] = rs.dataOld[j+1]
		}

		// put in new sample
		rs.dataOld[Quality-1] = data[i]
		rs.outTimeLeft -= 1.0

		// go output this stuff!
		for rs.outTimeLeft < 1.0 {
			outSample = float32(0)

			lutNum := int(rs.outTimeLeft * SQuality)

			for j := 0; j < Quality; j++ {
				outSample += rs.dataOld[j] * rs.sincLUT[lutNum][j]
			}

			rs.outBuffer[rs.outBufPtr] = outSample
			rs.outBufPtr += 1
			rs.outTimeLeft += ratio

			// if we have lots of data, shunt it to the next stage

			if rs.outBufPtr >= BufSize {
				next.Consume(rs.outBuffer)
				rs.outBufPtr = 0
			}
		}
	}
}
