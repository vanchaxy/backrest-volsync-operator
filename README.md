# Backrest VolSync Operator (prototype)

Prototype Kubernetes operator that watches VolSync `ReplicationSource` / `ReplicationDestination` objects and registers the referenced Restic repository in Backrest via the Backrest API.

## What it does

- Reads `spec.restic.repository` from a VolSync object.
- Reads the referenced Secret keys:
  - `RESTIC_REPOSITORY` -> Backrest repo URI
  - `RESTIC_PASSWORD` -> Backrest repo password
  - All other non-`RESTIC_*` keys -> Backrest `repo.env` as `KEY=value` (or restrict via `spec.repo.envAllowlist`).
- Calls Backrest `AddRepo` (upsert) using the generated Connect client.
- Stores only a hash + non-sensitive status fields in the CR status.

## Install CRD + RBAC + sample

```sh
kubectl apply -k config/
kubectl apply -f config/samples.yaml
```

## Backrest auth (optional)

Create a Secret referenced by `spec.backrest.authRef.name` in the same namespace as the Binding:

- Bearer token:
  - key: `token`
- Basic auth:
  - keys: `username`, `password`

## Build + deploy

This operator is standalone (no local `replace ../backrest`).

### Local build

```sh
make fmt
make test
```

Build and push an image to a registry reachable by your cluster nodes (Talos uses containerd):

```sh
make docker-build IMAGE=backrest-volsync-operator:dev
```

### GitHub Actions (GHCR)

The workflow builds/pushes `ghcr.io/<owner>/<repo>` on pushes to `main` and tags.
The default deployment image is set to `ghcr.io/jogotcha/backrest-volsync-operator:latest`.

If your repo name differs, update the `image:` in `config/deployment.yaml`.

### Deploy

Ensure the `monitoring` namespace exists:

```sh
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
```

Apply the operator resources:

```sh
kubectl apply -k config/
```

If pulling from a private GHCR image, create an image pull secret and uncomment `imagePullSecrets` in `config/deployment.yaml`:

```sh
kubectl -n monitoring create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-pat-with-read:packages> \
  --docker-email=unused@example.com
```

```sh
kubectl apply -f config/deployment.yaml
```

## Create a Binding

```yaml
apiVersion: backrest.garethgeorge.com/v1alpha1
kind: BackrestVolSyncBinding
metadata:
  name: my-app
  namespace: monitoring
spec:
  backrest:
    url: http://backrest.monitoring.svc:9898
  source:
    kind: ReplicationSource
    name: my-app-data
  repo:
    idOverride: my-app-data
    autoUnlock: true
```

## Notes / security

Backrest stores `Repo.password` as plaintext in its config file. The operator avoids calling `GetConfig` and never writes credentials into status or logs, but Backrest’s own at-rest format remains unchanged.

## Verify

Watch the controller come up:

```sh
kubectl -n monitoring rollout status deploy/backrest-volsync-operator
kubectl -n monitoring logs -f deploy/backrest-volsync-operator -c manager
```

Check Binding status (no secret material is stored in status):

```sh
kubectl -n monitoring get backrestvolsyncbindings
kubectl -n monitoring get backrestvolsyncbinding <name> -o yaml
```

Confirm Backrest shows the repo:

- Use the Backrest UI (or port-forward the service) and verify the repo ID exists/updates.
- The operator logs a line like `Backrest repo applied` with the repo ID on success.
