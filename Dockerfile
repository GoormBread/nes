FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    libportaudio2 \
    libgl1-mesa-dev \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    portaudio19-dev \
    git \
    golang-go \
    xorg \
    x11-apps \
    ffmpeg \
    xvfb \
    libasound2-dev \
    pkg-config \
    pulseaudio \
    pulseaudio-utils \
    libpulse-dev

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
pacmd load-module module-null-sink sink_name=v1\n\
pacmd set-default-sink v1\n\
pacmd set-default-source v1.monitor" > pulseaudio-setup.sh && \
chmod +x pulseaudio-setup.sh

CMD ["bash", "-c", "./pulseaudio-setup.sh && Xvfb :1 -screen 0 768x768x24 & sleep 10 && DISPLAY=:1 ./nesexe ./rom/Super_mario_brothers.nes"]

EXPOSE 8080