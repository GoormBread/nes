FROM ubuntu:latest

# 필요한 패키지 설치
RUN apt-get update && apt-get install -y \
    libpulse-dev \
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
    xvfb \
    pulseaudio \
    supervisor

# 작업 디렉토리 설정
WORKDIR /app

# 애플리케이션 소스 코드 복사
COPY . .

RUN go get -u github.com/fogleman/nes
RUN go build -v -o nesexe

# supervisord 설정 파일 복사
COPY supervisord.conf /app/supervisord.conf

# 포트 노출
EXPOSE 8080

# supervisord 실행 (설정 파일 경로 지정)
CMD ["supervisord", "-c", "/app/supervisord.conf"]