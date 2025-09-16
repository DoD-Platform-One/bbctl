{{/*
Name of the app
*/}}
{{- define "bigbang-preflight.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-preflight" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-preflight.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-preflight" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-preflight.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-preflight.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-preflight.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-preflight.base-override-selector-labels" .) -}}
{{- if .Values.bigbangPreflight.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangPreflight.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-preflight.base-override-selector-labels" -}}
app: {{ include "bigbang-preflight.name" . }}
app.kubernetes.io/name: {{ include "bigbang-preflight.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-preflight.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-preflight.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-preflight.service-account-name" -}}
{{- if .Values.bigbangPreflight.serviceAccount.create }}
{{- default (include "bigbang-preflight.fullname" .) .Values.bigbangPreflight.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangPreflight.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-preflight.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangPreflightAnnotations := .Values.bigbangPreflight.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangPreflightAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-preflight.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangPreflightAnnotations := .Values.bigbangPreflight.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangPreflightAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-preflight.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbPreflightConfig := .Values.bigbangPreflight.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbPreflightConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-preflight.command" -}}
{{- $args := "" }}
{{- $imageOverride := dig "baseConfig" "preflight-check" "image" "" .Values.AsMap }}
{{- if $imageOverride }}
{{- $args = print $args "-i " $imageOverride }}
{{- end }}

- "/bin/bash"
- "-c"
- | 
  echo {{ print "$(./bbctl preflight-check " (include "bbctl.registryOverride" . | trim) " " $args ")"}}
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-preflight.args" -}}
{{- end }}
