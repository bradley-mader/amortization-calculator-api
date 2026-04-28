# TLS Ingress Design for AmmortizationCalculatorAPI

## Overview

Add TLS termination at the Kubernetes Ingress layer for the AmmortizationCalculatorAPI. The Go application and Dockerfile remain unchanged — encryption is handled entirely by the nginx Ingress controller with certificates managed by cert-manager and Let's Encrypt.

## Platform

- **Kubernetes:** DigitalOcean Kubernetes Service (DOKS)
- **Ingress controller:** nginx (DOKS 1-click add-on)
- **Certificate management:** cert-manager with Let's Encrypt ACME

## Traffic Flow

```
Client → HTTPS (443) → nginx Ingress Controller → HTTP (8080) → amortization-api Service → Pod (h2c)
```

TLS terminates at the Ingress controller. Internal cluster traffic remains cleartext h2c, which is standard for service mesh-less clusters on DOKS.

## Environment Mapping

| Environment | Host                          | ClusterIssuer      | TLS Secret Name              |
|-------------|-------------------------------|--------------------|------------------------------|
| dev         | api-dev.bradley-mader.com     | letsencrypt-staging | amortization-api-tls-dev    |
| stage       | api-stage.bradley-mader.com   | letsencrypt-prod    | amortization-api-tls-stage  |
| prod        | api.bradley-mader.com         | letsencrypt-prod    | amortization-api-tls        |
| local-dev   | N/A (port-forward, no Ingress)| N/A                 | N/A                         |

## File Changes

### New Files

#### `k8s/base/ingress.yaml`

Base Ingress resource with prod defaults. Overlays patch host, TLS, and issuer as needed.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: amortization-api
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.bradley-mader.com
      secretName: amortization-api-tls
  rules:
    - host: api.bradley-mader.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: amortization-api
                port:
                  number: 8080
```

- `ingressClassName: nginx` — matches DOKS 1-click nginx Ingress controller.
- `pathType: Prefix` with `/` — catches all paths including ConnectRPC RPC paths (e.g., `/amortization.v1.AmortizationService/Calculate`).
- Base defaults to prod values so that an unpatched overlay gets the safest configuration.

#### `k8s/base/cluster-issuer-staging.yaml`

Let's Encrypt staging issuer for dev. Issues untrusted certificates but has no rate limits.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: mader.bradley@gmail.com
    privateKeySecretRef:
      name: letsencrypt-staging-key
    solvers:
      - http01:
          ingress:
            class: nginx
```

#### `k8s/base/cluster-issuer-prod.yaml`

Let's Encrypt production issuer for stage and prod. Issues browser-trusted certificates.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: mader.bradley@gmail.com
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
      - http01:
          ingress:
            class: nginx
```

- HTTP-01 solver — simplest ACME challenge type, works out of the box with nginx Ingress on DOKS.
- ClusterIssuer (not Issuer) — cluster-scoped, shareable across namespaces, only needs to be deployed once.

### Modified Files

#### `k8s/base/kustomization.yaml`

Add `ingress.yaml` to the resources list. ClusterIssuers are NOT added here — they are referenced by individual overlays to control which issuer each environment uses.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml
```

#### `k8s/overlays/dev/kustomization.yaml`

- Add `cluster-issuer-staging.yaml` as a resource.
- Remove the existing LoadBalancer Service patch (Ingress handles external access now).
- Add Ingress patch: host `api-dev.bradley-mader.com`, issuer `letsencrypt-staging`, secret `amortization-api-tls-dev`.

#### `k8s/overlays/stage/kustomization.yaml`

- Add `cluster-issuer-prod.yaml` as a resource.
- Add Ingress patch: host `api-stage.bradley-mader.com`, secret `amortization-api-tls-stage`.
- Existing replicas patch remains unchanged.

#### `k8s/overlays/prod/kustomization.yaml`

- Add `cluster-issuer-prod.yaml` as a resource.
- No Ingress patch needed — base Ingress already has prod values.
- Existing replicas/resources patch remains unchanged.

#### `k8s/overlays/local-dev/kustomization.yaml`

No changes. local-dev does not use Ingress or cert-manager.

## Prerequisites (Not Part of This Implementation)

These are cluster-level and DNS requirements that must be satisfied before the manifests work:

1. **Install nginx Ingress controller** on DOKS (1-click add-on or Helm chart).
2. **Install cert-manager** on the cluster (`kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.2/cert-manager.yaml` or Helm).
3. **Create DNS A records** pointing to the nginx Ingress controller's external IP:
   - `api-dev.bradley-mader.com` → Ingress controller external IP
   - `api-stage.bradley-mader.com` → Ingress controller external IP
   - `api.bradley-mader.com` → Ingress controller external IP
   - Get the IP via: `kubectl get svc -n ingress-nginx`
4. HTTP-01 challenges will fail if DNS is not pointing to the cluster.

## What Does NOT Change

- Go application code (`cmd/server/main.go`, `internal/`)
- Dockerfile
- Proto definitions
- local-dev overlay
- CI/CD workflow (`.github/workflows/docker-publish.yaml`)
