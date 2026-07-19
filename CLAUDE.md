# Notiflex Platform

## 프로젝트 개요

Notiflex — **B2B 알림 SaaS 플랫폼**. 고객사에 알림(notification)을 전송하는 서비스를 GKE 위에서 운영한다.
이 저장소는 애플리케이션 코드, 쿠버네티스 매니페스트, CI 파이프라인을 담는 GitOps 작업 저장소이다.

## 기술 스택

- **언어**: Go 표준 라이브러리만 사용 (외부 웹 프레임워크 없음)
- **컨테이너**: `scratch` 베이스 이미지 (최소 크기, 정적 바이너리)
- **인프라**: GKE Standard (Zonal), Spot VM
- **GitOps**: ArgoCD
- **관측 가능성**: Prometheus, Grafana, Loki, Fluent Bit, Tempo
- **배포 전략**: Rolling → Blue/Green → Canary (점진 진화)

## GCP 설정

| 항목 | 값 |
|------|-----|
| 프로젝트 ID | `hayley-gitaiops-project` |
| 리전 | `asia-northeast3` (서울) |
| 존 | `asia-northeast3-a` |

## Artifact Registry

컨테이너 이미지 저장소:

```
asia-northeast3-docker.pkg.dev/hayley-gitaiops-project/notiflex
```

## 행동 규칙

- **항상 확인 후 실행**: 변경 작업 전에 무엇을 왜 바꾸는지 먼저 설명한다.
- **변경 전 현재 상태 확인**: 리소스를 수정/삭제하기 전에 현재 상태를 먼저 조회한다.
- **kubectl 컨텍스트 명시**: 모든 kubectl 명령에 실습 클러스터 컨텍스트를 지정한다.

## 디렉터리 구조

```
notiflex-platform/
├── CLAUDE.md
├── app/           # Go 애플리케이션
├── k8s/
│   └── smb/       # K8s 매니페스트
└── .github/
    └── workflows/ # CI 파이프라인
```
