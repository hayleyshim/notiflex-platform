# 카프카 토픽 삭제 🔴

**위험도: 파괴적·비가역** — 토픽 삭제 시 그 안의 **모든 메시지가 영구 소실**되고,
아직 처리 못 한(consumer lag) 알림도 함께 사라진다. `notifications` 토픽을 지우면
비동기 처리 파이프라인이 끊긴다.

> 구성: 단일 KRaft 브로커(`app=kafka`, notiflex ns), CLI는 `/opt/kafka/bin/*.sh`.
> `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true`라 삭제 후 프로듀서가 다시 붙으면 **빈 토픽으로 재생성**된다(메시지는 복구 안 됨).

## 0. 컨텍스트

```bash
CTX=gke-sysnet4admin_book_gitaiops
KPOD=$(kubectl --context $CTX -n notiflex get pod -l app=kafka -o jsonpath='{.items[0].metadata.name}')
```

## 1. 사전 확인 (변경 전 현재 상태)

```bash
# 토픽 목록
kubectl --context $CTX -n notiflex exec $KPOD -- \
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list

# 대상 토픽 상세 (파티션·설정)
kubectl --context $CTX -n notiflex exec $KPOD -- \
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic <TOPIC>

# 미처리 메시지(lag) 확인 — lag>0이면 삭제 시 유실됨
kubectl --context $CTX -n notiflex exec $KPOD -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups
```

- [ ] 대상 토픽 이름이 정확한가
- [ ] `__consumer_offsets` 등 **내부 토픽은 절대 삭제 금지**
- [ ] lag>0(미처리 메시지)이면, 워커가 소진할 때까지 대기하거나 유실을 승인받았는가

## 2. 실행

```bash
kubectl --context $CTX -n notiflex exec $KPOD -- \
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic <TOPIC>
```

## 3. 검증

```bash
# 목록에서 사라졌는지
kubectl --context $CTX -n notiflex exec $KPOD -- \
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

## 4. 되돌리기 / 주의

- **메시지 복구 불가.** 되돌리기는 없다.
- 토픽 자체는 프로듀서가 다시 produce하면 auto-create로 **빈 상태로 재생성**된다.
  (즉시 재생성이 싫으면 프로듀서/워커를 먼저 멈춘다.)
- Kafka는 GitOps 매니페스트로 관리되지만 **토픽은 런타임 객체**라 ArgoCD가 되살리지 않는다.
