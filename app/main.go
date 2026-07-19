package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// counter는 /id 요청마다 순차적으로 증가하는 인메모리 카운터이다.
var counter atomic.Uint64

// podName은 요청을 처리한 Pod의 이름이다.
// Kubernetes Downward API로 주입된 POD_NAME을 우선 사용하고,
// 없으면 컨테이너 호스트명(HOSTNAME)으로 대체한다.
func podName() string {
	if n := os.Getenv("POD_NAME"); n != "" {
		return n
	}
	if n := os.Getenv("HOSTNAME"); n != "" {
		return n
	}
	return "unknown"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// healthHandler는 서비스 상태를 반환한다. (readiness/liveness probe용)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"pod":    podName(),
	})
}

// idHandler는 순차 고유 ID를 생성하고, 생성한 Pod 이름을 함께 반환한다.
func idHandler(w http.ResponseWriter, r *http.Request) {
	id := counter.Add(1)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":  id,
		"pod": podName(),
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/id", idHandler)

	addr := ":8080"
	log.Printf("Notiflex API listening on %s (pod=%s)", addr, podName())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
