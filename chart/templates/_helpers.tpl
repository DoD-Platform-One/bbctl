{{/*
Expand the name of the chart.
*/}}
{{- define "bbctl.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bbctl.fullname" -}}
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
{{- define "bbctl.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bbctl.labels" -}}
{{- $labels := fromYaml (include "bbctl.baseLabels" .) }}
{{- if .Values.baseLabels }}
{{ $labels := mustMergeOverwrite .Values.baseLabels $labels }}
{{- end }}
{{- toYaml $labels }}
{{- end }}


{{/*
bbctl base labels
*/}}
{{- define "bbctl.baseLabels" -}}
helm.sh/chart: {{ include "bbctl.chart" . }}
{{ include "bbctl.selector-labels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/component: "bigbang-automation"
app.kubernetes.io/part-of: "bigbang"
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bbctl.selector-labels" -}}
app: {{ include "bbctl.name" . }}
app.kubernetes.io/name: {{ include "bbctl.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bbctl.service-account-name" -}}
{{- if .Values.bigbangUpdater.serviceAccount.create }}
{{- default (include "bbctl.fullname" .) .Values.bigbangUpdater.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangUpdater.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bbctl.service-account-annotations" -}}
{{- if .Values.serviceAccount.annotations }}
{{ .Values.serviceAccount.annotations | toYaml }}
{{- else }}
{}
{{- end }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bbctl.pod-annotations" -}}
{{- if or .Values.podAnnotations .Values.bigbang.istio.enabled }}
{{- if .Values.bigbang.istio.enabled }}
sidecar.istio.io/inject: "true"
{{- end }}
{{- if .Values.podAnnotations }}
{{ .Values.podAnnotations | toYaml }}
{{- end }}
{{- else }}
{}
{{- end }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bbctl.config" -}}
{{- if .Values.baseConfig }}
{{ .Values.baseConfig | toYaml }}
{{- else }}
{}
{{- end }}
{{- end }}
