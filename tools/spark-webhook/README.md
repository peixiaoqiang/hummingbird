# Kubernetes Mutating Admission Webhook for Spark
## Prerequisites

Kubernetes 1.9.0 or above with the `admissionregistration.k8s.io/v1beta1` API enabled. Verify that by the following command:

```
kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:

```
admissionregistration.k8s.io/v1beta1
```
In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.
## Build

```
$ cd tools/spark-webhook
# build spark-webhook image
$ ./build_image.sh <image_name>
```
## Deploy

```
# create secret
$ ./webhook-create-signed-cert.sh --service <servie_name> --secret <secret_name> --namespace <namespace_name> --cluster <cluster_suffix>
# deploy to Kubernetes 
$ cat mutatingWebhookConfiguration.yaml | ./webhook-patch-ca-bundle.sh > mutatingwebhook-ca-bundle.yaml
$ kubectl apply -f mutatingwebhook-ca-bundle.yaml 
$ kubectl apply -f webhook.yaml
```