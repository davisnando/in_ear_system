package main

import (
	"github.com/gordonklaus/portaudio"
	"math"
)

type Master struct {
	Channels Buffers
	Main     Buffer
	Setting  Settings
}

type Buffer struct {
	Mono   []float32
	Temp   []float32
	Index  int
	Volume float32
}

type Buffers []Buffer

func (m *Master) InitializePortaudio() {
	portaudio.Initialize()
}

func (m *Master) Init() {
	m.Channels = make(Buffers, m.Setting.Channels)
	for i := range m.Channels {
		m.Channels[i].Mono = make([]float32, m.Setting.Buffer)
		m.Channels[i].Temp = make([]float32, m.Setting.Buffer)
		m.Channels[i].Index = i
		m.Channels[i].Volume = 1
	}
	m.Main.Mono = make([]float32, m.Setting.Buffer)
	m.Main.Temp = make([]float32, m.Setting.Buffer)
	m.Main.Volume = 1
}

func (m *Master) handleBuffers() {
	stream, err := portaudio.OpenDefaultStream(m.Setting.Channels, 0, m.Setting.SampleRate, m.Setting.Buffer, func(in []float32) {
		for i := 0; i < m.Setting.Buffer; i++ {
			for b := range m.Channels {
				m.Channels[b].Mono[i] = in[i*m.Setting.Channels+m.Channels[b].Index]
			}
		}
		m.Mix()
	})
	err = stream.Start()
	chk(err)
}
func (m *Master) audioMix(a float32, b float32) float32 {
	if a+b > 1 {
		if a > b {
			return a
		}
		return b
	}
	if a+b < -1 {
		if a < b {
			return a
		}
		return b
	}
	return 0.7 * (a + b)
}
func (m *Master) Mix() {
	m.Main.Temp = make([]float32, m.Setting.Buffer)
	for _, buffer := range m.Channels {
		for i := range buffer.Mono {
			if m.Main.Temp[i] == 0 {
				m.Main.Temp[i] = buffer.Volume * buffer.Mono[i]
			} else {
				m.Main.Temp[i] = m.audioMix(buffer.Volume*buffer.Mono[i], m.Main.Temp[i])
			}
		}
	}

	for i, buffer := range m.Main.Temp {
		m.Main.Temp[i] = buffer * m.Main.Volume
	}
	copy(m.Main.Mono, m.Main.Temp)
}

func (m *Master) mixLogarithmicRangeCompression(i float32) float32 {
	if i < -1 {
		return float32(-math.Log(-float64(i)-0.85)/14 - 0.75)
	} else if i > 1 {
		return float32(math.Log(float64(i)-0.85)/14 + 0.75)
	} else {
		return i
	}
}
