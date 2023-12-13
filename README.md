# GCS Proxy with Caching

## Local Setup

### Setup MiniKube

Go [HERE](https://minikube.sigs.k8s.io/docs/start/)

### Install Skaffold

```
brew install skaffold
```

### Creating the GCS secret

Set the bucket name in `kubernetes/gcs-handler/deployment.yml` env variable

Create service account to read from the bucket and it should have at least `buckets.get` and `Object Reader`

```
kubectl create secret generic gcs-service-account \
  --from-file=key.json=/path/to/your/service-account.json
```


### Bringing it up

```
skaffold dev
```


## In Action

First request, `X-Nginx-Cache-Status: MISS`, note the `x-nginx-server`

```
➜  ~ curl -v  localhost:8000/gcs-caching-proxy-test/test1.txt

< HTTP/1.1 200 OK
---snip---
< x-nginx-cache-status: MISS
< x-nginx-server: nginx-4
< x-haproxy-hostname: haproxy-55679b597f-9vwxs
<
test1
```

Subsequent Request is cached by Nginx, we can also see that haproxy made the request to the same nginx backend, using consistent hashing

```
➜  ~ curl -v  localhost:8000/gcs-caching-proxy-test/test1.txt

< HTTP/1.1 200 OK
---snio---
< x-nginx-cache-status: HIT
< x-nginx-server: nginx-4
< x-haproxy-hostname: haproxy-55679b597f-9vwxs
<
test1
➜  ~
```

## Cache Config

All caching is handled by Nginx, no reinventing caching.
Check the config [HERE](./kubernetes/nginx/config.yml)


These are important, must change to fit your needs
```
proxy_cache_valid 200 60m;     # Cache 200 for 60 minutes
proxy_cache_valid 404 1m;      # Cache 404 responses for 1 minute
proxy_cache_valid any 10s;     # Cache all other responses for 10s
proxy_cache_key "$host$request_uri";
```


## Haproxy Consistent Hashing with Bounded Loads

Consistent hashing with Bounded loads using Haproxy, which would increase the cache hit ratio.
We will add Nginx pods individually to the haproxy as backends instead of using a service, and
enable consistent hashing, that means the request to same URL will always go to the same pod.

Check the config [HERE](./kubernetes/haproxy/config.yml)

Here is the relevant part the does consistent hashing.
```
hash-type consistent
hash-balance-factor 150
```

> `factor` is a percentage greater than 100. For example, if `factor`` is 150,
> then no server will be allowed to have a load more than 1.5 times the average.
> If server weights are used, they will be respected.
