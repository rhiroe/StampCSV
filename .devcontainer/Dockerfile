FROM golang:1.24-bookworm

# Fyne が必要とする CGO 向けシステムライブラリ
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libgl1-mesa-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    xorg-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*
