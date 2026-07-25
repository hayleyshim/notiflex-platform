package main

import (
	"context"
	"log"
	"os"

	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// tracer는 애플리케이션 span 생성기.
var tracer = otel.Tracer("notiflex")

// initTracer는 OTLP(gRPC) 익스포터로 Tempo에 span을 보내는 TracerProvider를 설정한다.
// OTEL_EXPORTER_OTLP_ENDPOINT가 없으면 트레이싱을 끈다(no-op) — 로컬/테스트 호환.
func initTracer(serviceName string) func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		log.Printf("OTEL_EXPORTER_OTLP_ENDPOINT 미설정 → 트레이싱 비활성")
		return func() {}
	}
	exp, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("OTLP 익스포터 생성 실패: %v (트레이싱 비활성)", err)
		return func() {}
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	// W3C traceparent로 컨텍스트 전파 (HTTP 헤더 + Kafka 헤더)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	log.Printf("트레이싱 활성: service=%s → %s", serviceName, endpoint)
	return func() { _ = tp.Shutdown(context.Background()) }
}

// kafkaHeaderCarrier는 Kafka 메시지 헤더에 trace context를 싣고 꺼낸다.
// 비동기 경계(API produce ↔ worker consume)를 하나의 트레이스로 잇기 위함.
type kafkaHeaderCarrier struct{ msg *kafka.Message }

func (c kafkaHeaderCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c kafkaHeaderCarrier) Set(key, val string) {
	for i := range c.msg.Headers {
		if c.msg.Headers[i].Key == key {
			c.msg.Headers[i].Value = []byte(val)
			return
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{Key: key, Value: []byte(val)})
}

func (c kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.msg.Headers))
	for _, h := range c.msg.Headers {
		keys = append(keys, h.Key)
	}
	return keys
}
