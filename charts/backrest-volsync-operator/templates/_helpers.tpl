{{- define "backrest-volsync-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "backrest-volsync-operator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "backrest-volsync-operator.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "backrest-volsync-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "backrest-volsync-operator.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- required "serviceAccount.name is required when serviceAccount.create=false" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "backrest-volsync-operator.metricsBindAddress" -}}
{{- default (printf ":%d" (int .Values.metrics.port)) .Values.metrics.bindAddress -}}
{{- end -}}

{{- define "backrest-volsync-operator.healthProbeBindAddress" -}}
{{- default (printf ":%d" (int .Values.healthProbe.port)) .Values.healthProbe.bindAddress -}}
{{- end -}}
