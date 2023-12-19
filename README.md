# Helm chart for the [WAF certificate uploader](https://github.com/iits-consulting/waf-cert-uploader)

This documentation demonstrates how the [WAF certificate uploader](https://github.com/iits-consulting/waf-cert-uploader) can be configured and deployed to manage certificates automatically in the [Web Application Firewall (WAF)](https://docs.otc.t-systems.com/web-application-firewall/index.html) in [Open Telekom Cloud (OTC)](https://open-telekom-cloud.com/). The helm chart enables the automatic process of generating and attaching certificates to a given set of WAF domains, including their renewal process.


## Usage & Configuration

The helm chart can, for example, be deployed by terraform as described [here](https://github.com/iits-consulting/otc-terraform-template).
The following arguments must be set:
   
| Variable Name                                  | Explanation                                                                                                | Example                        |
|------------------------------------------------|------------------------------------------------------------------------------------------------------------|--------------------------------|
| `otcDomainName`                                 | The OTC account name                                                                  | `OTC-EU-DE-00000000`          |
| `tenantName`                                | Your OTC project name                                                                          | `eu-de_your_project`          |
| `email`                                   | The person to be notified on certificate events                                                                  |  `notifyme@telekom.de`        |
| `access_key`<br /> `secret_key`      | IAM Ak/Sk Pair to be authenticated with OTC<br />(no username and password is needed)                 |                             |
| `username`<br /> `password`          | IAM user credentials to be authenticated with OTC<br />(no Ak/Sk must be provided)     |                                             |
| `dockerhub`                          | Username and Access token to the dockerhub repository with the webhook docker image  |     <pre lang="yaml">{&#13;   accessToken: "4564dfsa",&#13;   username: "anyuser",&#13;}</pre>                                                          |
| `waf`                                | A list of WAF domain Ids with their corresponding CNAME record names     |      <pre lang="yaml">[&#13;  {&#13;    dnsName: "my.domain.com",&#13;    domainId: "abc123",&#13;  }&#13;]</pre>|

## Workflow explanation
This section provides a comprehensive overview of the implementation details and the realization of the aforementioned function.

### Creation of resources in the **Kubernetes Cluster (CCE)**

 - The **cert-manager**, responsible for automating the certification process.
     - The **letsencrypt cluster issuer**, tasked with issuing the CNAME record `my.domain.com`.
     - The **selfsigned cluster issuer**, responsible for issuing the WAF cert uploader webhook, allowing the communication between the Kubernetes API Server and the webhook.
     - A **CA certificate** for each CNAME record, containing the corresponding WAF domain ID in the secret template section.
     - A **selfsigned certificate** for the webhook.
 - A **secret** containing OTC-Credentials and which is mounted into the webhook container.
 - A **service** (type ClusterIP) to make the webhook deployment accessible from within the cluster.
 - Deployment of:
     - The webhook, which uploads the certificates to the WAF.
     - Upon startup, the mounted credentials secret is used to create an authenticated [gopher provider client](https://github.com/opentelekomcloud/gophertelekomcloud) and a service client for API calls.
 - **Mutating Webhook Configuration**:
     - Informs the Kubernetes API Server of the events that will trigger an admission review.
     - In this scenario, the API Server monitors updated secrets<br>with the following label: `"webhook-enabled" : "true"`.
     - Directs to the webhook service and endpoint.
 - **Traefik** (Ingress controller)
 - **Kyverno** (for default iits policies)

### Certification Process
- The cert-manager identifies the unsigned certificates and generates the objects necessary for the acme challenge process.
- It then produces a key pair and stores the private key in the certificate secret named `my.domain.com`. The WAF domain ID and the match label from the mutating webhook configuration are also transferred to the secret.
- The public key is sent to letsencrypt and a challenge token is received and stored in a file hosted by a small web application.
- The Ingress controller redirects incoming requests from<br> `http://my.domain.com/.well-known/acme-challenge/<TOKEN>` to the pod containing the challenge token file.
- letsencrypt attempts to request the token on port 80. If successful, it encrypts the public key with its private key and attaches this signature to the public key.
- The cert-manager receives the certificate chain and stores it in the certificate secret.

### Webhook Triggering
- With the occurrence of an update event on a secret with the match label from the webhook configuration, the API Server sends an admission review object to the webhook.
- The admission review comprises the old secret without the certificate chain as well as the target secret with the certificate chain. The mutating webhook can now manipulate the secret and either accept or reject the admission review.

### Certificate Uploading Process
- The webhook extracts the certificate content, WAF domain ID, and WAF certificate ID (if it exists initially) from the admission review object.
- The certificate content undergoes SHA256 encryption to generate a unique identifier, which serves as the name for the certificate.
- A request is made to the WAF API to retrieve all existing certificates, initiating a search process. If the certificate name already exists in the WAF, the process is terminated, and the admission review is accepted without any mutation.
- If the SHA256 name is not found in the WAF, the certificate is uploaded, and a certificate ID is received.
- The received certificate ID is then attached to the WAF using the domain ID from the certificate secret.
- The WAF domain is updated with an additional server address entry, enabling automatic forwarding of incoming and outgoing requests to port 443, and the certificate is utilized.
- If a WAF certificate ID exists in the certificate secret, the previous certificate is considered expired and is subsequently deleted.
- The admission review is accepted, and the secret is mutated to include an additional annotation with the new certificate ID.

# Workflow chart
![Workflow](flowchart/certuploader.svg)
