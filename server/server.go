package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/davisnando/in_ear_system/portaudio"
	"net"
	"net/http"
	"strconv"
)

const sampleRate = 44100

type Settings struct {
	SampleRate float64
	Buffer     int
	Channels   int
}

type Message struct {
	Message string
	Failed  bool
}

func main() {
	fmt.Println("STARTING SERVER....")
	var MasterMix Master
	fmt.Println("Initializing audio")
	MasterMix.InitializePortaudio()
	defer portaudio.Terminate()

	var settings Settings
	device, _ := portaudio.DefaultInputDevice()

	settings.SampleRate = sampleRate
	// settings.Channels = 1
	settings.Channels = device.MaxInputChannels

	settings.Buffer = 512
	MasterMix.Setting = settings
	fmt.Println("Setting up http server")
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
	http.HandleFunc(fmt.Sprintf("/CreateChannel"), func(w http.ResponseWriter, r *http.Request) {
		index := MasterMix.CreateChannel()
		type returnData struct {
			index            int
			amountOfChannels int
		}
		var data returnData
		data.index = index
		data.amountOfChannels = len(MasterMix.MasterBuffer)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		MasterMix.Mixes[index].Ips = append(MasterMix.Mixes[index].Ips, ip)
		w.Header().Set("Content-Type", "application/json")
		fmt.Println(ip)
		js, _ := json.Marshal(settings)
		w.Write(js)
	})

	http.HandleFunc(fmt.Sprintf("/SetVolume"), func(w http.ResponseWriter, r *http.Request) {
		var returnData Message
		w.Header().Set("Content-Type", "application/json")
		rawmix, ok := r.URL.Query()["mix"]
		if !ok {
			message(w, "Parameters incorrect", true)
			return
		}
		mix, err := strconv.ParseInt(rawmix[0], 10, 64)
		if err != nil {
			message(w, "Mix is not an integer", true)
			return
		}
		rawchannel, ok := r.URL.Query()["channel"]
		if !ok {
			message(w, "Parameters incorrect", true)
			return
		}
		channel, err := strconv.ParseInt(rawchannel[0], 10, 64)
		if err != nil {
			message(w, "Channel is not an integer", true)
			return
		}
		rawvolume, ok := r.URL.Query()["volume"]
		if !ok {
			message(w, "Parameters incorrect", true)
			return
		}
		volume, err := strconv.ParseFloat(rawvolume[0], 32)
		if err != nil {
			message(w, "Volume is not an integer", true)
			return
		}
		if int(mix) >= len(MasterMix.Mixes) {
			message(w, "Mix out of index", true)
			return
		}
		if int(channel) >= len(MasterMix.MasterBuffer) {
			message(w, "Channels out of index", true)
			return
		}
		if volume < 0 || volume > 1 {
			message(w, "Volume needs to be between 0 and 1", true)
			return
		}
		MasterMix.Mixes[mix].Channels[channel].volume = float32(volume)
		returnData.Failed = false
		returnData.Message = "test"
		js, _ := json.Marshal(returnData)
		w.Write(js)
	})
	fmt.Println("Serving on 0.0.0.0:5656")
	http.ListenAndServe("0.0.0.0:5656", nil)
}

func message(w http.ResponseWriter, message string, failed bool) {
	var msg Message
	msg.Failed = failed
	msg.Message = message
	js, _ := json.Marshal(msg)
	w.Write(js)
}

func chk(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
