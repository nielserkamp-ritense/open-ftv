{{/*
Expand the name of the chart.
*/}}
{{- define "open-ftv-pdp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "open-ftv-pdp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "open-ftv-pdp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "open-ftv-pdp.labels" -}}
helm.sh/chart: {{ include "open-ftv-pdp.chart" . }}
{{ include "open-ftv-pdp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: pdp
app.kubernetes.io/part-of: open-ftv
{{- end }}

{{/*
Selector labels
*/}}
{{- define "open-ftv-pdp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "open-ftv-pdp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "open-ftv-pdp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "open-ftv-pdp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Environment variable items when using a Postgresql database for authorization decision log
*/}}
{{- define "open-ftv-pdp.authorizationDecisionLogPostgresql" -}}
{{- $postgresql := . -}}
{{- with .existingSecret -}}
- name: X_PDP_ADL_PG_HOSTNAME
  {{- if $postgresql.hostname }}
  value: {{ $postgresql.hostname }}
  {{- else }}
  valueFrom:
    secretKeyRef:
      name: {{ .name }}
      key: {{ .hostnameKey }}
  {{- end }}
- name: X_PDP_ADL_PG_DATABASE
  {{- if $postgresql.database }}
  value: {{ $postgresql.database }}
  {{- else }}
  valueFrom:
    secretKeyRef:
      name: {{ .name }}
      key: {{ .databaseKey }}
  {{- end }}
- name: X_PDP_ADL_PG_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ .name }}
      key: {{ .usernameKey }}
- name: X_PDP_ADL_PG_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .name }}
      key: {{ .passwordKey }}
- name: PDP_ADL_PG_URL
  value: {{ printf "postgresql://$(X_PDP_ADL_PG_USERNAME):$(X_PDP_ADL_PG_PASSWORD)@$(X_PDP_ADL_PG_HOSTNAME):%d/$(X_PDP_ADL_PG_DATABASE)" 5432 }}
{{- end }}
{{- end }}
