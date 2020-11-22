# passbolt-secrets

Generate Kubernetes secrets from Passbolt.

## State

alpha

## PassboltSecret Resource

```yaml
apiVersion: passboltsecrestscontroller.greatlion.tech/v1alpha1
kind: PassboltSecret
metadata:
  name: example-secret
spec:
  source:
    name: my_very_secret
```

This will look for passbolt resource `my_very_secret` and create the following Kubernetes secret in default namespace.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-secret
type: Opaque
data:
  secret: <secret data>
```

### Get by ID

```yaml
apiVersion: passboltsecrestscontroller.greatlion.tech/v1alpha1
kind: PassboltSecret
metadata:
  id: 25d52ee9-efcd-443d-bee9-aa167d3b3da2
spec:
  source:
    name: my_very_secret
```

### Full Options

You can customize the resulting secret by providing key names for each passbolt resource filed and specify the name.

```yaml
apiVersion: passboltsecrestscontroller.greatlion.tech/v1alpha1
kind: PassboltSecret
metadata:
  name: example-secret
spec:
  source:
    name: my_very_secret
  name: my-secret
  secretKey: password
  usernameKey: username
  urlKey: target
```

This will result to

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
data:
  password: <secret data>
  username: <username data>
  target: <url data>
```

## Example

For this example we will use [minikube](https://minikube.sigs.k8s.io/docs/). Also, `kubectl` needs to be installed.

Default Minikube cluster CPU is `2` & memory is `2048`

```sh
minikube config set cpus 8
minikube config set memory 8192
```

Start `minikube` with all the ports available so we can bind 443 to passbolt service.

```sh
minikube start --extra-config=apiserver.service-node-port-range=1-65535
```


Add a line in `/etc/hosts` that point to minikube's ip address to access passbolt web UI later.

```sh
echo "$(minikube ip) pass.bolt" | sudo tee -a /etc/hosts
```



### Deploy mariadb

```sh
kubectl create -f example/db/namespace.yaml

kubectl create secret generic -n db mariadb --from-literal=password=mypassword

kubectl create configmap -n db mariadb --from-literal=dbname=passbolt

kubectl create -f example/db/pod.yaml

kubectl create -f example/db/service.yaml
```

### Deploy passbolt

```sh
kubectl create -f example/passbolt/namespace.yaml

kubectl create secret generic -n passbolt db --from-literal username=root --from-literal password=mypassword

kubectl create configmap -n passbolt config \
    --from-literal=dbhost=mariadb.db.svc.cluster.local \
    --from-literal=dbname=passbolt \
    --from-literal=base-url=https://pass.bolt

kubectl create -f example/passbolt/pod.yaml

kubectl create -f example/passbolt/service.yaml
```

After passbolt is up, initiate configuration with the following:

```sh
kubectl exec -it -n passbolt passbolt -- su -m -c "/var/www/passbolt/bin/cake passbolt register_user -u john.doe@example.com -f john -l doe -r admin" -s /bin/sh www-data
```

If all went well you should get

```
User saved successfully.
To start registration follow the link provided in your mailbox or here: 
https://pass.bolt/setup/install/...
```

Follow the link to finish setup. Save the `Server key`, generate a new key, set a passphrase & download your private key.

After the setup is finished, assuming the passbolt browser plugin is installed, login and create a secret with `username` & `url` set. Name your secret `my_very_secret` to continue copy-paste.

### Deploy passbolt-secret-controller

Build the controller and the image using Docker with minikube environment so the image will be available to our local cluster.

```sh
eval `minikube docker-env`
make image_build
```

Create Kubernetes resources for passbolt-secret-controller.

```sh
kubectl create -f artifacts/namespace.yaml

kubectl create -f artifacts/cluster-role.yaml

kubectl create -f artifacts/service-account.yaml

kubectl create -f artifacts/cluster-role-binding.yaml

kubectl create -f artifacts/service-token.yaml

kubectl create -f artifacts/custom-resource-definition.yaml
```

Use the `Server key` from earlier to generate the config map

```sh
kubectl create configmap -n passbolt-secrets passbolt-server \
    --from-literal fingerprint=<the_server_key> \
    --from-literal url=https://passbolt.passbolt.svc.cluster.local
```

Use your downloaded private key file and passphrase to create the secret.

```sh
kubectl create secret generic -n passbolt-secrets passbolt-server \
    --from-file key=<passbolt_private.txt> \
    --from-literal password=<passphrase>
```

Finally, create the controller pod.

```sh
kubectl create -f artifacts/controller-pod.yaml
```

You are ready now to create some PassboltSecrets!

### Secret & app example

```sh
kubectl create -f example/app/passbolt-secret.yaml

kubectl create -f example/app/pod.yaml
```

Check the logs of the pod to see the values of the passbolt secret you created.

```sh
kubectl logs test
```
