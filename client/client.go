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

const HOST = "http://localhost:5656"

func main() {
	resp, err := http.Get(HOST + "/settings")
	chk(err)
	body, _ := ioutil.ReadAll(resp.Body)
	var settings Settings
	json.Unmarshal(body, &settings)
	portaudio.Initialize()
	defer portaudio.Terminate()
	buffer := make([]float32, settings.Buffer)

	stream, err := portaudio.OpenDefaultStream(0, 1, settings.SampleRate, len(buffer), func(out []float32) {
		buffers := make([][]float32, settings.Channels)
		for i := 0; i < settings.Channels; i++ {
			buffers[i] = make([]float32, settings.Buffer)
			resp, err := http.Get(fmt.Sprintf("%s/audio_channel%d", HOST, i))
			chk(err)

			body, _ := ioutil.ReadAll(resp.Body)
			responseReader := bytes.NewReader(body)
			binary.Read(responseReader, binary.BigEndian, &buffers[i])
		}
		for i := range out {
			out[i] = mix(buffers[0][i], buffers[1][i])
		}
	})

	chk(err)
	chk(stream.Start())
	time.Sleep(time.Second * 40)
	chk(stream.Stop())
	defer stream.Close()

	if err != nil {
		fmt.Println(err)
	}

}

func mix(a float32, b float32) float32 {
	if a < 0 && b < 0 {
		return (a + b) - ((a * b) / -32767)
	} else if a > 0 && b > 0 {
		return (a + b) - ((a * b) / -32767)
	}
	return a + b
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
