FROM nvidia/cuda:12.4.0-base-ubuntu22.04

# 필요한 패키지 설치
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    git \
    golang-go \
    xorg \
    x11-apps \
    xvfb \
    libasound2-dev \
    pkg-config \
    pulseaudio \
    pulseaudio-utils \
    libpulse-dev

# ffmpeg 설치
RUN apt-get install -y ffmpeg

# NVIDIA 드라이버 설치
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility,video

WORKDIR /app
COPY . .
RUN go mod tidy && \
    go mod download

# 필요한 모듈 설치
RUN go get github.com/go-gl/gl/v2.1/gl && \
    go get github.com/go-gl/glfw/v3.2/glfw && \
    go get github.com/mesilliac/pulse-simple

RUN go build -v -o nesexe

ENV DISPLAY=:1
ENV RTSP_URL=rtsp://mtx:8554/mystream
ENV PULSE_LATENCY_MSEC=1

RUN echo "#!/bin/bash\n\
pulseaudio -D --exit-idle-time=-1 &\n\
sleep 5\n\
pacmd load-module module-null-sink sink_name=v1 rate=44100 channels=1\n\
pacmd set-default-sink v1\n\
pacmd set-default-source v1.monitor" > pulseaudio-setup.sh && \
chmod +x pulseaudio-setup.sh

# ffmpeg에서 GPU 가속 사용을 위한 옵션 추가
CMD ["bash", "-c", "./pulseaudio-setup.sh && Xvfb :1 -screen 0 768x768x24 & sleep 10 && DISPLAY=:1 ./nesexe \"./rom/${GAME}.nes\" & sleep 10 && ffmpeg -hwaccel cuda -f pulse -i default -f x11grab -s 768x768 -i :1 -map 0:a:0 -c:a libopus -compression_level 0 -b:a 24k -af aresample=async=0.7 -map 1:v:0 -r 60 -c:v h264_nvenc -preset p2 -tune ll -b:v 1000k -f rtsp rtsp://localhost:8554/mystream"]
EXPOSE 8080
