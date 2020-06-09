---
id: 10_nodes_management
title: Nodes Management
sidebar_label: Nodes Management
---

CassKop in duo with the Cassandra docker Image is responsible of the lifecycle of the Cassandra nodes.

## HealthChecks

Healthchecks are periodical tests which verify Cassandra's health. When the healthcheck fails, Kubernetes can assume
that the application is not healthy and attempt to fix it. Kubernetes supports two types of Healthcheck probes : 
- Liveness probes
- Readiness probes.

You can find more details in the [Kubernetes
documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#configure-probes).

Both `livenessProbe` and `readinessProbe` support two additional options:
- `initialDelaySeconds`: defines the initial delay before the probe is tried for the first time. Default is 15 seconds
- `timeoutSeconds`: defines the timeout of the probe. CassKop uses 20 seconds.
- `periodSeconds`: the period to wait between each call to a probe: CassKop uses 40 seconds.


You are now able to override this default values using the following fields in to the `CassandraCluster` definition : 

- `livenessInitialDelaySeconds`: defines initial delay for the liveness probe of the main
- `livenessHealthCheckTimeout`: defines health check timeout for the liveness probe of the main
- `livenessHealthCheckPeriod`: defines health check period for the liveness probe of the main
- `livenessFailureThreshold`: defines failure threshold for the liveness probe of the main
- `livenessSuccessThreshold`: defines success threshold for the liveness probe of the main

- `readinessInitialDelaySeconds`: defines initial delay for the readiness probe of the main
- `readinessHealthCheckTimeout`: defines health check timeout for the readiness probe of the main
- `readinessHealthCheckPeriod`: defines health check period for the readiness probe of the main
- `readinessFailureThreshold`: defines failure threshold for the readiness probe of the main
- `readinessSuccessThreshold`: defines success threshold for the readiness probe of the main

## Pod lifeCycle

The Kubernetes Pods allows user to define specific hooks to be executed at some times

### PreStop

CassKop uses the PreStop hook to execute some commands before the pod is going to be killed.
In first iteration we were executing a `nodetool drain` and it used to make some unpredictable behavior.
At the time of writing this document, there is no `PreStop` action executed. 


## Prometheus metrics export

We currently use the CoreOS Prometheus Operator to export the Cassandra nodes metrics. We must create a serviceMonitor
object in the prometheus namespaces, pointing to the exporter-prometheus-service which is created by CassKop:


```yaml
$ cat samples/prometheus-cassandra-service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prometheus-cassandra-jmx
  labels:
    k8s-apps: cassandra-k8s-jmx
    prometheus: kube-prometheus
    component: cassandra
    app: cassandra
spec:
  jobLabel: kube-prometheus-cassandra-k8s-jmx
  selector:
    matchLabels:
      k8s-app: exporter-cassandra-jmx
  namespaceSelector:
      matchNames:
      - cassandra
      - cassandra-demo
  endpoints:
  - port: promjmx
    interval: 15s
```

Actually the Cassandra nodes use the work of Oleg Glusahak https://github.com/oleg-glushak/cassandra-prometheus-jmx but
this may change in the futur.