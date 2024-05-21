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
    golang \
    xorg \
    x11-apps \
    ffmpeg \
    xvfb


WORKDIR /app

COPY . .

RUN go get -u github.com/fogleman/nes && \
    go build -v -o nesexe


CMD ["bash", "-c", "Xvfb :1 -screen 0 1920x1080x24 & sleep 5 && DISPLAY=:1 ./nesexe ./rom/Super_mario_brothers.nes & ffmpeg -f x11grab -video_size 1920x1080 -framerate 30 -i :1 -c:v libx264 -preset ultrafast -qp 0 -f rtsp rtsp://mediamtx:8554/mystream"]
EXPOSE 8080