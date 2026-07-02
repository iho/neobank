# Deploy and rollback

## Compose (staging VM)

```bash
export IMAGE_TAG=sha-<short>
make up-ghcr
curl -fsS http://localhost:8080/health
```

Rollback: pin `IMAGE_TAG` to the previous `sha-*` tag and re-run `make up-ghcr`.

## Kubernetes (Helm)

```bash
helm upgrade --install neobank deploy/helm/neobank \
  -f deploy/helm/neobank/values-production.yaml \
  -n neobank --set global.imageTag=sha-<short>
```

Rollback:

```bash
helm rollback neobank <revision> -n neobank
```

Migrations run automatically via the pre-upgrade Helm hook Job. If migrate fails, the release aborts — fix the DB issue before retrying.