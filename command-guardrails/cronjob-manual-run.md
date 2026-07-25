# 크론잡 수동 실행 🟡

**위험도: 낮음** (헬스체크 기준). 단, **데이터를 변경하는 크론잡**(정리·백업 등)을
스케줄 밖에서 수동 실행하면 부작용이 클 수 있으므로 🟠로 취급한다.

> 수동 실행은 CronJob으로부터 일회성 Job을 만든다. 이 Job은 CronJob의
> `successfulJobsHistoryLimit`/`failedJobsHistoryLimit` 관리 대상이 아니므로 **직접 정리**해야 한다.

## 0. 컨텍스트

```bash
CTX=gke-sysnet4admin_book_gitaiops
```

## 1. 사전 확인 (무엇을 실행하는지 반드시 파악)

```bash
# 크론잡 목록·스케줄·suspend 여부
kubectl --context $CTX -n notiflex get cronjob

# 실제 실행 커맨드 확인 — 수동 실행 전 "이 Job이 무슨 일을 하는지" 반드시 읽는다
kubectl --context $CTX -n notiflex get cronjob <CRONJOB> -o yaml | \
  grep -A20 "containers:"
```

- [ ] 대상 크론잡이 맞는가
- [ ] 이 Job이 **읽기 전용 점검**인가, **데이터를 변경**하는가 (후자면 승인 필요)
- [ ] `concurrencyPolicy: Forbid`이면 스케줄 실행과 겹치지 않게 타이밍 확인

## 2. 실행

```bash
# 일회성 Job 생성 (이름은 충돌 방지 위해 유니크하게)
kubectl --context $CTX -n notiflex create job <CRONJOB>-manual --from=cronjob/<CRONJOB>
```

## 3. 검증

```bash
kubectl --context $CTX -n notiflex wait --for=condition=complete job/<CRONJOB>-manual --timeout=60s
kubectl --context $CTX -n notiflex logs job/<CRONJOB>-manual
kubectl --context $CTX -n notiflex get job <CRONJOB>-manual   # COMPLETIONS 확인
```

## 4. 정리 (필수)

```bash
# 수동 Job은 히스토리 관리 대상이 아니므로 직접 삭제
kubectl --context $CTX -n notiflex delete job <CRONJOB>-manual
```

## 예시: notiflex-healthcheck

```bash
CTX=gke-sysnet4admin_book_gitaiops
kubectl --context $CTX -n notiflex create job healthcheck-manual --from=cronjob/notiflex-healthcheck
kubectl --context $CTX -n notiflex wait --for=condition=complete job/healthcheck-manual --timeout=60s
kubectl --context $CTX -n notiflex logs job/healthcheck-manual        # "[healthcheck] OK" 확인
kubectl --context $CTX -n notiflex delete job healthcheck-manual
```

## 되돌리기 / 주의

- 읽기 전용 점검(헬스체크)은 되돌릴 것이 없다.
- 데이터 변경 크론잡은 **수동 실행 전 반드시 커맨드를 읽고**, 가능하면 dry-run 옵션이나
  스테이징에서 먼저 검증한다. 수동 실행은 CronJob의 스케줄/동시성 보호를 우회할 수 있다.
