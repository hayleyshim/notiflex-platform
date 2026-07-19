# Notiflex 여정 기록

이 파일은 독자가 실제로 진행한 내용을 기록한다. AI가 각 챕터 완료 시 자동으로 업데이트한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-07-19 | Claude Code 2.1.215, statusline 설정 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-07-19 | SDK 576.0.0, 공식 스크립트로 설치, 서울 리전 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-07-19 | hayleyshim/notiflex-platform (private) |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-07-19 | notiflex-cluster, e2-medium×2 Spot, Gateway API |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-07-19 | Cloud Build → api:v0.1.0, replicas 2 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-07-19 | Initial commit |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-07-19 | ArgoCD v3.4.5, private repo 연결, notiflex-smb App |
| ch3 | 3.3 기능 추가 | ✅ | 2026-07-19 | /version 엔드포인트, 롤링 업데이트 + git revert 롤백 |
| ch3 | 3.4 CI | ✅ | 2026-07-19 | GitHub Actions, SHA 태그 빌드/푸시 (방식 A) |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-07-19 | CI가 매니페스트 자동 갱신→ArgoCD 배포, E2E 검증 |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-07-19 | kube-prometheus-stack, Notiflex 대시보드 |
| ch4 | 4.3 로그 수집 | ✅ | 2026-07-19 | Loki(SingleBinary) + Fluent Bit, Grafana Loki 데이터소스 |
| ch4 | 4.4 알림 | ✅ | 2026-07-19 | PrometheusRule 2종, firing E2E 검증 |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-07-19 | Gateway API, 외부 IP 35.216.9.148, HealthCheckPolicy |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-07-19 | Argo Rollouts Blue/Green, v0.2.0→v0.3.0 승격 검증 |
| ch5 | 5.4 ADR 기록 | ✅ | 2026-07-19 | docs/architecture-decisions.md 신설 (ADR-001~007) |
| ch6 | 6.1 캐시 | ✅ | 2026-07-19 | Valkey standalone, /id를 INCR로 (Pod 간 공유) |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-07-19 | CSI + Secret Manager + Workload Identity, 파일 마운트 |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-07-19 | Blue/Green→Canary(20/50/80), v0.6.0 점진 배포 |
| ch6 | 6.4 아키텍처 스냅샷 | ✅ | 2026-07-19 | claude-context/architecture.md 신설 |
| ch7 | 7.2 멀티 노드풀 | ⬜ | | |
| ch7 | 7.3 App of Apps | ⬜ | | |
| ch7 | 7.4 멀티테넌시 | ⬜ | | |
| ch8 | 8.1 메시징 | ⬜ | | |
| ch8 | 8.2 트레이싱 | ⬜ | | |
| ch8 | 8.3 CronJob | ⬜ | | |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

## 도구 선택 기록

독자가 3-프롬프트 패턴(탐색→비교→실행)에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| gcloud 설치 | 공식 설치 스크립트 (홈 디렉터리) | Homebrew | Homebrew 미설치 + sudo 대화형 입력 불가로 sudo 불필요한 공식 스크립트로 전환 |
| GitOps 도구 (ch3) | ArgoCD | Flux, Jenkins X, Spinnaker | Web UI로 배포 상태 시각화, Application CRD 선언적 관리, selfHeal, e2-medium에서 구동 가능 |
| CI 도구 (ch3) | GitHub Actions | Cloud Build, GitLab CI, Jenkins | GitHub 네이티브(별도 서버 불필요), YAML 선언적, private 월 2000분 무료, GCP 인증 간편 |
| 메트릭 모니터링 (ch4) | Prometheus + Grafana | Datadog, CloudWatch, Cloud Monitoring | 오픈소스 표준(무료), Helm 번들 일괄 설치, Grafana로 로그/트레이스까지 통합 |
| 로그 수집 (ch4) | Loki + Fluent Bit | ELK, CloudWatch, Cloud Logging | 경량(128Mi, ELK 2Gi+ 불가), Grafana 네이티브 통합, 라벨 인덱싱으로 저비용 |
| 알림 (ch4) | PrometheusRule + Alertmanager | Grafana Alerting, PagerDuty, Cloud Monitoring | GitOps 호환(CRD를 Git 관리), 이미 설치됨(추가 비용 0), 실무 표준 라우팅 |
| 외부 트래픽 (ch5) | Gateway API | Ingress NGINX, Istio, Traefik | K8s 공식 표준, GKE 네이티브(Controller 불필요), Gateway/HTTPRoute 역할 분리, Blue/Green 연동 |
| 무중단 배포 (ch5) | Blue/Green (Argo Rollouts) | 롤링 업데이트, Flagger, Istio | preview로 사전 검증, 한순간 cutover, 문제 시 승격 차단, ArgoCD와 GitOps 유지 |
| 캐시/상태 공유 (ch6) | Valkey | Redis, Memcached, DragonflyDB | Redis 호환+BSD 라이선스, INCR 원자적 ID, 영속성(재시작 유지), Bitnami 차트 |
| 시크릿 관리 (ch6) | CSI Driver + Secret Manager | Sealed Secrets, External Secrets, kubectl secret | GKE 네이티브, Workload Identity로 SA키 불필요, 단일 진실 소스, 파일 마운트 |
| 배포 전략 진화 (ch6) | Canary (Argo Rollouts) | Blue/Green 유지, Rolling | 20/50/80 점진으로 위험 최소화, 리소스 1.2x(vs B/G 2x), 도구 변경 없이 strategy만 전환 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | 초기 (OTel/valkey 대비 처음부터 1.25) |
| Notiflex 이미지 | api:sha-865dad5 (v0.6.0) | …→v0.4.0(Valkey)→v0.5.0(CSI Secret)→v0.6.0(Canary) |
| ArgoCD | v3.4.5 | ch3.2 설치 (stable manifest) |
| Argo Rollouts | v1.9.1 | ch5.3 설치, ch6.3 Canary 전략으로 전환 |
| Valkey | bitnami/valkey standalone | ch6.1 설치 (공유 카운터) |
| kube-prometheus-stack | 87.17.0 | ch4.2 설치 (Prometheus+Grafana+Alertmanager) |
| Loki | 3.6.8 (chart 7.1.0) | ch4.3 설치 (SingleBinary, PV 2Gi) |
| Fluent Bit | 2.1.0 (chart 2.6.0) | ch4.3 설치 (DaemonSet) |
| Kafka | - | 미설치 (ch8) |
| OTel SDK | - | 미설치 (ch8) |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot, disk 30GB) | **3 (임시 증설)** | notiflex-api(Canary), Valkey, ArgoCD, 관측 스택, CSI Secret DaemonSet |

