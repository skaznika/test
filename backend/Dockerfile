FROM --platform=$BUILDPLATFORM golang:bookworm AS ytdlp_cache
ARG TARGETOS
ARG TARGETARCH
RUN apt update && apt install -y wget
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp

FROM --platform=$BUILDPLATFORM golang:bookworm AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
RUN go build -o webwhisperplus .

FROM --platform=$BUILDPLATFORM python:3-bookworm
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY --from=builder /app/webwhisperplus ./webwhisperplus 
RUN chmod a+rx ./webwhisperplus

COPY --from=ytdlp_cache /usr/local/bin/yt-dlp /bin/yt-dlp

RUN mkdir /app/uploads

EXPOSE 3000
ENTRYPOINT [ "/app/webwhisperplus" ]