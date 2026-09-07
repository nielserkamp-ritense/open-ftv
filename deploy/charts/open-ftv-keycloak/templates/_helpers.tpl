{{/*
Expand the name of the chart.
*/}}
{{- define "open-ftv-keycloak.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "open-ftv-keycloak.fullname" -}}
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
{{- define "open-ftv-keycloak.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "open-ftv-keycloak.labels" -}}
helm.sh/chart: {{ include "open-ftv-keycloak.chart" . }}
{{ include "open-ftv-keycloak.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: keycloak
app.kubernetes.io/part-of: open-ftv
{{- end }}

{{/*
Selector labels
*/}}
{{- define "open-ftv-keycloak.selectorLabels" -}}
app.kubernetes.io/name: {{ include "open-ftv-keycloak.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Public hostname Keycloak stamps into issuer URLs. Defaults to the first routed
hostname so it cannot drift from httpRoute.hostnames; set config.hostname
explicitly for setups without an HTTPRoute.
*/}}
{{- define "open-ftv-keycloak.hostname" -}}
{{- if .Values.config.hostname }}
{{- .Values.config.hostname }}
{{- else if .Values.httpRoute.enabled }}
{{- printf "https://%s" (first .Values.httpRoute.hostnames) }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "open-ftv-keycloak.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "open-ftv-keycloak.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
