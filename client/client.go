package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Settings struct {
	SampleRate float64
	Buffer     int
	Channels   int
}

type Frame struct {
	Buffer []float32
}

const HOST = "http://localhost:5656"

var frame Frame

func listen() {
	pc, _ := net.ListenUDP("udp", &net.UDPAddr{
		Port: 4444,
		IP:   net.ParseIP("0.0.0.0"),
	})
	defer pc.Close()
	for {
		buf := make([]byte, 512*4)
		_, _, err := pc.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		responseReader := bytes.NewReader(buf)
		binary.Read(responseReader, binary.LittleEndian, &frame.Buffer)
	}
}

func main() {
	resp, err := http.Get(HOST + "/settings")
	chk(err)
	body, _ := ioutil.ReadAll(resp.Body)
	var settings Settings
	json.Unmarshal(body, &settings)
	portaudio.Initialize()
	defer portaudio.Terminate()
	frame.Buffer = make([]float32, settings.Buffer)
	go listen()

	var buffers [][]float32
	stream, err := portaudio.OpenDefaultStream(0, 1, settings.SampleRate, settings.Buffer, func(out []float32) {
		// buffer := make([]float32, settings.Buffer)
		// resp, err := http.Get(fmt.Sprintf("%s/channel0", HOST))
		// chk(err)
		// body, _ := ioutil.ReadAll(resp.Body)
		// responseReader := bytes.NewReader(body)
		// binary.Read(responseReader, binary.LittleEndian, &buffer)
		copy(out, frame.Buffer)
		buffers = append(buffers, out)
	})

	chk(err)
	chk(stream.Start())
	time.Sleep(time.Second * 400)
	chk(stream.Stop())
	defer stream.Close()

	if err != nil {
		fmt.Println(err)
	}

}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
