{{- define "neobank.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "neobank.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "neobank.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "neobank.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "neobank.labels" -}}
helm.sh/chart: {{ include "neobank.chart" . }}
{{ include "neobank.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "neobank.selectorLabels" -}}
app.kubernetes.io/name: {{ include "neobank.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "neobank.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "neobank.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "neobank.secretName" -}}
{{- .Values.secrets.existingSecret | default (printf "%s-app" (include "neobank.fullname" .)) -}}
{{- end -}}

{{- define "neobank.image" -}}
{{- $values := .Values -}}
{{- $registry := $values.global.imageRegistry -}}
{{- $tag := default $values.global.imageTag .tag -}}
{{- if .repository -}}
{{- printf "%s:%s" .repository $tag -}}
{{- else -}}
{{- printf "%s/%s:%s" $registry .image $tag -}}
{{- end -}}
{{- end -}}

{{- define "neobank.podSecurityContext" -}}
{{- if .Values.podSecurity.enabled }}
runAsNonRoot: true
runAsUser: 65532
runAsGroup: 65532
fsGroup: 65532
seccompProfile:
  type: RuntimeDefault
{{- end }}
{{- end -}}

{{- define "neobank.containerSecurityContext" -}}
{{- if .Values.podSecurity.enabled }}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
    - ALL
{{- end }}
{{- end -}}

{{- define "neobank.serviceHost" -}}
{{- printf "%s-%s" (include "neobank.fullname" .root) .name -}}
{{- end -}}

{{- define "neobank.env.workload" -}}
{{- $root := .root -}}
{{- $name := .name -}}
{{- $wl := .workload -}}
- name: APP_ENV
  value: {{ $root.Values.config.appEnv | quote }}
- name: HTTP_PORT
  value: {{ $wl.httpPort | quote }}
{{- if $wl.grpcPort }}
- name: GRPC_PORT
  value: {{ $wl.grpcPort | quote }}
{{- end }}
{{- if ne $name "gateway" }}
{{- if $root.Values.secrets.create }}
- name: DATABASE_URL
  value: {{ $root.Values.config.databaseURL | quote }}
{{- else }}
- name: DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: database-url
{{- end }}
{{- end }}
{{- if or (eq $name "gateway") (eq $name "user") (eq $name "payment") (eq $name "card") }}
{{- if $root.Values.secrets.create }}
- name: REDIS_URL
  value: {{ $root.Values.config.redisURL | quote }}
{{- else }}
- name: REDIS_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: redis-url
      optional: true
{{- end }}
{{- end }}
{{- if or (eq $name "gateway") (eq $name "user") }}
{{- if $root.Values.secrets.create }}
- name: JWT_SECRET
  value: {{ required "config.jwtSecret required" $root.Values.config.jwtSecret | quote }}
{{- else }}
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: jwt-secret
{{- end }}
{{- end }}
{{- if $root.Values.config.kafkaBrokers }}
- name: KAFKA_BROKERS
  value: {{ $root.Values.config.kafkaBrokers | quote }}
{{- else if not $root.Values.secrets.create }}
- name: KAFKA_BROKERS
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: kafka-brokers
      optional: true
{{- end }}
{{- if $root.Values.config.ledgerGrpcAddr }}
- name: LEDGER_GRPC_ADDR
  value: {{ $root.Values.config.ledgerGrpcAddr | quote }}
{{- else if not $root.Values.secrets.create }}
- name: LEDGER_GRPC_ADDR
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: ledger-grpc-addr
      optional: true
{{- end }}
{{- if $root.Values.config.otelEndpoint }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ $root.Values.config.otelEndpoint | quote }}
{{- else if not $root.Values.secrets.create }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: otel-endpoint
      optional: true
{{- end }}
{{- if eq $name "user" }}
{{- if $root.Values.config.vaultAddr }}
- name: VAULT_ADDR
  value: {{ $root.Values.config.vaultAddr | quote }}
{{- else if not $root.Values.secrets.create }}
- name: VAULT_ADDR
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: vault-addr
      optional: true
{{- end }}
{{- if not $root.Values.secrets.create }}
- name: VAULT_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: vault-token
      optional: true
{{- end }}
- name: NOTIFICATION_SERVICE_URL
  value: {{ printf "http://%s:8083/api/v1/internal/events" (include "neobank.serviceHost" (dict "root" $root "name" "notification")) }}
{{- end }}
{{- if eq $name "gateway" }}
- name: USER_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "user") }}:50052
- name: PAYMENT_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "payment") }}:50053
- name: CARD_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "card") }}:50054
- name: NOTIFICATION_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "notification") }}:50055
{{- end }}
{{- if or (eq $name "payment") (eq $name "card") }}
- name: USER_SERVICE_URL
  value: {{ printf "http://%s:8081" (include "neobank.serviceHost" (dict "root" $root "name" "user")) }}
- name: USER_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "user") }}:50052
- name: NOTIFICATION_SERVICE_URL
  value: {{ printf "http://%s:8083/api/v1/internal/events" (include "neobank.serviceHost" (dict "root" $root "name" "notification")) }}
{{- end }}
{{- if eq $name "notification" }}
- name: USER_GRPC_ADDR
  value: {{ include "neobank.serviceHost" (dict "root" $root "name" "user") }}:50052
{{- end }}
{{- if eq $name "card" }}
{{- if $root.Values.config.settlementLedgerAccountID }}
- name: SETTLEMENT_LEDGER_ACCOUNT_ID
  value: {{ $root.Values.config.settlementLedgerAccountID | quote }}
{{- else if not $root.Values.secrets.create }}
- name: SETTLEMENT_LEDGER_ACCOUNT_ID
  valueFrom:
    secretKeyRef:
      name: {{ include "neobank.secretName" $root }}
      key: settlement-ledger-account-id
      optional: true
{{- end }}
{{- end }}
{{- end -}}