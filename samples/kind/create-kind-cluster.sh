config=$TMPDIR/create-kind-cluster-no-networking-policies.yaml
cat samples/kind/create-kind-cluster.yaml|sed '/networking/,$d' > $config
kind create cluster --config $config
rm -f $config
kubectx kind-kind
. $(dirname $0)/setup-requirements.sh
