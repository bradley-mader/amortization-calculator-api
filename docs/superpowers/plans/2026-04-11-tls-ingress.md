# TLS Ingress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add TLS termination at the Kubernetes Ingress layer using nginx Ingress controller and cert-manager with Let's Encrypt on DOKS.

**Architecture:** Base Ingress resource with prod defaults in `k8s/base/`, patched per overlay. cert-manager ClusterIssuers (staging for dev, prod for stage/prod) referenced as resources by each overlay. No application code changes.

**Tech Stack:** Kubernetes Ingress (networking.k8s.io/v1), cert-manager (cert-manager.io/v1), Kustomize, nginx Ingress controller

---

### Task 1: Create cert-manager ClusterIssuer manifests

**Files:**
- Create: `k8s/base/cluster-issuer-staging.yaml`
- Create: `k8s/base/cluster-issuer-prod.yaml`

- [ ] **Step 1: Create the Let's Encrypt staging ClusterIssuer**

Create `k8s/base/cluster-issuer-staging.yaml`:

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

- [ ] **Step 2: Create the Let's Encrypt production ClusterIssuer**

Create `k8s/base/cluster-issuer-prod.yaml`:

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

- [ ] **Step 3: Validate YAML syntax**

Run:
```bash
cat k8s/base/cluster-issuer-staging.yaml | python3 -c "import sys,yaml; yaml.safe_load(sys.stdin)" && echo "staging: valid"
cat k8s/base/cluster-issuer-prod.yaml | python3 -c "import sys,yaml; yaml.safe_load(sys.stdin)" && echo "prod: valid"
```

Expected: Both print "valid" with no errors.

- [ ] **Step 4: Commit**

```bash
git add k8s/base/cluster-issuer-staging.yaml k8s/base/cluster-issuer-prod.yaml
git commit -m "feat: add cert-manager ClusterIssuer manifests for Let's Encrypt

Staging issuer for dev (no rate limits, untrusted certs).
Prod issuer for stage and prod (browser-trusted certs).
Both use HTTP-01 solver with nginx ingress class."
```

---

### Task 2: Create base Ingress resource and register it in Kustomization

**Files:**
- Create: `k8s/base/ingress.yaml`
- Modify: `k8s/base/kustomization.yaml`

- [ ] **Step 1: Create the base Ingress resource**

Create `k8s/base/ingress.yaml`:

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

- [ ] **Step 2: Add ingress.yaml to the base Kustomization**

Modify `k8s/base/kustomization.yaml` to become:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml
```

The only change is adding `- ingress.yaml` to the resources list. Do NOT add the ClusterIssuer files here — overlays reference those individually.

- [ ] **Step 3: Validate the base Kustomization builds**

Run:
```bash
kubectl kustomize k8s/base/
```

Expected: Output includes the Deployment, Service, and Ingress resources. The Ingress should show `api.bradley-mader.com` as the host with TLS configured.

- [ ] **Step 4: Commit**

```bash
git add k8s/base/ingress.yaml k8s/base/kustomization.yaml
git commit -m "feat: add base Ingress resource with prod TLS defaults

Ingress uses nginx class, cert-manager annotation for letsencrypt-prod,
and routes all paths to amortization-api service on port 8080.
Registered in base kustomization.yaml."
```

---

### Task 3: Update dev overlay — add Ingress patch and ClusterIssuer, remove LoadBalancer

**Files:**
- Modify: `k8s/overlays/dev/kustomization.yaml`

- [ ] **Step 1: Replace the dev overlay kustomization.yaml**

The current content is:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: dev
resources:
  - ../../base
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: dev-latest
patches:
  - patch: |
      apiVersion: v1
      kind: Service
      metadata:
        name: amortization-api
      spec:
        type: LoadBalancer
    target:
      kind: Service
      name: amortization-api
```

