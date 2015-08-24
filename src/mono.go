package pifm

import (
	"math"
	"os"
)

type Mono struct {
	next *SampleSink
}

func int16ToFloat32Truncated(d interface{}) (float32, bool) {
	cur, cOk := d.(int16)
	if !cOk {
		return 0.0, false
	}

	fCur := float32(cur) / 32768.0
	return fCur, true
}

func (m *Mono) Consume(data []interface{}) {
	for i, n := 0, len(data)/2; i < n; i++ {
		fCur, fOk := int16ToFloat32Truncated(data[i])
		if fOk {
			m.next.Consume([]float32{fCur})
		}

	}
}

type StereoSplitter struct {
	nextLeft  *SampleSink
	nextRight *SampleSink
}

func (sp *StereoSplitter) Consume(data []interface{}) {
	// expects num % 4 == 0
	for i, n := 0, len(data)/2; i < n; i++ {
		lCur, lOk := int16ToFloat32Truncated(data[i])
		if lOk {
			m.nextLeft.Consume([]float32{lCur})
		}

		rCur, rOk := int16ToFloat32Truncated(data[i+1])
		if rOk {
			m.nextRight.Consume([]float32{rCur})
		}
	}
}

type RDSEncoder struct {
	sinLut    [8]float32
	next      *SampleSink
	bitNum    int
	lastBit   int
	time_     int
	lastValue float32
}

func newRDSEncoder(next *SampleSink) *RDSEncoder {
	rds := RDSEncoder{}
	for i, n := 0, len(rds.sinLut); i < n; i++ {
		rds.sinLut[i] = math.Sin(float32(i * 2 * math.Pi * 3 / 8))
	}

	return &rds
}

func (rds *RDSEncoder) Consume(data []interface{}) {
	for i, n := 0, len(data); i < n; i++ {
		if !rds.time_ {
			newBit = (RDSData[rds.bitNum/8] >> (7 - (rds.bitNum % 8))) & 1

			rds.lastBit ^= newBit
		}

		outputBit := rds.lastBit
		if rds.time_ >= 192 {
			outputBit = 1 - rds.lastBit
		}

		rds.lastValue = int(float32(rds.lastValue)*0.99 + ((float32(outputBit)*2 - 1) * 0.01))
		if cur, cOk := data[i].(float32); cOk {
			cur += rds.lastValue * rds.sinLut[rds.time_%8] * 0.05
			data[i] = cur
		}

		rds.time_ = (rds.time_) % 384
	}

	rds.Consume(data)
}

type StereoModulator struct {
	buffer          [1024]float32
	bufferOwner     int
	state           int // 8 state, state machine
	sinLut          [16]float32
	bufferItemCount uint

	next *SampleSink
}

type Channel int

type ModulatorInput struct {
	mod     *StereoModulator
	channel Channel
}

func (m *ModularInput) Consume(data []float32) {
	m.mod.Consume(data, m.channel)
}

func newStereoModulator(next *SampleSink) *StereoModulator {
	sm = &StereoModulator{}
	sm.init(next)
	return sm
}

func (sm *StereoModulator) init(next *SampleSink) {
	sm.next = next
	sm.state = 0
	sm.bufferOwner = 0
	sm.bufferItemCount = 0

	for i, n := 0, len(sm.sinLut); i < n; i++ {
		sm.sinLut[i] = math.Sin(float32(i) * 2.0 * math.Pi / 8)
	}

	return sm
}

func (sm *StereoModulator) getChannel(channel Channel) *SampleSink {
	return &modulator{mod: sm, channel: channel}
}

func (sm *StereoModulator) Consume(data []float32, channel Channel) {
	if channel == sm.bufferOwner || sm.bufferItemCount == 0 {
		sm.bufferOwner = channel

		for bufIter, n := 0, len(sm.buffer); n > 0; n, bufIter = n-1, bufIter+1 {
			sm.buffer[sm.bufferItemCount] = data[bufIter]
			sm.bufferItemCount += 1
		}
	} else {
		consumable := sm.bufferItemCount
		n := len(data)
		if sm.bufferItem >= n {
			consumable = n
		}

		var left, right = []float32{}, []float32{}

		if sm.bufferOwner == 0 {
			left = sm.buffer
		} else {
			left = data
		}

		if sm.bufferOwner == 1 {
			right = sm.buffer
		} else {
			right = data
		}

		for i := 0; i < n; i++ {
			sm.state = (sm.state + 1) % 8

			// Equation straight from Wikipedia
			sm.buffer[i] = ((left[i]+right[i])/2+(left[i]-right[i])/2*sm.sinLut[sm.state*2])*0.9 + 0.1*sm.sinLut[sm.state]
		}

		sm.next.Consume(sm.buffer)

		// Move stuff along buffer
		for i := consumable; i < sm.bufferItemCount; i++ {
			sm.buffer[i-consumable] = sm.buffer[i]
		}

		sm.bufferItemCount = consumable

		// Reconsume any remaining data
		restOfData := data[consumable:]
		smConsume(restOfData, channel)
	}
}

func playWav(filename string, sampleRate float32, stereo bool) error {
	fp := os.Stdin

	var ss *SampleSink

	if len(filename) >= 1 && filename[0] == '-' {
		ffp, err := os.Open(filename, os.O_RDONLY)
		if err != nil {
			return err
		}
	}

	defer fp.Close()

	if !stereo {
		ss = newMono(newPreEmp(sampleRate, newOutputter(sampleRate)))
	} else {
		sm := newStereoModulator(newRDSEncoder(newOutputter(152000)))

		ss = newStereoSplitter(
			// left
			newPreEmp(sampleRate, newResampe(sampleRate, 152000, sm.getChannel(0))),

			// right
			newPreEmp(sampleRate, newResampe(sampleRate, 152000, sm.getChannel(1))),
		)
	}

	headerConsume := [2]char{}
	for i, n := 0, 22; i < n; i++ {
		_, err := ReadAtLeast(fp, headerConsume, len(headerConsume)) // read past header
		if err != nil {
			fmt.Fprintf(os.Stderr, "playWav:: readPastHeader %d/%d: err %v\n", i+1, n, err)

			if err == io.EOF {
				break
			}
		}
	}

	data := [1024]char{}

	for {
		readBytes, err := io.ReadAtLeast(fp, data, 1024)
		if err != nil {
			fmt.Fprintf(os.Stderr, "playWav err %v\n", err)

			if err == io.EOF {
				break
			}
			continue
		}

		if readBytes < 1 {
			break
		}

		ss.Consume(data)
	}

	return nil
}
