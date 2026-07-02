{{- define "platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "platform.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "platform.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "platform.labels" -}}
helm.sh/chart: {{ include "platform.chart" . }}
{{ include "platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: neobank
{{- end -}}

{{- define "platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "platform.namespace" -}}
{{- .Values.namespace.name | default "neobank" -}}
{{- end -}}

{{- define "platform.cnpg.clusterName" -}}
{{- .Values.cnpg.clusterName | default (printf "%s-postgres" (include "platform.fullname" .)) -}}
{{- end -}}

{{- define "platform.cnpg.rwService" -}}
{{- printf "%s-rw" (include "platform.cnpg.clusterName" .) -}}
{{- end -}}

{{- define "platform.redis.serviceName" -}}
{{- printf "%s-redis" (include "platform.fullname" .) -}}
{{- end -}}

{{- define "platform.goledger.image" -}}
{{- $registry := .Values.global.imageRegistry | default "ghcr.io/iho/neobank" -}}
{{- $tag := default .Values.global.imageTag .Values.goledger.image.tag -}}
{{- if .Values.goledger.image.repository -}}
{{- printf "%s:%s" .Values.goledger.image.repository $tag -}}
{{- else -}}
{{- printf "%s/goledger:%s" $registry $tag -}}
{{- end -}}
{{- end -}}

{{- define "platform.goledger.postgresCluster" -}}
{{- .Values.goledger.postgres.clusterName | default "goledger-postgres" -}}
{{- end -}}

{{- define "platform.goledger.postgresRW" -}}
{{- printf "%s-rw" (include "platform.goledger.postgresCluster" .) -}}
{{- end -}}