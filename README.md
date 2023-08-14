# Jaeger Demo

## Getting Started

### Prepare The Cluster

#### Cluster Creation

Create a Kubernetes cluster using `kind` with ingress configurations so that we can
access the service locally:

```bash
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

#### Ingress Controller For Accessing Services

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```
```bash
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```
#### Jaeger with Elasticsearch v7

```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
```

```bash
helm -n default install --wait jaeger jaegertracing/jaeger \
  --set provisionDataStore.cassandra=false \
  --set provisionDataStore.elasticsearch=true \
  --set storage.type=elasticsearch \
  --set elasticsearch.replicas=1 \
  --set elasticsearch.minimumMasterNodes=1 \
  --set collector.service.otlp.grpc.name=otlp \
  --set collector.service.otlp.http.name=otlp-http \
  --set query.ingress.enabled=true \
  --set 'query.ingress.hosts[0]=jaeger-127-0-0-1.nip.io'
```

### Deploy Our Application

```bash
helm install --wait jaeger-demo oci://ghcr.io/enesonus/jaeger-demo/jaeger-demo --version 0.2.2 \
  --set ingress.domain=enesonus-127-0-0-1.nip.io
```

### Access The Application

```bash
curl '127.0.0.1:8080/album?id=1'
```

## Query Traces

### Query Traces Using Jaeger UI

Go to http://jaeger-127-0-0-1.nip.io and query traces!