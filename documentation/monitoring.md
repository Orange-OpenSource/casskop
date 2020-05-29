# Monitoring

## Install Prometheus operator

```
helm install prometheus-monitoring stable/prometheus-operator --set prometheusOperator.createCustomResource=false --set grafana.plugins="{briangann-gauge-panel,grafana-clock-panel,grafana-piechart-panel,grafana-polystat-panel,savantly-heatmap-panel,vonage-status-panel}" --set grafana.image.tag=7.0.1
```

## Install our dashboard

```
kubectl apply -f monitoring/dashboards/
```

We provide a dashboard that is used to monitor clusters in production. Thanks to Ahmed Eljami for his contribution.

## Install Service monitor

You can update our service monitor if you want to create different service monitors like one per cluster for example. Otherwise just install ours:
```
kubectl apply -f monitoring/servicemonitor/prometheus-cassandra-service-monitor.yaml
```

Our ServiceMonitor monitors services created by CassKop that have label k8s-app set to exporter-cassandra-jmx. This service references the jmx exporter that is used by CassKop.