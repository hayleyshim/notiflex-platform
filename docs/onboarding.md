# Notiflex 온보딩

신규 합류자(사람·AI)가 이 저장소와 운영을 빠르게 파악하기 위한 출발점.

## 1. 먼저 읽을 것 (순서대로)

| 순서 | 문서 | 무엇을 얻나 |
|------|------|------------|
| 1 | [CLAUDE.md](../CLAUDE.md) | 프로젝트 개요·기술스택·GCP 설정·행동 규칙 |
| 2 | [claude-context/architecture.md](../claude-context/architecture.md) | 지금 어떻게 동작하나 (현재 스냅샷) |
| 3 | [docs/architecture-decisions.md](architecture-decisions.md) | 왜 이렇게 골랐나 (ADR-001~016) |
| 4 | [JOURNEY.md](../JOURNEY.md) | 진행 현황·도구 선택·버전·리소스·트러블슈팅 |
| 5 | [docs/retrospective.md](retrospective.md) | 여정 회고·쌓인 부채 ledger |

## 2. 저장소 지도

```
app/              Go 알림 API (stdlib, scratch). main.go + tracing.go
k8s/smb/          앱 스택: Rollout(Canary)·Service·Gateway·Kafka·worker·CronJob·CSI Secret
k8s/monitoring/   관측: 대시보드·알림·Loki/Tempo 데이터소스·Tempo
argocd/           App of Apps: root → smb·monitoring·tenants
tenants/          멀티테넌시: base(Kustomize) + customers/<고객>/ 오버레이
helm-values/      Helm 릴리스 값(kube-prometheus, loki, fluent-bit)
command-guardrails/  위험 작업 실행 절차(런북)
.github/workflows/   CI: build → push → 매니페스트 태그 갱신 → 커밋
```

## 3. 핵심 운영 사실

- **컨텍스트**: `kubectl --context gke-sysnet4admin_book_gitaiops ...` (모든 명령에 명시)
- **GCP**: 프로젝트 `hayley-gitaiops-project`, 리전 `asia-northeast3`(서울, zonal)
- **GitOps**: ArgoCD(selfHeal=true). **영구 변경은 Git에서** — kubectl 직접 변경은 되돌려진다.
- **배포 흐름**: `app/` 수정 → push → CI가 이미지 빌드·태그 갱신 커밋 → ArgoCD Canary 배포.
  - ⚠️ 워커 이미지(`k8s/smb/worker.yaml`)는 CI 자동갱신 대상이 아니라 **수동 동기화** 필요.
- **노드 배치**: app-pool(앱)·data-pool(Valkey·Kafka)·default-pool(플랫폼/관측). 전용 풀은 taint 격리.

## 4. 자주 하는 작업

| 하고 싶은 것 | 방법 |
|-------------|------|
| 앱 코드 배포 | `app/` 수정 → push (CI·ArgoCD 자동) → 워커 이미지 수동 동기화 |
| 새 고객 추가 | [tenants/README.md](../tenants/README.md) — `tenants/customers/<고객>/` 추가 |
| 새 ArgoCD 앱 추가 | `argocd/apps/`에 Application 파일 하나 추가(sync-wave 지정) |
| 배포/헬스 확인 | `kubectl -n argocd get applications`, `kubectl -n notiflex get all` |
| 요청 지연 진단 | Grafana → Explore → Tempo (트레이스로 구간별 지연) |
| 위험 작업 | 반드시 [command-guardrails/](../command-guardrails/) 절차대로 |

## 5. 주의 (현재 한계)

- 로그 수집(Loki) 미복원, Kafka·Tempo 비영속(emptyDir), NetworkPolicy 미강제, Spot 선점 가능.
- 상세·최신 부채는 [retrospective.md](retrospective.md) §5 참조.
