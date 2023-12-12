## WAF cert uploader

This service is a kubernetes mutating webhook. It manages certificates for a set of given domain names and
attaches them to a Web Application Firewall (WAF) in [Open Telekom Cloud (OTC)](https://open-telekom-cloud.com/).
It is intended to be used together with a set of kubernetes resources and thus needs to be deployed via a [helm chart](https://github.com/iits-consulting/waf-cert-uploader/tree/gh-pages).
An example project and the usage guide for the helm chart can be found [here](https://github.com/iits-consulting/waf-cert-uploader-terraform).
