kind create cluster --config samples/kind/create-kind-cluster.yaml
kind get nodes | xargs -n1 -I {} docker exec {} sysctl -w net.ipv4.conf.all.rp_filter=0
export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
curl https://docs.projectcalico.org/v3.8/manifests/calico.yaml| kubectl apply -f -
. $(dirname $0)/setup-requirements.sh
kubectl apply -f $(dirname $0)/../network-policies.yaml