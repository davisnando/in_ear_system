package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"io/ioutil"
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

func main() {
	resp, err := http.Get(HOST + "/settings")
	chk(err)
	body, _ := ioutil.ReadAll(resp.Body)
	var settings Settings
	json.Unmarshal(body, &settings)
	portaudio.Initialize()
	defer portaudio.Terminate()
	var buffers [][]float32
	stream, err := portaudio.OpenDefaultStream(0, 1, settings.SampleRate, settings.Buffer, func(out []float32) {
		buffer := make([]float32, settings.Buffer)
		resp, err := http.Get(fmt.Sprintf("%s/audio_channel", HOST))
		chk(err)
		body, _ := ioutil.ReadAll(resp.Body)
		responseReader := bytes.NewReader(body)
		binary.Read(responseReader, binary.LittleEndian, &buffer)
		copy(out, buffer)
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

func mix(a float32, b float32) float32 {
	return ((a + b) - (a * b)) * 2
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
