# Command Guardrails — 위험 작업 실행 절차

되돌리기 어렵거나 데이터를 파괴하는 작업의 **안전 실행 절차(런북)**를 모아둔다.
AI와 사람 모두 이 절차를 따라 실행한다. 즉흥 실행 금지.

## 공통 원칙 (CLAUDE.md 행동 규칙)

1. **항상 확인 후 실행** — 무엇을 왜 바꾸는지 먼저 설명하고 동의를 받는다.
2. **변경 전 현재 상태 확인** — 삭제/수정 전에 대상의 현재 상태를 반드시 먼저 조회한다.
3. **kubectl 컨텍스트 명시** — 모든 명령에 `--context gke-sysnet4admin_book_gitaiops`를 붙인다.
4. **GitOps를 먼저 생각한다** — 이 저장소는 ArgoCD(selfHeal=true)가 관리한다.
   클러스터에서 직접 지운 리소스는 Git 선언과 어긋나면 **자동 복구(selfHeal)**된다.
   영구 변경은 대개 **Git에서 바꿔야** 한다 (kubectl 직접 변경이 아니라).

> 편의를 위해 절차 예시는 컨텍스트를 변수로 둔다: `CTX=gke-sysnet4admin_book_gitaiops`

## 위험도 표기

| 등급 | 의미 |
|------|------|
| 🔴 파괴적·비가역 | 데이터/리소스가 영구 소실. 복구 불가. |
| 🟠 주의 | 되돌릴 수 있으나 오작동 시 영향 큼. |
| 🟡 낮음 | 대체로 안전하나 절차 준수 필요. |

## 절차 목록

| 작업 | 위험도 | 문서 |
|------|--------|------|
| 카프카 토픽 삭제 | 🔴 파괴적·비가역 | [kafka-topic-delete.md](kafka-topic-delete.md) |
| 크론잡 수동 실행 | 🟡 낮음 (대상에 따라 🟠) | [cronjob-manual-run.md](cronjob-manual-run.md) |
| 테넌트 네임스페이스 삭제 | 🔴 파괴적·비가역 | [tenant-namespace-delete.md](tenant-namespace-delete.md) |
