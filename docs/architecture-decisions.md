# Architecture Decision Records

이 문서는 Notiflex 플랫폼의 아키텍처 결정을 시간 순서로 누적한다.
각 결정의 "왜"를 기록해, 나중에 팀원이 합류하거나 회고할 때 맥락을 잃지 않도록 한다.

## ADR-001: GitOps 도구로 ArgoCD 채택 (3장)
**시점**: 2026-07 / **결정**: 배포 자동화 도구로 ArgoCD를 채택하고 Flux·Jenkins X·Spinnaker는 쓰지 않는다.
**이유**:
- Web UI로 배포 상태를 실시간 시각화 → 학습·운영 중 "지금 무슨 일이 일어나는지" 눈으로 확인
- Application CRD로 "어떤 Git 경로 → 어떤 네임스페이스"를 선언적으로 관리
- selfHeal로 누군가 kubectl로 직접 수정해도 Git 상태로 자동 복원 (Git = 단일 진실 소스)
- CNCF Graduated로 커뮤니티가 가장 활발하고, e2-medium 노드에서 ~500MB로 구동 가능

## ADR-002: CI 도구로 GitHub Actions 채택 (3장)
**시점**: 2026-07 / **결정**: CI로 GitHub Actions를 채택하고 Cloud Build 단독·GitLab CI·Jenkins는 쓰지 않는다.
**이유**:
- 코드가 GitHub에 있어 CI가 같은 플랫폼 → 별도 서버 설치·운영 불필요
- `.github/workflows/*.yaml` 한 파일로 파이프라인을 선언적으로 정의
- private 저장소도 월 2,000분 무료라 학습 환경 비용 부담 없음
- `google-github-actions/auth`로 GCP 인증이 간편하고, 빌드는 필요 시 Cloud Build에 위임 가능

## ADR-003: 메트릭 모니터링으로 Prometheus + Grafana 채택 (4장)
**시점**: 2026-07 / **결정**: 메트릭 스택으로 kube-prometheus-stack(Prometheus+Grafana)을 채택하고 Datadog·CloudWatch·Cloud Monitoring은 쓰지 않는다.
**이유**:
- Kubernetes 모니터링의 사실상 표준(CNCF Graduated)이며 SaaS 구독료 없이 자체 호스팅
- Helm 번들로 Prometheus·Grafana·Alertmanager·exporter를 검증된 조합으로 일괄 설치
- 이후 Loki(로그)·Tempo(트레이스)를 Grafana 하나로 통합 → 도구 파편화 없음
- e2-medium에서 requests 축소로 구동 가능

## ADR-004: 로그 수집으로 Loki + Fluent Bit 채택 (4장)
**시점**: 2026-07 / **결정**: 로그 스택으로 Loki + Fluent Bit를 채택하고 ELK·CloudWatch·Cloud Logging은 쓰지 않는다.
**이유**:
- 경량(Loki ~128Mi, Fluent Bit ~64Mi) — ELK(Elasticsearch 2Gi+)는 e2-medium에서 불가능
- 4장에서 설치한 Grafana에 데이터소스만 추가하면 메트릭과 같은 화면에서 로그 조회
- 라벨 기반 인덱싱으로 풀텍스트 대비 저장 비용이 낮음
- Fluent Bit DaemonSet이 모든 노드의 컨테이너 로그를 자동 수집

## ADR-005: 알림으로 PrometheusRule + Alertmanager 채택 (4장)
**시점**: 2026-07 / **결정**: 알림으로 PrometheusRule + Alertmanager를 채택하고 Grafana Alerting·PagerDuty·Cloud Monitoring은 쓰지 않는다.
**이유**:
- PrometheusRule을 YAML로 관리 → Git에 누적되고 ArgoCD가 동기화(GitOps 일관성 유지)
- Alertmanager는 4장 kube-prometheus-stack 설치 시 이미 포함 → 추가 설치·비용 0
- 규칙이 PR로 리뷰되고 `git blame`으로 "왜 이 임계값?"을 추적 가능
- Alertmanager의 그루핑·억제·라우팅으로 심각도별 다단계 알림 표현

