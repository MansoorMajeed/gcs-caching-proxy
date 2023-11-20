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
skaffold dev --port-forward
```


## In Action

First request, `X-Nginx-Cache-Status: MISS`

```
curl -vso /dev/null 'localhost:8081/gcs-caching-proxy-test/hello.txt?q=1'

----snip----

< HTTP/1.1 200 OK
< Server: nginx/1.25.3
< Date: Mon, 20 Nov 2023 15:46:31 GMT
< Content-Type: text/plain
< Content-Length: 6
< Connection: keep-alive
< Etag: CPTsvPax0YIDEAE=
< Last-Modified: Mon, 20 Nov 2023 01:08:13 GMT
< X-Nginx-Cache-Status: MISS

```

Subsequent Request is cached by Nginx

```
curl -vso /dev/null 'localhost:8081/gcs-caching-proxy-test/hello.txt?q=1'

---snip---

< HTTP/1.1 200 OK
< Server: nginx/1.25.3
< Date: Mon, 20 Nov 2023 15:46:33 GMT
< Content-Type: text/plain
< Content-Length: 6
< Connection: keep-alive
< Etag: CPTsvPax0YIDEAE=
< Last-Modified: Mon, 20 Nov 2023 01:08:13 GMT
< X-Nginx-Cache-Status: HIT
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


## TODO

Consistent hashing with Bounded loads using Haproxy, which would increase the cache hit ratio.
We will add Nginx pods individually to the haproxy as backends instead of using a service, and
enable consistent hashing, that means the request to same URL will always go to the same pod.