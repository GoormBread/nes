package ui

import (
    "github.com/mesilliac/pulse-simple"
)

type Audio struct {
    stream     *pulse.Stream
    sampleRate float64
    channel    chan float32
}

func NewAudio() *Audio {
    a := Audio{}
    a.channel = make(chan float32, 44100)
    return &a
}

func (a *Audio) Start() error {
    ss := pulse.SampleSpec{
        Format: pulse.SAMPLE_FLOAT32LE,
        Rate:   44100,
        Channels: 2,
    }

    stream, err := pulse.Playback("mystream", "PulseAudio", &ss)
    if err != nil {
        return err
    }

    a.stream = stream
    a.sampleRate = float64(ss.Rate)

    return nil
}

func (a *Audio) Stop() error {
    if a.stream != nil {
        a.stream.Drain()
        a.stream.Free()
        a.stream = nil
    }
    return nil
}

func (a *Audio) Callback(out []float32) {
    var output float32
    for i := range out {
        if i%2 == 0 {
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