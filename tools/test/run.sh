#!/usr/bin/env bash
image="ipallocator:$TAG"

function tar_src {
base_dir="../.."
curr_dir=$PWD
cd $base_dir
tar czvf $curr_dir/src-$TAG.tgz .
cd -
}

function k8s_build {
cat <<EOF | kubectl apply -f -
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: test-ipallocator-server-config
data:
  config.conf: |-
    {
      "port": 9000,
      "range_cidr": "10.0.0.0/16",
      "subnet": "10.0.0.0/16",
      "base_key": "/ipallocator/podips",
      "registry_key": "/ipallocator/ipregistry",
      "etcd_ips": [
        "$ETCD_IP"
      ],
      "routes": [
        {
          "dst": "0.0.0.0/0",
          "gw": "10.0.0.1"
        }
      ]
    }
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test-ipallocator-server
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: test-ipallocator-server
    spec:
      nodeSelector:
        node: test
      containers:
      - name: ipallocator-server
        image: $REPO_URL/$image
        command:
        - /usr/local/bin/ipallocator_server
        - -config=/var/lib/ipallocator-server/config.conf
        - -alsologtostderr
        ports:
        - containerPort: 9000
        volumeMounts:
        - mountPath: /var/lib/ipallocator-server
          name: ipallocator-server
      - name: ipalloc-testclient
        image: $REPO_URL/$image
        command: ["/bin/sh", "-c", "ipalloc_testclient -concurrent_num=$CONCURRENT_NUM -method=$TEST_METHOD -alsologtostderr && sleep infinity"]
      volumes:
      - configMap:
          defaultMode: 420
          name: test-ipallocator-server-config
        name: ipallocator-server
EOF
}

function build_image {
docker build -t $REPO_URL/$image --build-arg src_tar=src-$TAG.tgz .
docker login $DOCKER_URL -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker push $REPO_URL/$image
}

tar_src
build_image
k8s_build
