
```sh
$ export SERVICE_ACCOUNT_KEY_PATH=/tmp/keys/poc-rtc.json
$ terraform init    
```

Creation du cluster master : 

- Planification : 

```sh
$ terraform plan \
    -var-file="env/master.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}"
```

- Application : 

```sh
$ terraform workspace new master
$ terraform workspace select master
$ terraform apply \
    -var-file="env/master.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}"
```

```sh
$ terraform workspace new slave
$ terraform workspace select slave
$ terraform apply \
    -var-file="env/slave.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}"
```

Debug : 

Install : 

```sh
$ gcloud container clusters get-credentials cassandra-europe-west1-b-master --zone europe-west1-b --project poc-rtc
$ kubens cassandra-demo
$ kubectl apply -f samples/cassandracluster-demo-gke-zone.yaml
```

```sh
$ gcloud container clusters get-credentials cassandra-europe-west1-c-slave --zone europe-west1-c --project poc-rtc
$ kubens cassandra-demo
$ kubectl apply -f samples/cassandracluster-demo-gke-zone.yaml
```

Destruction du cluster : 

```sh
$ terraform destroy \
    -var-file="env/master.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}"
```

> Installation du multi-casskop 
> Déploiement

Keep : 

```sh
$ gcloud config set project poc-rtc
$ gcloud dns managed-zones create "external-dns-test-gcp-trycatchlearn-fr" \
    --dns-name "external-dns-test.gcp.trycatchlearn.fr." \
    --description "Automatically managed zone by kubernetes.io/external-dns"
$ gcloud dns record-sets list \
    --zone "external-dns-test-gcp-trycatchlearn-fr" \
    --name "external-dns-test.gcp.trycatchlearn.fr." \
    --type NS
$ gcloud dns record-sets transaction start --zone "tracking-pdb"
$ gcloud dns record-sets transaction add ns-cloud-d{1..4}.googledomains.com. \
    --name "external-dns-test.gcp.trycatchlearn.fr." --ttl 300 --type NS --zone "tracking-pdb"
$ gcloud dns record-sets transaction execute --zone "tracking-pdb"
```