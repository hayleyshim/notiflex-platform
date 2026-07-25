# Notiflex 회고 (ch9)

> 3장~8장의 여정을 돌아보고, 쌓인 부채와 배움을 정리한다.
> 상세 결정은 [architecture-decisions.md](architecture-decisions.md), 진행 기록은 [../JOURNEY.md](../JOURNEY.md),
> 현재 아키텍처는 [../claude-context/architecture.md](../claude-context/architecture.md) 참조.

## 1. 여정의 궤적

작은 Go 알림 API 하나가 6단계에 걸쳐 운영 플랫폼으로 진화했다.

| 축 | 시작 | 현재 |
|----|------|------|
| 배포 | 수동 → Rolling | Canary(20/50/80) + Argo Rollouts |
| GitOps | 단일 Application | App of Apps + sync-wave + ApplicationSet |
| 노드 | 단일 풀 | 역할별 3풀(app/data/platform), taint 격리 |
| 확장 | 단일 테넌트 | 멀티테넌시(고객별 ns, 요금제 오버레이) |
| 처리 | 동기 | Kafka 비동기(API 202 즉시 / 워커 소비) |
| 관측 | 메트릭만 | 메트릭 + 추적(OTel→Tempo) |

일관 원칙: **"도구 교체가 아닌 점진 진화"** — Rollout CRD를 유지하며 strategy만, ApplicationSet generator만 바꿨다.

## 2. 저장소 규모 (ch9 시점)

- 46개 파일 · 커밋 66개 · ADR 16건 · 완료 서브챕터 26개
- Go 341줄(stdlib only, scratch 이미지) / YAML 1,211줄 / 문서 818줄
- 저작: 사람 82% + CI 봇 18%(자동 배포 커밋), 23커밋에 AI 협업(Co-Authored-By)

## 3. 잘 된 것 (자산)

- **문서 규율** — 3층 지식 구조(CLAUDE.md / claude-context / ADR) + 가드레일 + 온보딩. 매 챕터 문서 동반 커밋.
- **GitOps 투명성** — 선언=실제, 봇 배포 커밋으로 "무엇이 언제" 추적. 드리프트 0.
- **정직한 이력** — 실패(Revert)·충돌(merge)·임시조치를 숨기지 않고 기록.

## 4. 힘들었던 것 · 배운 것

| 사건 | 교훈 |
|------|------|
| Spot VM 전원 TERMINATED로 클러스터 다운 | Spot은 상시 선점. 학습엔 싸지만 가용성과 맞바꿈 |
| kubectl 삭제가 selfHeal로 되살아남(2회) | **영구 변경 = Git 변경**, kubectl 직접 삭제가 아님 |
| Canary 파드 Pending (app-pool CPU 예약 98%) | 작은 클러스터엔 surge 여유 없음. CPU는 예약 과다·실사용 미미 |
| 워커가 Kafka 준비 전 시작해 멈춤 | 앱이 의존성 기동 순서를 방어해야 함(waitForKafka) |
| gh 계정·workflow 스코프 마찰 | 자동화의 마지막 1%는 인증 경계에서 막힘 |

## 5. 쌓인 부채 ledger

| # | 부채 | 언제부터 | 위험 | 우선순위 |
|---|------|---------|------|---------|
| 1 | 로그 수집 중단 (Loki/FluentBit 제거) | ch6.2 | 관측 3종 중 로그 공백 | 🔴 |
| 2 | Kafka·Tempo emptyDir (비영속) | ch8 | 재시작·선점 시 큐/트레이스 소실 | 🔴 |
| 3 | 단일 인스턴스 + Zonal + Spot | 전반 | 가용성 없음 | 🟠 |
| 4 | NetworkPolicy 미강제 (엔포서 없음) | ch7.4 | 테넌트 네트워크 격리 무효 | 🟠 |
| 5 | 관측 requests 축소·replicas 1, default-pool 메모리 압박 | ch6.2 | 성능·복원력 저하 | 🟠 |
| 6 | prod Valkey·kube-prometheus가 Helm(비-GitOps) | ch6 | 단일 진실원천 이탈 | 🟡 |
| 7 | 워커 이미지 CI 자동갱신 불가(수동 동기화) | ch8 | 배포 시 수동 단계 | 🟡 |
| 8 | 테스트 부재(go test 없음) | 전반 | 회귀 안전망 없음 | 🟡 |
| 9 | 인프라(노드풀/클러스터)가 Git 밖 | 전반 | GitOps·selfHeal 미적용 영역 | 🟡 |

> 관통 원인: **자원이 빠듯한 단일존 Spot 클러스터.** CPU/메모리 확보용 임시조치가 계속 쌓였다.
> 자원 관찰(ch9): CPU는 예약 과다·실사용 13~31%(부족 아님), 메모리만 default-pool 한 노드에서 관측 스택 때문에 압박.

## 6. GitAIOps 결론 (ch9.4)

**Git이 AI와 Ops를 잇는 신뢰 경계이자 조율 계층이다.**

```
사람 의도 → AI(탐색→비교→실행) → Git 커밋(선언+ADR) → ArgoCD reconcile(selfHeal)
  → 클러스터(Ops) → 관측(메트릭·트레이스·헬스체크) → 피드백(AI 검증/사람 검토)
```

- AI는 클러스터를 직접 만지지 않고 **Git에 쓴다** → 검토·되돌리기·감사 가능한 채로 운영이 된다.
- **selfHeal이 안전 레일** — Git과 어긋난 직접 변경을 되돌린다(실증됨). 규율이 시스템으로 강제된다.
- 빈틈: **인프라가 Git 밖**(gcloud 직접) — IaC(Terraform)로 편입해야 루프가 완결된다. AI 직접-변경 경로는 가드레일/권한 정책으로 좁힌다.

## 7. 종합

학습·레퍼런스로는 완성형 — 넓은 주제를 일관된 서사와 근거(ADR)로 엮었고 운영 성숙 신호(가드레일·온보딩·트러블슈팅)까지 갖췄다. 프로덕션 전환의 관문은 부채 1~3(로그 복원·상태 영속화·최소 가용성)이며, 셋 다 노드 자원 확보가 선행이다.
