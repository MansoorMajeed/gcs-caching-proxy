# GCS Proxy with Caching

## Local Setup

### Setup MiniKube

### Install Skaffold

### Creating the GCS secret

Set the bucket name in `kubernetes/gcs-handler/deployment.yml` env variable

Create service account to read from the bucket and it should have at least `buckets.get` and `Object Reader`

```
kubectl create secret generic gcs-service-account \
  --from-file=key.json=/path/to/your/service-account.json
```


### Bringing it up

```
skaffold dev --port-forward
```