package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"net/http"
)

const sampleRate = 44100

type Settings struct {
	SampleRate float64
	Buffer     int
	Channels   int
}

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	var settings Settings
	settings.SampleRate = sampleRate
	device, err := portaudio.DefaultInputDevice()
	settings.Channels = device.MaxInputChannels
	settings.Buffer = 44100 * 20

	buffers := make([][]float32, settings.Channels)
	for i := range buffers {
		buffers[i] = make([]float32, settings.Buffer)
	}
	stream, err := portaudio.OpenDefaultStream(settings.Channels, 0, settings.SampleRate, settings.Buffer, func(in []float32) {
		go func(in []float32) {
			count := 0
			index := 0
			for j := range in {
				buffers[count][index] = in[j]
				if count == settings.Channels-1 {
					count = 0
					index++
				} else {
					count++
				}
			}
		}(in)
	})
	chk(err)
	err = stream.Start()
	chk(err)
	defer stream.Close()

	http.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		js, _ := json.Marshal(settings)
		w.Write(js)
	})

	for i := range buffers {
		http.HandleFunc(fmt.Sprintf("/audio_channel%d", i), func(w http.ResponseWriter, r *http.Request) {
			flusher, ok := w.(http.Flusher)
			if !ok {
				panic("expected http.ResponseWriter to be an http.Flusher")
			}

			w.Header().Set("Connection", "Keep-Alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Transfer-Encoding", "chunked")
			w.Header().Set("Content-Type", "audio/wave")
			for true {
				binary.Write(w, binary.BigEndian, &buffers[i])
				flusher.Flush() // Trigger "chunked" encoding and send a chunk...
				return
			}
		})
	}

	http.ListenAndServe(":5656", nil)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
