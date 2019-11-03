# Messageman

Lightweight, distributed job/worker managing application. Everything is a simple HTTP request. You can use it as a sidecar or a gateway. Distributed messages handled by built in RabbitMQ implementation.

## Highlights

* Language agnostic way to implement distributed messaging. Best fit for microservices.
* Everything is a simple HTTP request.
* Lightweight. Written with golang. It doesn't serialize/deserialize the message when proxying. It only passes the body as a byte array by using fasthttp module.
* `Sidecar` and `gateway` modes.
* Super easy configuration.
* Connection and channel management.
* Failure management. (Retry logic.)
* Recover functionality. Do not worry about connection/channel loses.
* Built in `prometheus exporter` and `grafana dashboard`.
* Ready to use `helm` chart for gateway mode.
* Simple example for sidecar mode on k8s.
* Fully compatible with containerized environments such as Kubernetes.

Check the docker-compose [examples](examples/README.md) to run the complete environment(2 api to send and receive messages, messageman, rabbitmq server, prometheus and grafana).

![examples](docs/img/messageman-gateway.svg)
![examples](docs/img/messageman-sidecar.svg)

## Usage

```bash
docker run --rm --name messageman -p 8015:8015 turgayozgur/messageman
```

Thats it! The messageman is up and ready to running with default configurations. For now, no any service registered, even so we can still send messages to queues.

### Send messages

```bash
curl "http://localhost:8015/queue?name=greeting" -d '{"say":"hi!"}'
```

Make sure you have a rabbitmq server available at this url: `amqp://guest:guest@localhost:5672/`
```bash
docker run --rm --name rabbitmq -p 5672:5672 rabbitmq:3.8
```

*Note:* For gateway mode, specifying the service name on the header named `x-service-name` is recommended. Otherwise, all queue requests will use the default connection.

Everyone can call the `queue` endpoint with a queue name. What if we need to receive messages when the message arrived to the queue? At this point, just provide a messageman config that contain details about the receiver service endpoints.

### Receive messages

messageman.yml
```yaml
services:
  - name: ReceiverApiName
    url: http://localhost:81 # your receiver api endpoint.
    workers:
      - path: /api/welcome
        queue: greeting
```

Run messageman with this config file. Also you can specify the config file location with using `-c` or `--config-file` arguments.

```bash
docker run --rm --name messageman -p 8015:8015 --volume=/path/to/messageman.yml:/app/messageman.yml turgayozgur/messageman
```

If the message queued to the `greeting` queue, messagemen will post the body `{"say":"hi!"}` to url `http://localhost:81/api/welcome`

## Complete example of messageman.yml

```yaml
mode: gateway # sidecar, gateway
metric: 
  enabled: true # available at /metrics path.
  exporter: prometheus
rabbitmq:
  url: amqp://guest:guest@localhost:5672/
services:
  - name: ReceiverApiName
    url: http://localhost:81
    readiness:
      path: /readiness
    workers:
      - path: /api/email/send
        queue: email
  - name: SenderApiName
    url: http://localhost:82
    readiness:
      path: /readiness
```

*Note:* Only one service allowed for `sidecar` mode. Workers are not required.

## Gateway mode on k8s

Before applying the helm template, create your own `values.yaml` file from the default one. `./.helm/values.yaml`

```bash
helm upgrade --install messageman --values values.yaml ./.helm
```

Thats it!. You can access the messageman with this endpoint `messageman.yournamespace.svc.cluster.local`

## Sidecar mode on k8s

Apply a configmap special for the service you want to deploy.

messageman.yourservicename.configmap.yml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: messageman-yourservicename
data:
  messageman.yml: |
    mode: sidecar
    rabbitmq:
      url: amqp://guest:guest@rabbitmqhostname:5672/
    services:
      - name: SenderApiName
        url: http://localhost:82
        readiness:
          path: /readiness
```

```bash
kubectl apply -f messageman.yourservicename.configmap.yml
```

Add the messageman container to containers of your k8s deployment configuration.

yourservicename.deployment.yml
```yaml
apiVersion: apps/v1
kind: Deployment
...
... # other lines
...
spec:
  ...
  ... # other lines.
  ...
  spec:
    containers:
    ...
    ... # other containers
    ...
    - name: messageman
      image: turgayozgur/messageman:latest
      ports:
      - containerPort: 8015
      livenessProbe:
        failureThreshold: 5
        httpGet:
        path: "/healthz"
        port: 8015
        scheme: HTTP
        initialDelaySeconds: 10
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 10
      env:
        - name: MESSAGEMAN_PORT
          value: 8015
        - name: LOG_LEVEL
          value: info
        - name: LOG_HUMANIZE
          value: true
      volumeMounts:
        - name: messageman
          mountPath: /app
    volumes:
      - name: messageman
        configMap:
          name: messageman-yourservicename
          items:
            - key: messageman.yml
              path: messageman.yml
```

```bash
kubectl apply -f yourservicename.deployment.yml
```

## License ##
The Tossit is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).
