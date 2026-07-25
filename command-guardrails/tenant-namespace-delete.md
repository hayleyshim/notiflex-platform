# 테넌트 네임스페이스 삭제 🔴

**위험도: 파괴적·비가역** — 네임스페이스 삭제는 그 안의 **모든 리소스를 cascade로 영구 삭제**한다.
고객(테넌트) 데이터가 사라지므로 오프보딩 확정 후에만 실행한다.

> 핵심: 테넌트는 **ApplicationSet(git 디렉터리 generator)** 이 관리한다.
> `tenants/customers/<이름>/` 이 Git에 남아 있으면, 네임스페이스를 kubectl로 지워도
> ApplicationSet이 앱을 재생성하고 `CreateNamespace=true`로 **네임스페이스·워크로드가 되살아난다**(selfHeal).
> → **반드시 Git 오프보딩을 먼저** 하고, 그 뒤 남은 빈 네임스페이스만 삭제한다.

## 0. 컨텍스트

```bash
CTX=gke-sysnet4admin_book_gitaiops
```

## 1. 먼저: Git 오프보딩 (정상 경로)

```bash
# 고객 오버레이 디렉터리 제거 → ApplicationSet이 tenant-<이름> 앱을 prune
#   (템플릿의 resources-finalizer가 워크로드·Quota·NetworkPolicy를 cascade 삭제)
git rm -r tenants/customers/<이름>
git commit -m "chore: offboard <이름>"
git push origin main
```

자세한 배경: [../tenants/README.md](../tenants/README.md) 의 "고객 제거(오프보딩)".

## 2. 사전 확인 (변경 전 현재 상태)

네임스페이스를 지우기 전에 **오프보딩이 실제로 반영됐는지** 확인한다.

```bash
# ApplicationSet이 더 이상 이 테넌트 앱을 만들지 않는지 (앱이 없어야 함)
kubectl --context $CTX -n argocd get application tenant-<이름>
# → "NotFound" 여야 안전

# 네임스페이스가 비었는지 (워크로드가 prune됐는지)
kubectl --context $CTX -n tenant-<이름> get all
# → "No resources found" 여야 안전
```

- [ ] `tenants/customers/<이름>/` 가 Git에서 제거·푸시됐는가
- [ ] `tenant-<이름>` ArgoCD Application이 사라졌는가 (NotFound)
- [ ] 네임스페이스가 비었는가 (되살아날 워크로드 없음)
- [ ] 고객 데이터 삭제를 승인받았는가

## 3. 실행

```bash
kubectl --context $CTX delete namespace tenant-<이름>
```

> ⚠️ 이 명령은 위험 명령 차단 정책(`.claude/settings.local.json` 의 deny 규칙,
> 설정돼 있는 경우)에 걸릴 수 있다. 그때는 정책상 **사람이 직접** 실행해야 한다.

## 4. 검증

```bash
kubectl --context $CTX get ns tenant-<이름>
# → "NotFound" 여야 완료
```

## 5. 되돌리기 / 주의

- 네임스페이스 삭제는 **비가역**이다. 복구는 오프보딩을 되돌리는 것(=`tenants/customers/<이름>/`
  디렉터리를 다시 추가·푸시)뿐이며, 이는 **빈 환경을 새로 만드는 것**이지 데이터 복원이 아니다.
- 순서를 뒤집지 말 것: Git에 남은 활성 테넌트의 네임스페이스를 먼저 지우면 selfHeal이 되살려
  "안 지워지는" 혼란만 생긴다(2번째 턴에서 겪은 GitOps 원리와 동일).
