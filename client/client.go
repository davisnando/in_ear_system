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
		buf := make([]byte, 512*5)
		_, _, err := pc.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		go func(buf []byte) {
			responseReader := bytes.NewReader(buf)
			binary.Read(responseReader, binary.LittleEndian, &frame.Buffer)
		}(buf)

	}
}

func main() {
	resp, err := http.Get(HOST + "/settings")
	chk(err)
	body, _ := ioutil.ReadAll(resp.Body)
	var settings Settings
	json.Unmarshal(body, &settings)

	resp, err = http.Get(HOST + "/CreateChannel")
	chk(err)
	body, _ = ioutil.ReadAll(resp.Body)
	type returnData struct {
		index            int
		amountOfChannels int
	}
	var data returnData
	json.Unmarshal(body, &data)

	portaudio.Initialize()
	defer portaudio.Terminate()
	frame.Buffer = make([]float32, settings.Buffer)
	go listen()

	var buffers [][]float32
	stream, err := portaudio.OpenDefaultStream(0, 1, settings.SampleRate, settings.Buffer, func(out []float32) {
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
