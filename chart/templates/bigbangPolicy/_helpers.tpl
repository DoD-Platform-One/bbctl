{{/*
Name of the app
*/}}
{{- define "bigbang-policy.name" -}}
{{- $baseName := include "bbctl.name" . }}
{{- printf "%s-%s" $baseName "bigbang-policy" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bigbang-policy.fullname" -}}
{{- $baseName := include "bbctl.fullname" . }}
{{- printf "%s-%s" $baseName "bigbang-policy" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bigbang-policy.labels" -}}
{{- $appLabels := fromYaml (include "bbctl.labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-policy.selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $appLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Override Selector labels
*/}}
{{- define "bigbang-policy.override-selector-labels" -}}
{{- $labels := fromYaml (include "bigbang-policy.base-override-selector-labels" .) -}}
{{- if .Values.bigbangPolicy.selectorLabels }}
{{- $labels := mustMergeOverwrite $labels .Values.bigbangPolicy.selectorLabels -}}
{{- end }}
{{ toYaml $labels }}
{{- end }}

{{/*
Base Override Selector labels
*/}}
{{- define "bigbang-policy.base-override-selector-labels" -}}
app: {{ include "bigbang-policy.name" . }}
app.kubernetes.io/name: {{ include "bigbang-policy.name" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bigbang-policy.selector-labels" -}}
{{- $bbctlSelectorLabels := fromYaml (include "bbctl.selector-labels" .) -}}
{{- $selectorLabels := fromYaml (include "bigbang-policy.override-selector-labels" .) -}}
{{- $labels := mustMergeOverwrite $bbctlSelectorLabels $selectorLabels -}}
{{ toYaml $labels }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bigbang-policy.service-account-name" -}}
{{- if .Values.bigbangPolicy.serviceAccount.create }}
{{- default (include "bigbang-policy.fullname" .) .Values.bigbangPolicy.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.bigbangPolicy.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the service account annotations
*/}}
{{- define "bigbang-policy.service-account-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.service-account-annotations" .) -}}
{{- $bigbangPolicyAnnotations := .Values.bigbangPolicy.serviceAccount.annotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangPolicyAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the pod annotations
*/}}
{{- define "bigbang-policy.pod-annotations" -}}
{{- $bbctlAnnotations := fromYaml (include "bbctl.pod-annotations" .) -}}
{{- $bigbangPolicyAnnotations := .Values.bigbangPolicy.podAnnotations -}}
{{- $annotations := mustMergeOverwrite $bbctlAnnotations $bigbangPolicyAnnotations -}}
{{ toYaml $annotations }}
{{- end }}

{{/*
Create the config file
*/}}
{{- define "bigbang-policy.config" -}}
{{- $bbctlConfig := fromYaml (include "bbctl.config" .) -}}
{{- $bbPolicyConfig := .Values.bigbangPolicy.config -}}
{{- $config := mustMergeOverwrite $bbctlConfig $bbPolicyConfig -}}
{{ toYaml $config }}
{{- end }}

{{/*
Create the command
*/}}
{{- define "bigbang-policy.command" -}}
- "/bin/bash"
- "-c"
- | 
  echo "$(./bbctl policy --{{ .Values.bigbangPolicy.policyEnforcer }})"
{{- end }}

{{/*
Create the args
*/}}
{{- define "bigbang-policy.args" -}}
{{- end }}
