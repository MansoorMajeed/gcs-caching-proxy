# GCS Proxy with Consistent Hash Caching

This is a simple Google Cloud Storage proxy with caching along with consistent hashing built with Kubernetes in mind.

## Give me TL;DR

- You use Google Cloud Storage (GCS) to store files which are read by some of your service(s)
- They retrieve from GCS quite a lot and the egress cost is getting out of control
- You can introduce a local disk cache that will help a lot, depending on how distributed your services are.
    - For example, if you use Kubernetes and your service have 50 pods, even if you cache the objects in local disk, there is only 1/50 chance you hit that for every request.
- With the gcs-caching-proxy, you can make sure that once an object is cached, it will always (almost) be served from the cache until the cache expires
    - This is made possible via consistent hashing using Haproxy. Read more about it [HERE](https://docs.haproxy.org/2.8/configuration.html#hash-type)
    - Nginx does all the caching, we are not re-inventing caching. 
- Works great with Kubernetes, works without Kubernetes too, you just need to be able to run Haproxy and Nginx along with the Binary of the gcs-handler

### Some Diagrams

#### Without the proxy

Your service pods/VMs are reading directly from GCS. You may have some sort of disk caching, this is fine for the most part unless the number of pods/VMs
grow to a large number, then you are looking at a lot more retrieval from GCS causing a ton more egress cost

![image](https://github.com/MansoorMajeed/gcs-caching-proxy/assets/12676196/c5240b32-b4d3-4392-8933-18cbc9d2fd2d)

#### With the proxy and caching setup

- You can either run the `gcs-handler` (the proxy that reads from GCS) as a sidecar or as a separate service
- Nginx uses the `gcs-handler` as the upstream. Check the config [HERE](./kubernetes/nginx/config.yml)
- Haproxy uses the Nginx pods addresses as the backend. Check the Haproxy config [HERE](./kubernetes/haproxy/config.yml)
    - The reason why it uses the pods address directly is for consistent hashing. If you don't want to use consistent hashing, you can just use a service in front of Nginx

![image](https://github.com/MansoorMajeed/gcs-caching-proxy/assets/12676196/c260f900-1835-4312-9423-8ffd05527900)

## Ok, but why?

Because $$$$$

You don't need any of these if your GCS egress costs are not a problem. But as the scale goes up, so does the egress cost.

## Local Setup -  Without Kubernetes

### gcs-handler

This is the service that reads from GCS.

To set it up without Kubernetes, you can build the `gcs-handler` go service from [HERE](./cache-service/gcs-handler/)

To build a binary directly, just do `go build`.

There is a `Dockerfile`, so you just gotta do `docker build` in that directory for the image, if you want to use Docker.

### Nginx

Nginx does all the caching. You can run your own Nginx however you like. Get the config from [HERE](./kubernetes/nginx/config.yml) and change it to your needs.

### Haproxy

Haproxy is completely optional, but is useful if you want to run the whole thing at scale and want
consistent hashing.

## Local Setup - With Skaffold and Kubernetes

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


## Cache Purging

Purging needs to be done on Nginx, should be pretty straight forward to do depending on how you plan on using it.
Nginx has some docs [HERE](https://docs.nginx.com/nginx/admin-guide/content-cache/content-caching/#purging-content-from-the-cache)