Replace the entire file with:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: dev
resources:
  - ../../base
  - ../../base/cluster-issuer-staging.yaml
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: dev-latest
patches:
  - patch: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: amortization-api
        annotations:
          cert-manager.io/cluster-issuer: letsencrypt-staging
      spec:
        tls:
          - hosts:
              - api-dev.bradley-mader.com
            secretName: amortization-api-tls-dev
        rules:
          - host: api-dev.bradley-mader.com
            http:
              paths:
                - path: /
                  pathType: Prefix
                  backend:
                    service:
                      name: amortization-api
                      port:
                        number: 8080
    target:
      kind: Ingress
      name: amortization-api
```

Changes from the original:
- Removed the LoadBalancer Service patch (Ingress handles external access).
- Added `../../base/cluster-issuer-staging.yaml` to resources.
- Added Ingress patch: overrides host to `api-dev.bradley-mader.com`, issuer to `letsencrypt-staging`, secret to `amortization-api-tls-dev`.

- [ ] **Step 2: Validate the dev overlay builds**

Run:
```bash
kubectl kustomize k8s/overlays/dev/
```

Expected: Output includes:
- Deployment in namespace `dev` with image `ghcr.io/bradley-mader/ammortizationcalculatorapi:dev-latest`
- Service (ClusterIP, not LoadBalancer) in namespace `dev`
- Ingress with host `api-dev.bradley-mader.com`, issuer annotation `letsencrypt-staging`, secret `amortization-api-tls-dev`
- ClusterIssuer `letsencrypt-staging` with the staging ACME server URL

- [ ] **Step 3: Commit**

```bash
git add k8s/overlays/dev/kustomization.yaml
git commit -m "feat: add TLS Ingress to dev overlay with staging Let's Encrypt

Removes LoadBalancer service patch (Ingress handles external access).
Adds letsencrypt-staging ClusterIssuer and patches Ingress to use
api-dev.bradley-mader.com with staging certificates."
```

---

### Task 4: Update stage overlay — add Ingress patch and ClusterIssuer

**Files:**
- Modify: `k8s/overlays/stage/kustomization.yaml`

- [ ] **Step 1: Update the stage overlay kustomization.yaml**

The current content is:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: stage
resources:
  - ../../base
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: stage-latest
patches:
  - patch: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: amortization-api
      spec:
        replicas: 2
    target:
      kind: Deployment
      name: amortization-api
```

Replace the entire file with:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: stage
resources:
  - ../../base
  - ../../base/cluster-issuer-prod.yaml
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: stage-latest
patches:
  - patch: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: amortization-api
      spec:
        replicas: 2
    target:
      kind: Deployment
      name: amortization-api
  - patch: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: amortization-api
      spec:
        tls:
          - hosts:
              - api-stage.bradley-mader.com
            secretName: amortization-api-tls-stage
        rules:
          - host: api-stage.bradley-mader.com
            http:
              paths:
                - path: /
                  pathType: Prefix
                  backend:
                    service:
                      name: amortization-api
                      port:
                        number: 8080
    target:
      kind: Ingress
      name: amortization-api
```

Changes from the original:
- Added `../../base/cluster-issuer-prod.yaml` to resources.
- Added Ingress patch: overrides host to `api-stage.bradley-mader.com`, secret to `amortization-api-tls-stage`.
- No issuer annotation override needed — base already uses `letsencrypt-prod`, which is correct for stage.
- Existing replicas patch preserved unchanged.

- [ ] **Step 2: Validate the stage overlay builds**

Run:
```bash
kubectl kustomize k8s/overlays/stage/
```

Expected: Output includes:
- Deployment in namespace `stage` with 2 replicas
- Service (ClusterIP) in namespace `stage`
- Ingress with host `api-stage.bradley-mader.com`, issuer annotation `letsencrypt-prod`, secret `amortization-api-tls-stage`
- ClusterIssuer `letsencrypt-prod` with the production ACME server URL

- [ ] **Step 3: Commit**

```bash
git add k8s/overlays/stage/kustomization.yaml
git commit -m "feat: add TLS Ingress to stage overlay with prod Let's Encrypt

