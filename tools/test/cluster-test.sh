#!/bin/zsh
test_image=$1

unhealth_nodes=$(kubectl get nodes | awk 'NR>1 && $2 != "Ready"{print $1}' | sort)
echo "========================== unhealthy nodes =========================="
echo ${unhealth_nodes}

health_nodes=$(kubectl get nodes | awk 'NR>1 && $2 == "Ready"{print $1}' | sort)
nodes_num=$(echo $health_nodes | wc -l)

function create_test_pods {
cat <<EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 500
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: $test_image
        ports:
        - containerPort: 80
EOF
}

function do_test {
result=$(kubectl get pods  -o wide | grep test-nginx-deployment | awk '$3=="Running"{print $6" "$7}')

nodes=$(echo $result | awk '{print $2}' | awk '!a[$0]++' | sort)

diff_nodes=$(comm -23 <(echo "$health_nodes" | sort) <(echo "$nodes" | sort))
echo "========================== fail to deploy pods on nodes =========================="
echo $diff_nodes

failed_nodes=""
IFS=$'\n'
for i in $result
do
if [ $(curl -I -m 1 -o /dev/null -s -w %{http_code} $(echo $i | awk '{print $1}')) -ne 200 ]; then
failed_nodes=$failed_nodes" "$(echo $i | awk '{print $2}')
fi
done

echo "========================== fail to validate pod on nodes =========================="
echo ${failed_nodes# }
}


#create_test_pods
#sleep 10
do_test
#kubectl delete deployment test-nginx-deployment