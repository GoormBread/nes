package ui

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
)

type Audio struct {
	sampleRate     float64
	outputChannels int
	channel        chan float32
	ffmpegStdin    io.WriteCloser
	ffmpegCmd      *exec.Cmd
}

func NewAudio() *Audio {
	a := Audio{}
	a.channel = make(chan float32, 4096)
	return &a
}

func (a *Audio) Start() error {
	// FFmpeg 프로세스를 시작합니다.
	cmd := exec.Command("ffmpeg",
		"-f", "f32le",          // 오디오 포맷
		"-ar", "44100",         // 오디오 샘플링 레이트
		"-ac", "1",             // 모노 오디오
		"-i", "pipe:0",         // 오디오 입력은 파이프를 통해
		"-f", "x11grab",        // 비디오 캡처 형식
		"-s", "768x768",        // 비디오 해상도
		"-i", ":1",             // X11 디스플레이 캡처
		"-map", "0:a:0",        // 첫 번째 입력의 오디오를 맵핑
		"-c:a", "libopus",      // 오디오 코덱을 libopus로 설정
		"-compression_level", "0", // 압축 레벨
		"-b:a", "24k",          // 오디오 비트레이트
		"-af", "aresample=async=0.7", // 오디오 리샘플링 필터
		"-map", "1:v:0",        // 두 번째 입력의 비디오를 맵핑
		"-r", "60",             // 비디오 프레임 레이트
		"-c:v", "libx264",      // 비디오 코덱을 libx264로 설정
		"-b:v", "1000k",        // 비디오 비트레이트
		"-f", "rtsp",           // 출력 포맷을 RTSP로 설정
		"rtsp://mtx:8554/mystream", // RTSP 스트림 URL
	)

	// FFmpeg의 출력과 오류를 터미널로 리다이렉트
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("FFmpeg stdin pipe 생성 실패: %v", err)
	}

	a.sampleRate = 44100.0
	a.outputChannels = 1
	a.ffmpegStdin = stdin
	a.ffmpegCmd = cmd

	// FFmpeg 명령 실행
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("FFmpeg 실행 실패: %v", err)
	}

	go a.play()

	// FFmpeg 프로세스가 종료될 때까지 대기
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("FFmpeg 실행 중 에러 발생: %v", err)
		}
	}()

	return nil
}

func (a *Audio) play() {
	defer a.ffmpegStdin.Close()

	for {
		// 오디오 데이터를 채널에서 읽어서 FFmpeg stdin에 씁니다.
		buf := make([]float32, 1024)
		for i := range buf {
			select {
			case sample := <-a.channel:
				buf[i] = sample
			default:
				buf[i] = 0
			}
		}

		// []float32를 []byte로 변환하여 FFmpeg stdin에 씁니다.
		byteBuf := make([]byte, len(buf)*4)
		for i, sample := range buf {
			binary.LittleEndian.PutUint32(byteBuf[i*4:], math.Float32bits(sample))
		}

		_, err := a.ffmpegStdin.Write(byteBuf)
		if err != nil {
			fmt.Printf("데이터 쓰기 실패: %v\n", err)
			break
		}
	}
}

func (a *Audio) Stop() error {
	if a.ffmpegStdin != nil {
		a.ffmpegStdin.Close() // FFmpeg stdin을 닫아 프로세스를 종료
	}
	if a.ffmpegCmd != nil {
		err := a.ffmpegCmd.Wait() // FFmpeg 프로세스 종료 대기
		if err != nil {
			return fmt.Errorf("FFmpeg 종료 중 에러 발생: %v", err)
		}
	}
	return nil
}