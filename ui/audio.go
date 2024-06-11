package ui

import (
   "encoding/binary"
   "fmt"
   "math"
   "github.com/mesilliac/pulse-simple"
)

type Audio struct {
   stream         *pulse.Stream
   sampleRate     float64
   outputChannels int
   channel        chan float32
}

func NewAudio() *Audio {
   a := Audio{}
   a.channel = make(chan float32, 4096)
   return &a
}

func (a *Audio) Start() error {
   // PulseAudio 스트림 생성
   ss := pulse.SampleSpec{
      Format:   pulse.SAMPLE_FLOAT32LE,
      Rate:     44100,
      Channels: 1,
   }
   stream, err := pulse.Playback("Simple Playback", "Audio Stream", &ss)
   if err != nil {
      return fmt.Errorf("PulseAudio 스트림 생성 실패: %v", err)
   }
   a.stream = stream
   a.sampleRate = float64(ss.Rate)
   a.outputChannels = int(ss.Channels)

   go a.play()

   return nil
}

func (a *Audio) play() {
   for {
      // 오디오 데이터를 채널에서 읽어서 스트림에 씁니다.
      buf := make([]float32, 1024)
      for i := range buf {
         select {
         case sample := <-a.channel:
            buf[i] = sample
         default:
            buf[i] = 0
         }
      }

      // []float32를 []byte로 변환
      byteBuf := make([]byte, len(buf)*4)
      for i, sample := range buf {
         binary.LittleEndian.PutUint32(byteBuf[i*4:], math.Float32bits(sample))
      }

      a.stream.Write(byteBuf)
   }
}

func (a *Audio) Stop() error {
   if a.stream != nil {
      a.stream.Free()
   }
   return nil
}