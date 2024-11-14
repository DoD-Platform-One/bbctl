{{/*
Name of the app
*/}}
{{- define "bigbang-updater.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-updater" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-updater.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-updater" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-updater.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-updater.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-updater.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-updater.base-override-selector-labels" .) -}}
{{- if .Values.bigbangUpdater.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangUpdater.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-updater.base-override-selector-labels" -}}
app: {{ include "bigbang-updater.name" . }}
app.kubernetes.io/name: {{ include "bigbang-updater.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-updater.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-updater.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-updater.service-account-name" -}}
{{- if .Values.bigbangUpdater.serviceAccount.create }}
{{- default (include "bigbang-updater.fullname" .) .Values.bigbangUpdater.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangUpdater.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-updater.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangUpdaterAnnotations := .Values.bigbangUpdater.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangUpdaterAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-updater.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangUpdaterAnnotations := .Values.bigbangUpdater.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangUpdaterAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-updater.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbUpdaterConfig := .Values.bigbangUpdater.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbUpdaterConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-updater.command" -}}
- "/bin/bash"
- "-c"
- | 
  echo "$(./bbctl version -U -A)"
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-updater.args" -}}
{{- end }}