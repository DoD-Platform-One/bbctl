{{/*
Name of the app
*/}}
{{- define "bigbang-status.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-status" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-status.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-status" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-status.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-status.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-status.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-status.base-override-selector-labels" .) -}}
{{- if .Values.bigbangStatus.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangStatus.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-status.base-override-selector-labels" -}}
app: {{ include "bigbang-status.name" . }}
app.kubernetes.io/name: {{ include "bigbang-status.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-status.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-status.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-status.service-account-name" -}}
{{- if .Values.bigbangStatus.serviceAccount.create }}
{{- default (include "bigbang-status.fullname" .) .Values.bigbangStatus.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangStatus.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-status.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangStatusAnnotations := .Values.bigbangStatus.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangStatusAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-status.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangStatusAnnotations := .Values.bigbangStatus.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangStatusAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-status.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbStatusConfig := .Values.bigbangStatus.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbStatusConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-status.command" -}}
- "./bbctl"
- "status"
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-status.args" -}}
{{- end }}
