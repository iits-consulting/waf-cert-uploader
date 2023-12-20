Example Secret:
```yaml
apiVersion: v1
type: kubernetes.io/tls
kind: Secret
metadata:
  name: oidc-forward-auth-cert
  annotations:
    waf-cert-uploader.iits.tech/waf-id: "123"
  labels:
    waf-cert-uploader.iits.tech/enabled: "true"
data:
  tls.crt: LS0tLS1CRUdJ
  tls.key: LS0tLS1CRUdJ
```