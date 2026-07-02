#!/usr/bin/env bash
# Local k8s dev cluster with ingress-ready kind config.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_NAME="${CLUSTER_NAME:-neobank}"

if ! command -v kind >/dev/null 2>&1; then
  echo "Install kind: https://kind.sigs.k8s.io/" >&2
  exit 1
fi

if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
  kind create cluster --name "${CLUSTER_NAME}" --config "${ROOT}/deploy/kind/kind-config.yaml"
fi

echo "==> Cluster ${CLUSTER_NAME} ready. Apply platform deps:"
echo "    kubectl apply -f deploy/argocd/project.yaml"
echo "    kubectl apply -f deploy/argocd/deps/"
echo ""
echo "Or install charts directly:"
echo "    helm upgrade --install neobank-platform deploy/helm/platform -f deploy/helm/platform/values-staging.yaml -n neobank --create-namespace"
echo "    helm upgrade --install neobank deploy/helm/neobank -f deploy/helm/neobank/values-staging.yaml -n neobank --create-namespace"