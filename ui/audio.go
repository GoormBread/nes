// package ui

// import "github.com/gordonklaus/portaudio"

// type Audio struct {
// 	stream         *portaudio.Stream
// 	sampleRate     float64
// 	outputChannels int
// 	channel        chan float32
// }

// func NewAudio() *Audio {
// 	a := Audio{}
// 	a.channel = make(chan float32, 44100)
// 	return &a
// }

// func (a *Audio) Start() error {
// 	host, err := portaudio.DefaultHostApi()
// 	if err != nil {
// 		return err
// 	}
// 	parameters := portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
// 	stream, err := portaudio.OpenStream(parameters, a.Callback)
// 	if err != nil {
// 		return err
// 	}
// 	if err := stream.Start(); err != nil {
// 		return err
// 	}
// 	a.stream = stream
// 	a.sampleRate = parameters.SampleRate
// 	a.outputChannels = parameters.Output.Channels
// 	return nil
// }

// func (a *Audio) Stop() error {
// 	return a.stream.Close()
// }

//	func (a *Audio) Callback(out []float32) {
//		var output float32
//		for i := range out {
//			if i%a.outputChannels == 0 {
//				select {
//				case sample := <-a.channel:
//					output = sample
//				default:
//					output = 0
//				}
//			}
//			out[i] = output
//		}
//	}
package ui

import (
	"encoding/binary"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

// Audio struct represents the audio stream
type Audio struct {
	stream         *portaudio.Stream
	sampleRate     float64
	outputChannels int
	channel        chan float32
	mu             sync.Mutex // Mutex for channel access
}

// NewAudio creates a new Audio instance
func NewAudio() *Audio {
	a := Audio{}
	a.channel = make(chan float32, 44100)
	return &a
}

// Start starts the audio stream
func (a *Audio) Start() error {
	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return err
	}
	parameters := portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	stream, err := portaudio.OpenStream(parameters, a.Callback)
	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		return err
	}
	a.stream = stream
	a.sampleRate = parameters.SampleRate
	a.outputChannels = parameters.Output.Channels
	return nil
}

// Stop stops the audio stream
func (a *Audio) Stop() error {
	return a.stream.Close()
}

// Callback is the callback function for the audio stream
func (a *Audio) Callback(out []float32) {
	var output float32
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range out {
		if i%a.outputChannels == 0 {
			select {
			case sample := <-a.channel:
				output = sample
			default:
				output = 0
			}
		}
		out[i] = output
	}
}

// WsHandler handles WebSocket connections
// WsHandler handles WebSocket connections
func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	audio := NewAudio()
	if err := audio.Start(); err != nil {
		log.Printf("Failed to start audio: %v", err)
		return
	}
	defer audio.Stop()

	buf := make([]byte, 4) // float32 크기의 버퍼

	for {
		audio.mu.Lock()
		audioData := <-audio.channel
		audio.mu.Unlock()

		// float32를 []byte로 변환하여 전송
		binary.LittleEndian.PutUint32(buf, math.Float32bits(audioData))
		if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
			log.Printf("Failed to write audio data to WebSocket: %v", err)
			break
		}
	}
}
