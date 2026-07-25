package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/valkey-io/valkey-go"
)

// version은 현재 실행 중인 Notiflex API의 버전이다.
const version = "v0.7.0"

// valkeyClient는 Pod 간 공유되는 중앙 카운터(Valkey)에 연결하는 클라이언트다.
var valkeyClient valkey.Client

// kafkaWriter는 /notify 요청을 비동기 처리 큐(Kafka)에 넣는 프로듀서다.
var kafkaWriter *kafka.Writer

const (
	idKey        = "notiflex:id"        // 순차 ID 카운터
	processedKey = "notiflex:processed" // 워커가 처리 완료한 알림 수
	notifyTopic  = "notifications"      // 비동기 알림 처리 토픽
)

// podName은 요청을 처리한 Pod의 이름이다.
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
// 10회(3초 간격) 재시도한다.
func connectValkey() valkey.Client {
	addr := os.Getenv("VALKEY_ADDR")

	// 비밀번호는 CSI로 마운트된 파일(VALKEY_PASSWORD_FILE)을 우선 사용하고,
	// 없으면 환경변수(VALKEY_PASSWORD)로 폴백한다.
	password := os.Getenv("VALKEY_PASSWORD")
	if pwFile := os.Getenv("VALKEY_PASSWORD_FILE"); pwFile != "" {
		if data, err := os.ReadFile(pwFile); err == nil {
			password = string(data)
			log.Printf("Valkey 비밀번호를 파일에서 로드: %s", pwFile)
		} else {
			log.Printf("VALKEY_PASSWORD_FILE 읽기 실패(%s), env로 폴백: %v", pwFile, err)
		}
	}

	var client valkey.Client
	var err error
	for i := 0; i < 10; i++ {
		client, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    password,
		})
		if err == nil {
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

// kafkaBrokers는 KAFKA_ADDR 환경변수에서 브로커 주소를 읽는다.
func kafkaBrokers() []string {
	addr := os.Getenv("KAFKA_ADDR")
	if addr == "" {
		addr = "kafka:9092"
	}
	return []string{addr}
}

// healthHandler는 서비스 상태를 반환한다. (readiness/liveness probe용)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "pod": podName()})
}

// versionHandler는 현재 API 버전과 처리한 Pod 이름을 반환한다.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": version, "pod": podName()})
}

// idHandler는 Valkey INCR로 순차 고유 ID를 생성한다.
func idHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id, err := valkeyClient.Do(ctx, valkeyClient.B().Incr().Key(idKey).Build()).AsInt64()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "valkey unavailable", "pod": podName()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "pod": podName()})
}

// notifyHandler는 알림 요청을 Kafka 큐에 넣고 즉시 202를 반환한다 (비동기).
// 실제 전송(느린 작업)은 워커가 뒤에서 처리하므로 요청이 몰려도 API는 빠르게 응답한다.
func notifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 최대 1MiB
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := kafkaWriter.WriteMessages(ctx, kafka.Message{Value: body}); err != nil {
		log.Printf("Kafka produce 실패: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "queue unavailable", "pod": podName()})
		return
	}
	// 202 Accepted: 접수만 하고 처리는 비동기
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued", "topic": notifyTopic, "pod": podName()})
}

// runAPI는 HTTP 서버 + Kafka 프로듀서를 구동한다 (기본 모드).
func runAPI() {
	valkeyClient = connectValkey()
	defer valkeyClient.Close()

	kafkaWriter = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaBrokers()...),
		Topic:                  notifyTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer kafkaWriter.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/id", idHandler)
	mux.HandleFunc("/version", versionHandler)
	mux.HandleFunc("/notify", notifyHandler)

	log.Printf("Notiflex API %s (mode=api) listening on :8080 (pod=%s)", version, podName())
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runWorker는 Kafka 컨슈머로 알림을 뒤에서 처리한다 (mode=worker).
func runWorker() {
	valkeyClient = connectValkey()
	defer valkeyClient.Close()

	// 워커도 probe용 /health를 노출한다.
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", healthHandler)
		_ = http.ListenAndServe(":8080", mux)
	}()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kafkaBrokers(),
		Topic:   notifyTopic,
		GroupID: "notiflex-workers", // 컨슈머 그룹: 워커를 늘리면 부하 분산
	})
	defer reader.Close()

	log.Printf("Notiflex worker %s (mode=worker) consuming topic=%s (pod=%s)", version, notifyTopic, podName())
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka consume 실패(재시도): %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		// 실제 알림 전송을 가정한 처리(느린 작업 시뮬레이션)
		time.Sleep(200 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		n, _ := valkeyClient.Do(ctx, valkeyClient.B().Incr().Key(processedKey).Build()).AsInt64()
		cancel()
		log.Printf("처리 완료 #%d: %s", n, string(m.Value))
	}
}

func main() {
	if os.Getenv("MODE") == "worker" {
		runWorker()
	} else {
		runAPI()
	}
}
