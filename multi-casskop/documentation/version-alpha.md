# Alpha Documentation

This is an early work on documentation not yet well structured, things I note here at this time :)

## What is Multi-CassKop

Multi-CassKop is a new Kubernetes operator that sits above CassKop operator.
CassKop allows to create a Cassandra cluster in a Kubernetes cluster, using the custom ressource **CassandraCluster**.

Multi-CassKop goal is to bring the ability to deploy a Cassandra cluster within different regions, each of them running
an independant Kubernetes cluster.

Multi-CassKop will be talking to several Kubernetes clusters APIs, and is managing a new custom ressource named
**MultiCasskop**.

The only goal of Multi-CassKop is to create **CassandraCluster** ressources in each of k8s clusters, where there is a
locally instance of CassKop that will deploy Cassandra cluster according to the required definition.

Multi-Casskop insure that the Cassandra nodes deployed by each local CassKop will be part of the same Cassandra ring by
managing a coherent creation of CassandraCluster objects from it's own MultiCasskop custom ressource.

## Software Organisation

This operator uses: 
- Operator sdk to setup the new CRD `MultiCasskop`.
- Admiralty's [multicluster-controller](https://github.com/admiraltyio/multicluster-controller) to allow talking to
  several kubernetes api servers.
  

## Launching MultiCassKop operator

MultiCassKop take into parameters a list of kubernetes contexts names as defined in the KUBECONFIG file.

example: 
```
./multiCassKop dex-sallamand-kaas-prod-priv-sph dex-sallamand-kaas-prod-priv-bgl
```

>**Note:** MultiCassKop will ONLY uses the first kubernetes passed as parameter to MultiCasskop objects

# How MultiCassKop works

MultiCassKop starts by iterrating on every contexts passed in parameters then it register the controller. 
The controller needs to be able to interract with MultiCasskop and CassandraCluster CRD objetcs.
In addition the controller needs to watch for MultiCasskop as it will need to react on any changes occurs on
thoses objects for the given namespace.


## MultiCasskop definition

Multi-CassKop introduce a new custom ressource and will have the charge to create CassandraCluster ressoruces in each
k8s cluster.

The Spec field of MultiCasskop has a `base` parameter which contain a valid CassandraCluster object.
It also have an `override` section that will allow specify part of the base CassandraCluster definition that will be
override depending on the target cluster.

Multi-CassKop make uses of kubectl context to target different clusters, so that the key of each override section must
be a valid kubectl context name.



### Override

Example of the override section

```
  override:
    dex-kaas-prod-priv-sph:
      spec:
        topology:
          dc:
            - name: dc1
              nodesPerRacks: 2
              numTokens: 256
              labels:
                location.dfy.orange.com/site : Valbonne
                location.dfy.orange.com/building : HT2
              rack:
                - name: rack1
                  labels: 
                    location.dfy.orange.com/room : Salle_1
                    location.dfy.orange.com/street : Rue_9
                - name: rack2
                  labels: 
                    location.dfy.orange.com/room : Salle_1
                    location.dfy.orange.com/street : Rue_10
                - name: rack3
                  labels: 
                    location.dfy.orange.com/room : Salle_1
                    location.dfy.orange.com/street : Rue_11

    dex-kaas-prod-priv-bgl:
      spec:
        #imagepullpolicy: Always
        topology:
          dc:
            - name: dc2
              nodesPerRacks: 2
              numTokens: 256
              labels:
                location.dfy.orange.com/site : Bagnolet
                location.dfy.orange.com/building : Immeuble_Gambetta
              rack:
                - name: rack4
                  labels: 
                    location.dfy.orange.com/room : Salle_B2
                    location.dfy.orange.com/street : Rue_3
                    location.dfy.orange.com/bay : "1"
                - name: rack5
                  labels: 
                    location.dfy.orange.com/room : Salle_B2
                    location.dfy.orange.com/street : Rue_6
                    location.dfy.orange.com/bay : "5"
                - name: rack6
                  labels: 
                    location.dfy.orange.com/room : Salle_B2
                    location.dfy.orange.com/street : Rue_5
                    location.dfy.orange.com/bay : "10"
```

> We can defined has many entries we want in the override sections which correspond also to the clusters we want to
> deploy onto.
