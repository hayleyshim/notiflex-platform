# ── 빌드 스테이지 ────────────────────────────────────
FROM golang:1.25-alpine AS build
WORKDIR /src

# 소스 전체 복사 후 go mod tidy로 의존성/go.sum 해결
# (로컬에 Go가 없어 go.sum을 만들 수 없으므로 빌드 컨테이너에서 생성)
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notiflex .

# ── 실행 스테이지 (scratch: 최소 이미지) ──────────────
FROM scratch
COPY --from=build /notiflex /notiflex
EXPOSE 8080
ENTRYPOINT ["/notiflex"]
