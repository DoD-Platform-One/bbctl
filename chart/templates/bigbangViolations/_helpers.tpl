{{/*
Name of the app
*/}}
{{- define "bigbang-violations.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-violations" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-violations.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-violations" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-violations.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-violations.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-violations.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-violations.base-override-selector-labels" .) -}}
{{- if .Values.bigbangViolations.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangViolations.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-violations.base-override-selector-labels" -}}
app: {{ include "bigbang-violations.name" . }}
app.kubernetes.io/name: {{ include "bigbang-violations.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-violations.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-violations.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-violations.service-account-name" -}}
{{- if .Values.bigbangViolations.serviceAccount.create }}
{{- default (include "bigbang-violations.fullname" .) .Values.bigbangViolations.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangViolations.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-violations.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangViolationsAnnotations := .Values.bigbangViolations.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangViolationsAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-violations.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangViolationsAnnotations := .Values.bigbangViolations.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangViolationsAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-violations.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbViolationsConfig := .Values.bigbangViolations.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbViolationsConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-violations.command" -}}
- "/bin/bash"
- "-c"
- | 
  echo "$(./bbctl violations)"
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-violations.args" -}}
{{- end }}
