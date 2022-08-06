package main

import (
	"encoding/binary"
	"github.com/davisnando/in_ear_system/portaudio"
	"net"
)

type Buffer struct {
	Mono   []int16
	Temp   []int16
	Index  int
	Volume int16
}

type Buffers []Buffer

type Master struct {
	MasterBuffer Buffers
	Main         Buffer
	Setting      Settings
	Mixes        []Mix
}

type Channel struct {
	buffer Buffer
	volume float32
}

type Mix struct {
	Channels []Channel
	Out      Buffer
	index    int
	Ips      []string
}

func (m *Mix) Init(settings Settings) {
	m.Channels = make([]Channel, settings.Channels)
	m.Out.Mono = make([]int16, settings.Buffer)
	m.Out.Temp = make([]int16, settings.Buffer)
	m.Out.Volume = 1
	for i := range m.Channels {
		m.Channels[i].buffer.Mono = make([]int16, settings.Buffer)
		m.Channels[i].buffer.Temp = make([]int16, settings.Buffer)
		m.Channels[i].volume = 1
	}
}
func (m *Mix) Send() {
	for _, ip := range m.Ips {

		go func(dstIP string) {
			RemoteEP := net.UDPAddr{IP: net.ParseIP(dstIP), Port: 4444}
			conn, err := net.DialUDP("udp", nil, &RemoteEP)
			chk(err)
			if err != nil {
				return
			}
			binary.Write(conn, binary.LittleEndian, &m.Out.Mono)
			conn.Close()
		}(ip)
	}

}

func (m *Mix) Mix() {
	m.Out.Temp = make([]int16, len(m.Out.Temp))
	for _, channel := range m.Channels {
		for i := range channel.buffer.Mono {
			if m.Out.Temp[i] == 0 {
				m.Out.Temp[i] = int16(channel.volume * float32(channel.buffer.Mono[i]))
			} else {
				m.Out.Temp[i] = audioMix(int16(channel.volume*float32(channel.buffer.Mono[i])), m.Out.Temp[i])
			}
		}
	}

	for i, buffer := range m.Out.Temp {
		m.Out.Temp[i] = buffer * m.Out.Volume
	}
	copy(m.Out.Mono, m.Out.Temp)
	m.Send()
}

func (m *Master) CreateChannel() int {
	var mix Mix
	mix.index = len(m.Mixes)
	mix.Init(m.Setting)
	m.Mixes = append(m.Mixes, mix)
	return mix.index
}

func (m *Master) InitializePortaudio() {
	portaudio.Initialize()
}

func (m *Master) Init() {
	m.MasterBuffer = make(Buffers, m.Setting.Channels)
	for i := range m.MasterBuffer {
		m.MasterBuffer[i].Mono = make([]int16, m.Setting.Buffer)
		m.MasterBuffer[i].Temp = make([]int16, m.Setting.Buffer)
		m.MasterBuffer[i].Index = i
		m.MasterBuffer[i].Volume = 1
	}
	m.Main.Mono = make([]int16, m.Setting.Buffer)
	m.Main.Temp = make([]int16, m.Setting.Buffer)
	m.Main.Volume = 1
}

func (m *Master) handleBuffers() {
	stream, err := portaudio.OpenDefaultStream(m.Setting.Channels, 0, m.Setting.SampleRate, m.Setting.Buffer, func(in []int16) {
		for i := 0; i < m.Setting.Buffer; i++ {
			for b := range m.MasterBuffer {
				m.MasterBuffer[b].Mono[i] = in[i*m.Setting.Channels+m.MasterBuffer[b].Index]
			}
		}
		for i := range m.Mixes {
			for b, buffer := range m.MasterBuffer {
				copy(m.Mixes[i].Channels[b].buffer.Mono, buffer.Mono)
				
			}
                        m.Mixes[i].Mix()
		}
		m.Mix()
	})
	chk(err)
	err = stream.Start()
	chk(err)
}

func (m *Master) Mix() {
	m.Main.Temp = make([]int16, m.Setting.Buffer)
	for _, buffer := range m.MasterBuffer {
		for i := range buffer.Mono {
			if m.Main.Temp[i] == 0 {
				m.Main.Temp[i] = buffer.Volume * buffer.Mono[i]
			} else {
				m.Main.Temp[i] = audioMix(buffer.Volume*buffer.Mono[i], m.Main.Temp[i])
			}
		}
	}

	for i, buffer := range m.Main.Temp {
		m.Main.Temp[i] = buffer * m.Main.Volume
	}
	copy(m.Main.Mono, m.Main.Temp)
}
