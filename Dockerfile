FROM ubuntu:latest

# 필요한 패키지 설치
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
    golang \
    xorg \
    x11-apps \
    ffmpeg \
    xvfb

# 작업 디렉토리 설정
WORKDIR /app

# 애플리케이션 소스 코드 복사
COPY . .

# Go 애플리케이션 빌드
RUN go get -u github.com/fogleman/nes && \
    go build -v -o nesexe

# 환경 변수 설정
ENV DISPLAY=:1
# ENV RTSP_URL=rtsp://mtx:8554/mystream

# 애플리케이션 및 FFmpeg 명령어 실행
CMD ["bash", "-c", "Xvfb :1 -screen 0 768x768x24 & sleep 5 && DISPLAY=:1 ./nesexe ./rom/Super_mario_brothers.nes & ffmpeg -f x11grab -i :1 -map 0:v:0 -c:v libx264 -preset ultrafast -tune zerolatency -r 151 -b:v 1000k -s 1024x768 -f rtsp rtsp://localhost:8554/mystream"]

# 포트 노출
EXPOSE 8080