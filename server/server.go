package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"net/http"
	// "strconv"
)

const sampleRate = 44100

type Settings struct {
	SampleRate float64
	Buffer     int
	Channels   int
}

func main() {
	var MasterMix Master
	MasterMix.InitializePortaudio()
	defer portaudio.Terminate()

	var settings Settings
	device, _ := portaudio.DefaultInputDevice()

	settings.SampleRate = sampleRate
	// settings.Channels = 1
	settings.Channels = device.MaxInputChannels
	settings.Buffer = 512
	MasterMix.Setting = settings
	MasterMix.Init()
	MasterMix.handleBuffers()

	http.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		js, _ := json.Marshal(settings)
		w.Write(js)
	})

	http.HandleFunc(fmt.Sprintf("/audio_channel"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "audio/wave")
		binary.Write(w, binary.LittleEndian, &MasterMix.Main.Mono)
		return
	})

	http.ListenAndServe(":5656", nil)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
