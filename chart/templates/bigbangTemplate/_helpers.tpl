{{/*
Name of the app
*/}}
{{- define "bigbang-template.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-template" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-template.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-template" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-template.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-template.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-template.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-template.base-override-selector-labels" .) -}}
{{- if .Values.bigbangTemplate.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangTemplate.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-template.base-override-selector-labels" -}}
app: {{ include "bigbang-template.name" . }}
app.kubernetes.io/name: {{ include "bigbang-template.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-template.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-template.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-template.service-account-name" -}}
{{- if .Values.bigbangTemplate.serviceAccount.create }}
{{- default (include "bigbang-template.fullname" .) .Values.bigbangTemplate.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangTemplate.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-template.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangTemplateAnnotations := .Values.bigbangTemplate.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangTemplateAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-template.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangTemplateAnnotations := .Values.bigbangTemplate.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangTemplateAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-template.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbTemplateConfig := .Values.bigbangTemplate.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbTemplateConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-template.command" -}}
- "/bin/bash"
- "-c"
- | 
  echo "$(./bbctl template)"
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-template.args" -}}
{{- end }}
