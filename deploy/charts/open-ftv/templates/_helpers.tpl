{{/*
Expand the name of the chart.
*/}}
{{- define "open-ftv.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "open-ftv.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "open-ftv.labels" -}}
helm.sh/chart: {{ include "open-ftv.chart" . }}
{{ include "open-ftv.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: open-ftv
{{- end }}

{{/*
Selector labels
*/}}
{{- define "open-ftv.selectorLabels" -}}
app.kubernetes.io/name: {{ include "open-ftv.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
