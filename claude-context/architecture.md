# Notiflex 아키텍처 스냅샷 (6장 완료 시점)

> 이 문서는 **현재 시점의 아키텍처 한눈 보기**다. AI가 매 대화에서 전체 그림을 빠르게 잡도록 돕는다.
> 세부 진행 기록은 `JOURNEY.md`, 결정 이유는 `docs/architecture-decisions.md`(ADR), 매니페스트는 `k8s/`를 참조한다.

## 3층 지식 구조

| 문서 | 역할 | 성격 |
|------|------|------|
| **CLAUDE.md** | 프로젝트 메타데이터 (기술스택·GCP설정·행동규칙) | 매 대화 자동 로드 |
| **claude-context/** | 현재 아키텍처 스냅샷 (지금 어떻게 동작하나) | AI 자동 참조 |
| **docs/architecture-decisions.md** | 결정 누적 (왜 이걸 골랐나, ADR-001~007) | 사람+AI 검토 |

세 층이 분리되어야 작업 컨텍스트(memory)·현재 그림(claude-context)·과거 결정(ADR)이 섞이지 않는다.

## 클러스터 토폴로지

| 항목 | 값 |
|------|-----|
| 클러스터 | `notiflex-cluster` (GKE Standard, Zonal) |
| 리전/존 | `asia-northeast3` / `asia-northeast3-a` (서울) |
| 노드풀 | `default-pool` — e2-medium **3노드**(Spot, disk 30GB) |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| 활성화 기능 | Gateway API(standard), Workload Identity, Secret Manager CSI addon |
| 외부 IP | `35.216.9.148` (Gateway) |

> ⚠️ default-pool은 ch6.2 CSI(240m) 수용을 위해 2→3노드로 임시 증설. ch7 노드풀 추가 후 복원 검토.
> ⚠️ ch6.2에서 CPU 확보 위해 Loki·FluentBit 임시 제거됨(로그 수집 일시 중단). ch7.2에서 복원 예정.

## 컴포넌트 다이어그램 (트래픽 흐름)

```
인터넷
  │ http://35.216.9.148/...
  ▼
GKE Regional External LB (gke-l7-regional-external-managed)
  │ proxy-only-subnet 172.16.0.0/23
  ▼
Gateway(notiflex-gateway) → HTTPRoute(notiflex-route, path /)
  │ HealthCheckPolicy: /health:8080
  ▼
Service notiflex-api (stable) ─┐
Service notiflex-api-preview (canary) ─┤  ← Argo Rollouts Canary가 관리
  ▼
Rollout notiflex-api (Canary 20→50→80→100)
  └─ Pod (notiflex API, scratch 이미지)
        ├─ INCR → Valkey (valkey-primary:6379, 공유 카운터)
        └─ 비밀번호 ← CSI 마운트 /mnt/secrets/valkey-password
                       ↑ GCP Secret Manager (Workload Identity, KSA notiflex-sa)
```

## 배포 파이프라인 (GitOps)

```
개발자: app/ 수정 → git push (main)
   ▼
GitHub Actions CI: docker build → push api:sha-<커밋>
   → k8s/smb/rollout.yaml 이미지 태그 갱신 → 봇 커밋 [skip ci] → push
   ▼
ArgoCD (auto-sync, selfHeal): k8s/smb 감시 → Canary 배포
   ▼
Argo Rollouts: 20%→50%→80%→100% 점진 전환 (각 30초 pause)
```

- 이미지 저장소: `asia-northeast3-docker.pkg.dev/hayley-gitaiops-project/notiflex/api`
- 현재 실행: `api:sha-865dad5` (v0.6.0)
- 배포 전략 진화: Rolling(3장) → Blue/Green(5장) → **Canary(6장, 현재)**

## 관측 가능성

| 도구 | 역할 | 상태 |
|------|------|------|
| Prometheus | 메트릭 수집(scrape) | 동작 (requests 5m로 축소) |
| Grafana | 메트릭·로그 통합 대시보드 | 동작 (Notiflex 대시보드 포함) |
| Alertmanager + PrometheusRule | 알림 (PodRestartTooMany, NotiflexApiDown) | 동작 |
| Loki + Fluent Bit | 로그 수집 | **ch6.2에서 임시 제거** (ch7 복원 예정) |
| Tempo | 분산 트레이싱 | 미설치 (ch8 예정) |

## 주요 네임스페이스

| 네임스페이스 | 주요 워크로드 |
|-------------|--------------|
| `notiflex` | Rollout notiflex-api(Canary), StatefulSet valkey-primary, Gateway/HTTPRoute, SecretProviderClass |
| `argocd` | ArgoCD v3.4.5 (7 워크로드: server, repo-server, application-controller 등) |
| `argo-rollouts` | Argo Rollouts 컨트롤러 v1.9.1 |
| `monitoring` | kube-prometheus-stack (Prometheus, Grafana, Alertmanager, operator, kube-state-metrics, node-exporter) |
| `kube-system` | CSI Secret Store DaemonSet (secrets-store-gke) |
