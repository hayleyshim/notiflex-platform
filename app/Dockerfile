# ── 빌드 스테이지 ────────────────────────────────────
FROM golang:1.25-alpine AS build
WORKDIR /src

# 의존성 캐시 (go.mod만 먼저 복사)
COPY go.mod ./
RUN go mod download

# 소스 복사 후 정적 바이너리 빌드
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notiflex .

# ── 실행 스테이지 (scratch: 최소 이미지) ──────────────
FROM scratch
COPY --from=build /notiflex /notiflex
EXPOSE 8080
ENTRYPOINT ["/notiflex"]
