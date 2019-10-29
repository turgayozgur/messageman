# Examples

## Sidecar

```bash
docker build -t messageman .
cd examples/job
docker-compose -f docker-compose-sidecar.yml up
```

## Gateway

```bash
docker build -t messageman .
cd examples/job
docker-compose -f docker-compose-gateway.yml up
```