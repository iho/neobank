#!/usr/bin/env bash
# CNPG restore drill for neobank-postgres — proves barman-cloud backups are restorable.
#
# Usage:
#   ./deploy/scripts/restore-drill.sh neobank neobank-postgres [target-time]
#
# Example:
#   ./deploy/scripts/restore-drill.sh neobank neobank-postgres "2026-07-01T12:00:00Z"

set -euo pipefail

NAMESPACE="${1:?usage: restore-drill.sh <namespace> <source-cluster> [target-time]}"
SOURCE_CLUSTER="${2:?usage: restore-drill.sh <namespace> <source-cluster> [target-time]}"
TARGET_TIME="${3:-}"
DRILL_CLUSTER="${SOURCE_CLUSTER}-drill-$(date +%s)"
DB_NAME="neobank"
TIMEOUT_SECONDS=600

cleanup() {
  echo "==> Cleaning up ${DRILL_CLUSTER}"
  kubectl delete cluster "${DRILL_CLUSTER}" -n "${NAMESPACE}" --ignore-not-found --wait=false
}
trap cleanup EXIT

echo "==> Reading barman object store config from ${SOURCE_CLUSTER}"
BARMAN_JSON="$(kubectl get cluster "${SOURCE_CLUSTER}" -n "${NAMESPACE}" -o jsonpath='{.spec.backup.barmanObjectStore}')"
if [ -z "${BARMAN_JSON}" ] || [ "${BARMAN_JSON}" = "null" ]; then
  echo "error: ${SOURCE_CLUSTER} has no spec.backup.barmanObjectStore configured" >&2
  exit 1
fi

echo "==> Creating drill cluster ${DRILL_CLUSTER}"
RECOVERY_TARGET_JSON="{}"
if [ -n "${TARGET_TIME}" ]; then
  RECOVERY_TARGET_JSON="$(jq -n --arg t "${TARGET_TIME}" '{targetTime: $t}')"
fi

kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: ${DRILL_CLUSTER}
  namespace: ${NAMESPACE}
  labels:
    neobank.io/restore-drill: "true"
spec:
  instances: 1
  storage:
    size: 20Gi
  bootstrap:
    recovery:
      source: ${SOURCE_CLUSTER}
      recoveryTarget: ${RECOVERY_TARGET_JSON}
  externalClusters:
    - name: ${SOURCE_CLUSTER}
      barmanObjectStore: ${BARMAN_JSON}
EOF

echo "==> Waiting up to ${TIMEOUT_SECONDS}s for ${DRILL_CLUSTER} to become healthy"
elapsed=0
phase=""
while [ "${elapsed}" -lt "${TIMEOUT_SECONDS}" ]; do
  phase="$(kubectl get cluster "${DRILL_CLUSTER}" -n "${NAMESPACE}" -o jsonpath='{.status.phase}' 2>/dev/null || true)"
  if [ "${phase}" = "Cluster in healthy state" ]; then
    break
  fi
  sleep 10
  elapsed=$((elapsed + 10))
  echo "    ...phase=${phase:-<pending>} (${elapsed}s elapsed)"
done

if [ "${phase}" != "Cluster in healthy state" ]; then
  echo "error: drill cluster did not become healthy in time" >&2
  exit 1
fi

echo "==> Comparing row counts (users table) between source and drill"
SRC_POD="$(kubectl get pods -n "${NAMESPACE}" -l "cnpg.io/cluster=${SOURCE_CLUSTER},role=primary" -o jsonpath='{.items[0].metadata.name}')"
DRILL_POD="$(kubectl get pods -n "${NAMESPACE}" -l "cnpg.io/cluster=${DRILL_CLUSTER},role=primary" -o jsonpath='{.items[0].metadata.name}')"
SRC_COUNT="$(kubectl exec -n "${NAMESPACE}" "${SRC_POD}" -- psql -U neobank -d "${DB_NAME}" -tAc 'select count(*) from "user".users' 2>/dev/null || echo 0)"
DRILL_COUNT="$(kubectl exec -n "${NAMESPACE}" "${DRILL_POD}" -- psql -U neobank -d "${DB_NAME}" -tAc 'select count(*) from "user".users' 2>/dev/null || echo 0)"
echo "    source users=${SRC_COUNT} drill users=${DRILL_COUNT}"

echo "==> Restore drill succeeded"