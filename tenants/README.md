# 고객(테넌트) 온보딩

Notiflex는 **네임스페이스/테넌트** 방식으로 고객사마다 격리된 환경을 제공한다.
고객 온보딩은 **Git에 파일 하나 추가 → ArgoCD가 자동 반영**하는 선언적 절차다.

## 구조

```
tenants/
├── base/                     # 모든 고객 공통 매니페스트 (Kustomize base)
│   ├── notiflex.yaml         #   notiflex API + Service
│   ├── valkey.yaml           #   테넌트 전용 Valkey (무인증)
│   ├── quota.yaml            #   ResourceQuota + LimitRange (기본/standard 티어)
│   └── networkpolicy.yaml    #   네트워크 격리 정책
└── customers/                # 고객별 오버레이 (여기에 고객이 쌓인다)
    ├── acme/kustomization.yaml     # enterprise 티어
    └── globex/kustomization.yaml   # standard 티어
```

- `argocd/apps/tenants-appset.yaml` 의 **ApplicationSet(git 디렉터리 generator)** 이
  `tenants/customers/*` 를 스캔해 고객마다 Application `tenant-<이름>` 을 만들고
  네임스페이스 `tenant-<이름>` 에 배포한다.
- `argocd/apps/tenants-appproject.yaml` 의 **AppProject `tenants`** 가
  배포 대상을 `tenant-*` 네임스페이스 + 이 저장소로 제한한다(경계).

## 새 고객 추가하기

1. 오버레이 디렉터리를 만든다 (standard 티어는 base 그대로):

   ```
   tenants/customers/<고객명>/kustomization.yaml
   ```

   ```yaml
   apiVersion: kustomize.config.k8s.io/v1beta1
   kind: Kustomization
   resources:
     - ../../base
   labels:
     - pairs:
         tenant: <고객명>
         tier: standard
       includeSelectors: false
   ```

2. (선택) 요금제별 조정이 필요하면 patch 추가 — 예: enterprise 티어

   ```yaml
   patches:
     - patch: |-
         apiVersion: apps/v1
         kind: Deployment
         metadata: { name: notiflex-api }
         spec: { replicas: 2 }
     - patch: |-
         apiVersion: v1
         kind: ResourceQuota
         metadata: { name: tenant-quota }
         spec:
           hard:
             requests.cpu: "500m"
             requests.memory: 512Mi
             limits.cpu: "2"
             limits.memory: 2Gi
             pods: "20"
   ```

3. 커밋 & 푸시 → ArgoCD가 자동으로 `tenant-<고객명>` 네임스페이스와 워크로드를 생성한다.

   ```bash
   git add tenants/customers/<고객명>
   git commit -m "feat: onboard <고객명>"
   git push origin main
   ```

## 고객 제거(오프보딩)

`tenants/customers/<고객명>/` 디렉터리를 삭제하고 푸시하면, ApplicationSet이
해당 Application을 제거하고, ApplicationSet 템플릿의 `resources-finalizer`가
**워크로드·ResourceQuota·NetworkPolicy를 cascade 삭제(prune)** 한다.

```bash
git rm -r tenants/customers/<고객명>
git commit -m "chore: offboard <고객명>"
git push origin main
```

⚠️ 네임스페이스 `tenant-<고객명>` 은 `CreateNamespace=true` 부수효과로 생성돼
git이 관리하지 않으므로 **빈 상태로 남는다**. 완전히 지우려면 수동 삭제가 필요하다:

```bash
kubectl --context <ctx> delete namespace tenant-<고객명>
```

(단, `.claude/settings.local.json` 의 deny 규칙이 `kubectl delete namespace` 를
차단하므로, 정책상 사람이 직접 실행해야 한다.)

## 참고 / 한계

- 격리: 네임스페이스 · ResourceQuota/LimitRange · AppProject 경계는 **강제됨**.
- ⚠️ **NetworkPolicy는 현재 미강제** — 클러스터에 엔포서(GKE Dataplane V2/Calico) 미설치.
  실제 네트워크 차단은 `gcloud container clusters update notiflex-cluster --enable-network-policy` 후 유효.
- 테넌트 파드는 taint 없는 `default-pool` 에 스케줄된다(soft 격리). 노드 레벨 격리가
  필요하면 고객 전용 노드풀(taint) + 오버레이의 nodeSelector/toleration 으로 확장한다.
- 이 PoC는 무인증 Valkey를 쓴다(프로덕션 `k8s/smb` 의 CSI 시크릿 방식과 다름).