Adds letsencrypt-prod ClusterIssuer and patches Ingress to use
api-stage.bradley-mader.com with production certificates."
```

---

### Task 5: Update prod overlay — add ClusterIssuer resource

**Files:**
- Modify: `k8s/overlays/prod/kustomization.yaml`

- [ ] **Step 1: Update the prod overlay kustomization.yaml**

The current content is:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: prod
resources:
  - ../../base
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: prod-latest
patches:
  - patch: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: amortization-api
      spec:
        replicas: 3
        template:
          spec:
            containers:
              - name: server
                resources:
                  requests:
                    cpu: 250m
                    memory: 128Mi
                  limits:
                    cpu: 500m
                    memory: 256Mi
    target:
      kind: Deployment
      name: amortization-api
```

Replace the entire file with:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: prod
resources:
  - ../../base
  - ../../base/cluster-issuer-prod.yaml
images:
  - name: amortization-api
    newName: ghcr.io/bradley-mader/ammortizationcalculatorapi
    newTag: prod-latest
patches:
  - patch: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: amortization-api
      spec:
        replicas: 3
        template:
          spec:
            containers:
              - name: server
                resources:
                  requests:
                    cpu: 250m
                    memory: 128Mi
                  limits:
                    cpu: 500m
                    memory: 256Mi
    target:
      kind: Deployment
      name: amortization-api
```

The only change is adding `- ../../base/cluster-issuer-prod.yaml` to resources. No Ingress patch needed — base Ingress already has `api.bradley-mader.com`, `letsencrypt-prod`, and `amortization-api-tls` which are the correct prod values.

- [ ] **Step 2: Validate the prod overlay builds**

Run:
```bash
kubectl kustomize k8s/overlays/prod/
```

Expected: Output includes:
- Deployment in namespace `prod` with 3 replicas and higher resource limits
- Service (ClusterIP) in namespace `prod`
- Ingress with host `api.bradley-mader.com`, issuer annotation `letsencrypt-prod`, secret `amortization-api-tls`
- ClusterIssuer `letsencrypt-prod` with the production ACME server URL

- [ ] **Step 3: Commit**

```bash
git add k8s/overlays/prod/kustomization.yaml
git commit -m "feat: add letsencrypt-prod ClusterIssuer to prod overlay

Base Ingress already has correct prod values (api.bradley-mader.com,
letsencrypt-prod issuer). Only change is adding the ClusterIssuer
resource reference."
```

---

### Task 6: Validate all overlays build correctly

**Files:**
- None (validation only)

- [ ] **Step 1: Build all overlays and verify no errors**

Run:
```bash
kubectl kustomize k8s/overlays/local-dev/
kubectl kustomize k8s/overlays/dev/
kubectl kustomize k8s/overlays/stage/
kubectl kustomize k8s/overlays/prod/
```

Expected: All four commands produce valid YAML output with no errors.

- [ ] **Step 2: Verify local-dev does NOT include Ingress**

Run:
```bash
kubectl kustomize k8s/overlays/local-dev/ | grep "kind: Ingress"
```

Expected: Output shows `kind: Ingress`. This is expected because local-dev inherits from base which now includes `ingress.yaml`. The Ingress will be present in the rendered output but will have no effect without an Ingress controller running in Minikube. If you want to explicitly exclude it from local-dev, you can add a patch to remove it — but it's harmless as-is.

- [ ] **Step 3: Verify dev uses staging issuer**

Run:
```bash
kubectl kustomize k8s/overlays/dev/ | grep "letsencrypt-staging"
```

Expected: Matches in both the ClusterIssuer resource name and the Ingress annotation.

- [ ] **Step 4: Verify stage and prod use prod issuer**

Run:
```bash
kubectl kustomize k8s/overlays/stage/ | grep "letsencrypt-prod"
kubectl kustomize k8s/overlays/prod/ | grep "letsencrypt-prod"
```

Expected: Both show matches for `letsencrypt-prod` in the ClusterIssuer and Ingress annotation.

- [ ] **Step 5: Verify each environment has the correct hostname**

Run:
```bash
kubectl kustomize k8s/overlays/dev/ | grep "api-dev.bradley-mader.com"
kubectl kustomize k8s/overlays/stage/ | grep "api-stage.bradley-mader.com"
kubectl kustomize k8s/overlays/prod/ | grep "api.bradley-mader.com"
```

Expected: Each command produces matches confirming the correct hostname is set in the Ingress TLS hosts and rules.