## ADR-006: 외부 트래픽 관리로 Gateway API 채택 (5장)
**시점**: 2026-07 / **결정**: 외부 노출로 GKE Gateway API(`gke-l7-regional-external-managed`)를 채택하고 Ingress NGINX·Istio·Traefik은 쓰지 않는다.
**이유**:
- Kubernetes 공식 차세대 표준(1.27 GA)으로 Ingress의 한계를 개선
- GKE 네이티브 지원 → 별도 Ingress Controller 설치 없이 로드밸런서 자동 생성
- Gateway(인프라팀)/HTTPRoute(앱팀)로 관심사 분리
- HTTPRoute의 backendRefs가 5장 Blue/Green 트래픽 분배의 기반이 됨

## ADR-007: 무중단 배포로 Blue/Green(Argo Rollouts) 채택 (5장)
**시점**: 2026-07 / **결정**: 무중단 배포 전략으로 Argo Rollouts 기반 Blue/Green을 채택한다. (6장에서 Canary로 진화 예정)
**이유**:
- 신버전을 preview 서비스로 격리 기동해 active(사용자) 트래픽에 영향 없이 사전 검증
- cutover가 한순간에 일어나 롤링 업데이트처럼 두 버전이 섞이지 않음
- 문제 발견 시 승격을 막아 사용자 노출을 원천 차단 가능
- ArgoCD와 같은 클러스터에서 동작하며 Rollout CRD로 GitOps 흐름 유지

## ADR-008: 캐시/상태 공유로 Valkey 채택 (6장)
**시점**: 2026-07 / **결정**: Pod 간 공유 카운터로 Valkey(standalone)를 채택하고 Redis·Memcached·DragonflyDB는 쓰지 않는다.
**이유**:
- Redis 100% 호환이면서 BSD 라이선스라 상용 환경에서도 라이선스 걱정 없음 (Redis는 SSPL)
- `INCR` 원자적 증가로 여러 Pod이 동시 호출해도 ID 중복 없이 순차 발급
- 영속성이 있어 Pod 재시작 후에도 카운터 유지 (Memcached는 영속성·INCR 부재로 부적합)
- Bitnami Helm 차트로 간편 설치, standalone 모드로 경량 구동

## ADR-009: 시크릿 관리로 CSI Driver + GCP Secret Manager 채택 (6장)
**시점**: 2026-07 / **결정**: 시크릿을 Secrets Store CSI Driver + GCP Secret Manager + Workload Identity로 관리하고 Sealed Secrets·External Secrets Operator·평문 K8s Secret은 쓰지 않는다.
**이유**:
- GKE 네이티브 통합 — Workload Identity가 GKE KSA와 GCP IAM을 직접 연결
- SA 키 파일 불필요 (OIDC 토큰) → 키 유출 위험 제거
- Secret Manager가 시크릿의 단일 진실 소스, CSI로 Pod에 파일 마운트
- K8s Secret은 base64 인코딩일 뿐 암호화가 아니고 Git 커밋도 불가

## ADR-010: 배포 전략을 Blue/Green에서 Canary로 진화 (6장)
**시점**: 2026-07 / **결정**: 무중단 배포를 Canary(20/50/80/100 점진)로 전환한다. 도구는 Argo Rollouts를 유지하고 strategy만 변경한다.
**이유**:
- 전체가 아닌 20% 사용자만 먼저 노출해 위험을 최소화 (Blue/Green은 0→100% 즉시)
- 각 단계에서 pause로 관찰, 문제 시 abort로 stable 즉시 복원
- 리소스 효율 — Canary 1.2x vs Blue/Green 2x (CPU 제약 환경에 유리)
- 새 도구 설치 없이 같은 Rollout CRD에서 strategy만 전환 ("점진적 고도화")
