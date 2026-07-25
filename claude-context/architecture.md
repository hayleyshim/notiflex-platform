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
| 노드풀 (역할별, ch7.2) | `app-pool` role=app · `data-pool` role=data · `default-pool` role=platform (모두 e2-medium Spot, disk 30GB) |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| 활성화 기능 | Gateway API(standard), Workload Identity, Secret Manager CSI addon |
| 외부 IP | `35.216.9.148` (Gateway) |

### 역할별 노드풀 (ch7.2 멀티 노드풀)

| 노드풀 | role 라벨 | taint | 노드 수 | 워크로드 |
|--------|-----------|-------|---------|----------|
| `app-pool` | `role=app` | `dedicated=app:NoSchedule` | 1 | notiflex-api |
| `data-pool` | `role=data` | `dedicated=data:NoSchedule` | 1 | valkey-primary (stateful 격리) |
| `default-pool` | `role=platform` | 없음 | 2 | ArgoCD, kube-prometheus-stack, Argo Rollouts |

- 배치 방식: 전용 풀은 **taint로 격리**, 대상 워크로드는 `nodeSelector`+`toleration`으로 유입. platform은 taint 없는 default-pool에 자연 수렴.
- notiflex-api 배치는 `k8s/smb/rollout.yaml`(GitOps)에 선언. Valkey는 리포 밖 Helm 릴리스라 StatefulSet에 직접 패치(durable화하려면 Helm values에 반영 필요).
- GKE 시스템 애드온(kube-dns, konnectivity 등)은 모든 taint를 tolerate → 전역 배치(역할 격리와 무관).

> ✅ ch6.2 임시 3노드 증설은 ch7.2에서 해소: app/data를 전용 풀로 분리하고 default-pool은 platform 전용 2노드로 복원.
> ⚠️ Spot 선점 이력: 5일 방치 중 default-pool Spot VM 전원 TERMINATED된 적 있음 → `gcloud compute instances start`로 복구. Spot은 상시 선점 가능.
> ⚠️ ch6.2에서 CPU 확보 위해 Loki·FluentBit 임시 제거됨(로그 수집 일시 중단). 복원 예정.

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

### App of Apps (ch7.3)

여러 ArgoCD Application을 루트 하나로 일괄 관리한다.

```
argocd/root.yaml  (Application: notiflex-root, path=argocd/apps recurse)
   └─ 스캔 → 자식 Application 자동 생성·동기화 (sync-wave 순서대로)
        ├─ [wave 0] argocd/apps/notiflex-monitoring.yaml → k8s/monitoring (ns monitoring)
        └─ [wave 1] argocd/apps/notiflex-smb.yaml        → k8s/smb (ns notiflex)
```

- 설치 순서: `argocd.argoproj.io/sync-wave`로 제어. **관측(wave 0) → 앱(wave 1)** — 앱이 뜰 때 알림·대시보드가 이미 준비됨. 낮은 wave가 Healthy 된 뒤 다음 wave 진행.
- 새 앱 추가 = `argocd/apps/`에 Application 파일 하나 추가(원하는 wave 지정) → 루트가 자동 편입.
- 범위: git 매니페스트만(smb·monitoring). Helm 릴리스(valkey, kube-prometheus-stack)는 아직 GitOps 밖.
- k8s/monitoring은 이전까지 `kubectl apply` 수동 관리 → 이번에 GitOps로 편입.

### 멀티테넌시 (ch7.4, 네임스페이스/테넌트 PoC)

고객사(테넌트)마다 격리된 환경을 **한 클러스터 안 네임스페이스 단위**로 제공한다.

```
ApplicationSet(tenants, git 디렉터리 generator: tenants/customers/*)
   └─ 고객 디렉터리마다 Application(tenant-<이름>) 자동 생성
        → tenants/customers/<이름> (Kustomize 오버레이) 를 ns tenant-<이름> 에 배포
             └─ ../../base 참조:
                  ├─ notiflex-api (Deployment) + Service
                  ├─ valkey (테넌트 전용 무인증, Deployment) + Service
                  ├─ ResourceQuota + LimitRange   (자원 상한)
                  └─ NetworkPolicy                (네트워크 격리)
             + 오버레이 패치: 요금제별 레플리카·쿼터·라벨(tenant/tier)
AppProject(tenants)  ← 배포 대상을 tenant-* ns + 이 저장소로 제한(경계)
```

- **고객 추가 = `tenants/customers/<이름>/kustomization.yaml` 디렉터리 하나 추가** → 자동 편입. (절차: `tenants/README.md`)
- 요금제별 오버레이: acme=enterprise(레플리카 2, 쿼터 500m/20p), globex=standard(레플리카 1, 250m/10p).
- 검증: acme `/id` 1→2→3, globex 독립적으로 1부터 → **테넌트별 데이터(Valkey) 분리 확인**.
- 효과 있는 격리: 네임스페이스, ResourceQuota/LimitRange, AppProject 경계.
- ⚠️ **NetworkPolicy 미강제**: 클러스터에 엔포서(Dataplane V2/Calico) 미설치. 매니페스트는 존재하나 실제 차단은 `--enable-network-policy` 후에야 유효.
- 배치: 테넌트 파드는 taint 없는 default-pool(platform)에 스케줄. 노드 레벨 격리가 필요하면 테넌트 전용 노드풀로 확장 가능.
- 프로덕션 `k8s/smb`(단일 notiflex)와 별개 경로(`tenants/`)라 서로 영향 없음.

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
