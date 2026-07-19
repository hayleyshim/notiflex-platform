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
| ch5 | 5.2 트래픽 관리 | ⬜ | | |
| ch5 | 5.3 무중단 배포 | ⬜ | | |
| ch6 | 6.1 캐시 | ⬜ | | |
| ch6 | 6.2 시크릿 관리 | ⬜ | | |
| ch6 | 6.3 Canary 전환 | ⬜ | | |
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

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | 초기 (OTel/valkey 대비 처음부터 1.25) |
| Notiflex 이미지 | api:sha-cf00e05 (v0.1.2) | v0.1.0 → v0.1.1 → CI SHA 태그 전환 (ch3.4) |
| ArgoCD | v3.4.5 | ch3.2 설치 (stable manifest) |
| kube-prometheus-stack | 87.17.0 | ch4.2 설치 (Prometheus+Grafana+Alertmanager) |
| Loki | 3.6.8 (chart 7.1.0) | ch4.3 설치 (SingleBinary, PV 2Gi) |
| Fluent Bit | 2.1.0 (chart 2.6.0) | ch4.3 설치 (DaemonSet) |
| Kafka | - | 미설치 (ch8) |
| OTel SDK | - | 미설치 (ch8) |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot, disk 30GB) | 2 | notiflex-api(2), ArgoCD, 관측 스택(Prometheus/Grafana/Loki/Fluent Bit/Alertmanager) |

> ⚠️ ch4 관측 스택 설치 후 노드 CPU 할당이 84~90%로 상승. ch6 진입(CSI DaemonSet 240m) 전에 Prometheus/Grafana/Alertmanager requests를 5m으로 축소 필요.

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| 2.5 | 클러스터 생성 직후 GatewayClass 인스턴스가 비어 있음 | CRD는 정상 설치됨. GKE 컨트롤러 리컨실까지 약 2분 대기하니 자동 등장 (재활성화 불필요) |
| 2.6 | `gcloud builds submit`이 두 번 PERMISSION_DENIED | Cloud Build API를 빌드 직전 활성화해 권한 전파 미완료가 원인. 서비스 에이전트 프로비저닝 후 재시도하니 성공 |
| 4.3 | Loki가 `mkdir /var/loki: read-only file system`으로 CrashLoop | persistence 비활성 시 쓰기 볼륨이 없어 발생. SingleBinary에 PV 2Gi 부여. StatefulSet 볼륨템플릿은 in-place 수정 불가라 StatefulSet 삭제 후 helm upgrade로 재생성 |
| 4.4 | 알림 테스트 시 Pod 삭제로는 restartCount가 안 오름 | Pod 삭제는 새 Pod(restart=0) 생성이라 무효. liveness probe를 잘못된 포트로 패치해 컨테이너 크래시 루프 유발(같은 Pod 내 재시작). 테스트 중엔 ArgoCD auto-sync를 잠시 해제, 후 복원 |
