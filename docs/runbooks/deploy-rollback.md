# Deploy and rollback

## Compose (staging VM)

```bash
export IMAGE_TAG=sha-<short>
make up-ghcr-ledger    # includes goledger from GHCR
curl -fsS http://localhost:8080/health
```

Rollback: pin `IMAGE_TAG` to the previous `sha-*` tag and re-run `make up-ghcr-ledger`.

## Kubernetes (Helm)

Platform deps first, then app:

```bash
helm upgrade --install neobank-platform deploy/helm/platform \
  -f deploy/helm/platform/values-production.yaml -n neobank --create-namespace

helm upgrade --install neobank deploy/helm/neobank \
  -f deploy/helm/neobank/values-production.yaml \
  -n neobank --set global.imageTag=sha-<short>
```

Rollback:

```bash
helm rollback neobank <revision> -n neobank
helm rollback neobank-platform <revision> -n neobank
```

## ArgoCD

```bash
kubectl apply -f deploy/argocd/project.yaml
kubectl apply -f deploy/argocd/deps/          # see deps/README.md for order
kubectl apply -f deploy/argocd/application-platform.yaml
kubectl apply -f deploy/argocd/application-neobank.yaml
```

Rollback: `argocd app rollback neobank` or sync a previous git revision.

Migrations run automatically via the pre-upgrade Helm hook Job. If migrate fails, the release aborts — fix the DB issue before retrying.