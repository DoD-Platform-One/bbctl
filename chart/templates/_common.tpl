{{/*
Create the cluster role binding
*/}}
{{- define "bbctl.common.cluster-admin-cluster-role-binding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: "{{ .custom.serviceAccountName }}-cluster-role-binding"
  labels:
    {{- .custom.labels | nindent 4 }}
  annotations:
    {{- .custom.annotations | nindent 4 }}
subjects:
- kind: ServiceAccount
  name: {{ .custom.serviceAccountName }}
  namespace: {{ .scope.Release.Namespace }}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
{{- end }}

{{/*
Create the configmap
*/}}
{{- define "bbctl.common.config-maps" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name:  {{ .custom.fullname }}-config
  labels:
    {{- .custom.labels | nindent 4 }}
data:
  config.yaml: |
    # The configuration for the {{ .custom.fullname }}, this comment provides a properly indented first line
    {{- .custom.config | nindent 4 }}
  credentials.yaml: |
    # The credentials file for the {{ .custom.fullname }}, this comment provides a properly indented first line
    {{- .scope.Values.credentialsFile | toYaml | nindent 4 }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name:  {{ .custom.fullname }}-kube-config
  labels:
    {{- .custom.labels | nindent 4 }}
data:
  config: |
    # The kube-config for the bigbang-updater, this comment provides a properly indented first line
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        server: https://kubernetes.default.svc
        certificate-authority-data: REPLACED_WITH_INIT_CONTAINER
      name: default
    contexts:
    - context:
        cluster: default
        user: default
      name: default
    current-context: default
    users:
    - name: default
      user:
        token: REPLACED_WITH_INIT_CONTAINER
---
{{- end }}

{{/*
Create the cron job
*/}}
{{- define "bbctl.common.cron-job" -}}
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ .custom.fullname }}
  labels:
    {{- .custom.labels | nindent 4 }}
spec:
  schedule: {{ .custom.schedule }}
  jobTemplate:
    spec:
      template:
        metadata:
          annotations:
            {{- .custom.annotations | nindent 12 }}
          labels:
            {{- .custom.labels | nindent 12 }}
        spec:
          {{- with .scope.Values.imagePullSecrets }}
          imagePullSecrets:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          serviceAccountName: {{ .custom.serviceAccountName }}
          securityContext:
            {{- toYaml .scope.Values.podSecurityContext | nindent 12 }}
          initContainers:
          - name: {{ .custom.fullname }}-init
            securityContext:
              {{- toYaml .scope.Values.securityContext | nindent 14 }}
            image: "{{ .scope.Values.yqImage.repository }}:{{ .scope.Values.yqImage.tag }}"
            imagePullPolicy: {{ .scope.Values.yqImage.pullPolicy }}
            command:
            - "/bin/sh"
            args:
            - "-c"
            - |
              echo "starting yq to update kubeconfig" \
              && export SERVICE_ACCOUNT_FOLDER=/var/run/secrets/kubernetes.io/serviceaccount \
              && export TOKEN=$(cat $SERVICE_ACCOUNT_FOLDER/token) \
              && export CACERT_FILE=$SERVICE_ACCOUNT_FOLDER/ca.crt \
              && export CACERT_ENCODED=$(cat $CACERT_FILE | base64) \
              && export KUBECONFIG=/home/bigbang/.kube/config
              if [ -f "$KUBECONFIG" ]; then
                echo "kubeconfig already exists, not updating"
              else
                cp /home/bigbang/ro.kube/config $KUBECONFIG \
                && chmod 700 $KUBECONFIG \
                && chown 1000:1000 $KUBECONFIG \
                && yq '.clusters[0].cluster.certificate-authority-data = "'"$CACERT_ENCODED"'"' -i $KUBECONFIG \
                && yq '.users[0].user.token = "'"$(echo $TOKEN)"'"' -i $KUBECONFIG \
                && echo "yq done updating kubeconfig"
              fi
              # curl -s -X GET --cacert "${CACERT}" --header "Authorization: Bearer ${TOKEN}" ${HOST}/api/v1/configmaps | yq -r '[.items[] | select(.metadata.name=="kube-root-ca.crt") | .data."ca.crt"] | unique | .[]'
            volumeMounts:
            - name: config-volume
              mountPath: /home/bigbang/.bbctl/
              readOnly: true
            - name: ro-kube-config-volume
              mountPath: /home/bigbang/ro.kube/
              readOnly: true
            - name: kube-config-volume
              mountPath: /home/bigbang/.kube/
              readOnly: false
            - name: tmp
              mountPath: /tmp
              readOnly: false
          containers:
          - name: {{ .scope.Chart.Name }}
            securityContext:
              {{- toYaml .scope.Values.securityContext | nindent 14 }}
            image: "{{ .scope.Values.image.repository }}:{{ .scope.Values.image.tag | default .scope.Chart.AppVersion }}"
            imagePullPolicy: {{ .scope.Values.image.pullPolicy }}
            command: 
            {{- .custom.command | nindent 12 }}
            {{- with .custom.args }}
            args:
            {{- . | nindent 12 }}
            {{- end }}
            volumeMounts:
            - name: config-volume
              mountPath: /home/bigbang/.bbctl/
              readOnly: true
            - name: kube-config-volume
              mountPath: /home/bigbang/.kube/
              readOnly: true
            {{- with .scope.Values.resources }}
            resources:
              {{- toYaml . | nindent 14 }}
            {{- end }}
          volumes:
          - name: config-volume
            configMap:
              name: {{ .custom.fullname }}-config
          - name: ro-kube-config-volume
            configMap:
              name: {{ .custom.fullname }}-kube-config
          - name: kube-config-volume
            emptyDir: {}
          - name: tmp
            emptyDir: {}
          restartPolicy: Never
          {{- with .scope.Values.nodeSelector }}
          nodeSelector:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with .scope.Values.affinity }}
          affinity:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with .scope.Values.tolerations }}
          tolerations:
            {{- toYaml . | nindent 12 }}
          {{- end }}
{{- end }}

{{/*
Create the service account
*/}}
{{- define "bbctl.common.service-account" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .custom.serviceAccountName }}
  labels:
    {{- .custom.labels | nindent 4 }}
  annotations:
    {{- .custom.annotations | nindent 4 }}
{{- end }}
