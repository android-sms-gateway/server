{{- if and .Values.config.secret.enabled .Values.config.unsecure.enabled -}}
{{- fail "Only one of config.secret.enabled or config.unsecure.enabled may be true." -}}
{{- end -}}