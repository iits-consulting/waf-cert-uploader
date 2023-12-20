Example Secret:
```yaml
apiVersion: v1
type: kubernetes.io/tls
kind: Secret
metadata:
  name: oidc-forward-auth-cert
  annotations:
    #optional, if not provided it will auto create it
    waf-manager.iits.tech/waf-id: "123"
  labels:
    waf-manager.iits.tech/enabled: "true"
data:
  tls.crt: LS0tLS1CRUdJ
  tls.key: LS0tLS1CRUdJ
```