> ⚠️ ch6.2 CSI(240m) 수용 위해 default-pool 2→3노드 임시 증설 + Loki/FluentBit 임시 제거 + 관측 스택 requests 5m~2m 축소 + replicas 1. **ch7.2 노드풀 추가 후 복원** (memory: todo-ch7-restore-resources).

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| 2.5 | 클러스터 생성 직후 GatewayClass 인스턴스가 비어 있음 | CRD는 정상 설치됨. GKE 컨트롤러 리컨실까지 약 2분 대기하니 자동 등장 (재활성화 불필요) |
| 2.6 | `gcloud builds submit`이 두 번 PERMISSION_DENIED | Cloud Build API를 빌드 직전 활성화해 권한 전파 미완료가 원인. 서비스 에이전트 프로비저닝 후 재시도하니 성공 |
| 4.3 | Loki가 `mkdir /var/loki: read-only file system`으로 CrashLoop | persistence 비활성 시 쓰기 볼륨이 없어 발생. SingleBinary에 PV 2Gi 부여. StatefulSet 볼륨템플릿은 in-place 수정 불가라 StatefulSet 삭제 후 helm upgrade로 재생성 |
| 4.4 | 알림 테스트 시 Pod 삭제로는 restartCount가 안 오름 | Pod 삭제는 새 Pod(restart=0) 생성이라 무효. liveness probe를 잘못된 포트로 패치해 컨테이너 크래시 루프 유발(같은 Pod 내 재시작). 테스트 중엔 ArgoCD auto-sync를 잠시 해제, 후 복원 |
| 5.3 | app/ 수정 시마다 CI SHA 태그 vs 로컬 버전 태그 충돌 반복 | CI가 커밋을 sha 태그로 다시 태깅해 rollout.yaml을 덮어씀. `git pull --no-rebase`로 충돌을 로컬 버전 태그(v0.x.0)로 해소 후 병합 커밋. 코드는 항상 최신 버전으로 수렴. 근본 해결은 SHA 태그 통일 또는 ArgoCD Image Updater |
| 5.3 | 데모용 직접 배포 후 auto-sync 복원 시 ArgoCD가 옛 리비전 배포 | 폴링 지연으로 최신 커밋 전 상태를 동기화. `kubectl annotate app ... argocd.argoproj.io/refresh=hard`로 강제 새로고침해 수렴 |
| 6.2 | CSI(240m) 활성화 후 valkey/notiflex Pending, 2노드 CPU 96%+ | Loki/FluentBit 제거+관측 requests 축소로도 부족 → default-pool 3노드로 임시 증설해 해소. valkey STS 50m 유령 pod은 helm upgrade로 재생성 |
| 6.2 | 로컬 Go/Docker 없어 go.sum 생성 불가 | Dockerfile 빌드 단계에서 `go mod tidy` 실행해 컨테이너 안에서 의존성/go.sum 해결 (COPY go.sum 불필요) |
| 6.3 | Canary weight가 0%로 표시됨 | 트래픽 라우터(Gateway 플러그인) 미연동이라 replica 기반 canary. 스텝(Step 6/6)·pause는 정상 동작. 정밀 %분할은 trafficRouting 플러그인 추가 시 |
