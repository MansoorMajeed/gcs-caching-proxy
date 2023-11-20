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
