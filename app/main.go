package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/valkey-io/valkey-go"
)

// version은 현재 실행 중인 Notiflex API의 버전이다.
const version = "v0.4.0"

// valkeyClient는 Pod 간 공유되는 중앙 카운터(Valkey)에 연결하는 클라이언트다.
var valkeyClient valkey.Client

// idKey는 Valkey에서 순차 ID를 저장하는 키다. 모든 Pod이 같은 키를 INCR한다.
const idKey = "notiflex:id"

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

// connectValkey는 Valkey에 연결한다. Pod 시작 시 DNS/Valkey 기동 지연에 대비해
// 10회(3초 간격) 재시도한다. 재시도가 없으면 CrashLoopBackOff에 빠질 수 있다.
func connectValkey() valkey.Client {
	addr := os.Getenv("VALKEY_ADDR")
	password := os.Getenv("VALKEY_PASSWORD")

	var client valkey.Client
	var err error
	for i := 0; i < 10; i++ {
		client, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    password,
		})
		if err == nil {
			// 실제 연결 확인 (PING)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			perr := client.Do(ctx, client.B().Ping().Build()).Error()
			cancel()
			if perr == nil {
				log.Printf("Valkey 연결 성공 (%s)", addr)
				return client
			}
			err = perr
			client.Close()
		}
		log.Printf("Valkey 연결 재시도 %d/10: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("Valkey 연결 실패 (10회 시도): %v", err)
	return nil
}

// healthHandler는 서비스 상태를 반환한다. (readiness/liveness probe용)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"pod":    podName(),
	})
}

// versionHandler는 현재 API 버전과 처리한 Pod 이름을 반환한다.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version": version,
		"pod":     podName(),
	})
}

// idHandler는 Valkey INCR로 순차 고유 ID를 생성한다.
// 모든 Pod이 같은 키(notiflex:id)를 원자적으로 증가시키므로,
// Pod이 여러 개여도 ID가 중복되지 않고 순차적으로 발급된다.
func idHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id, err := valkeyClient.Do(ctx, valkeyClient.B().Incr().Key(idKey).Build()).AsInt64()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "valkey unavailable",
			"pod":   podName(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":  id,
		"pod": podName(),
	})
}

func main() {
	valkeyClient = connectValkey()
	defer valkeyClient.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/id", idHandler)
	mux.HandleFunc("/version", versionHandler)

	addr := ":8080"
	log.Printf("Notiflex API %s listening on %s (pod=%s)", version, addr, podName())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
