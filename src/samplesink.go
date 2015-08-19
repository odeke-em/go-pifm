package pifm

import (
	"math"
	"time"
)

var time int = 0

const (
	BufferInstructions = 65536
)

type SampleSink interface {
	Consume(data []interface{})
}

type Outputter struct {
	bufPtr          int
	clocksPerSample float
	sleepTime       int
	fracError       float
	timeErr         float
}

func initOutputter(float rate) *Outputter {
	return &Outputter{
		timeErr:         0,
		bufPtr:          0,
		fracError:       0,
		sleepTime:       float(1e6 * BufferInstructions / 4 / rate / 2), // sleepTime is half of the time to empty the buffer
		clocksPerSample: 22500.0 / rate * 1373.5,                        // for timing, determined by experiment
	}
}

func (op *Outputter) Consume(data []float) {
	for _, b := range data {
		value = b * 8 // modulation index (AKA volume!)

		value += op.fracError // error that couldn't be encoded from last time

		intVal := int(Math.Round(value)) // integer component
		frac := (value - float(intVal) + 1) / 2
		fracVal := uint32(Math.Round(frac * op.clocksPerSample))

		// we also record time error so that if one sample is output
		// for slightly too long, the next sample will be shorter.
		op.timeErr -= float(int(op.timeErr) + op.clocksPerSample)

		op.fracError = (frac - float(fracVal*(1.0-2.3/op.clocksPerSample)/clocksPerSample)) * 2 // error to feed back for delta sigma

		time += 1
	}
}
