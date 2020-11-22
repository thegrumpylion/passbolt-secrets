#!/bin/bash

# MINIKUBE

minikube start --extra-config=apiserver.service-node-port-range=1-65535

# DB

kubectl create -f example/db/namespace.yaml

kubectl create secret generic -n db mariadb --from-literal=password=mypassword

kubectl create configmap -n db mariadb --from-literal=dbname=passbolt

kubectl create -f example/db/pod.yaml

kubectl create -f example/db/service.yaml

# PASSBOLT

kubectl create -f example/passbolt/namespace.yaml

kubectl create secret generic -n passbolt db --from-literal username=root --from-literal password=mypassword

kubectl create configmap -n passbolt config \
    --from-literal=dbhost=mariadb.db.svc.cluster.local \
    --from-literal=dbname=passbolt \
    --from-literal=base-url=https://pass.bolt

kubectl create -f example/passbolt/pod.yaml

kubectl create -f example/passbolt/service.yaml

kubectl exec -it -n passbolt passbolt -- su -m -c "/var/www/passbolt/bin/cake passbolt register_user -u john.doe@example.com -f john -l doe -r admin" -s /bin/sh www-data

# PASSBOLT-SECRETS

kubectl create -f artifacts/namespace.yaml

kubectl create -f artifacts/cluster-role.yaml

kubectl create -f artifacts/service-account.yaml

kubectl create -f artifacts/cluster-role-binding.yaml

kubectl create -f artifacts/service-token.yaml

kubectl create -f artifacts/custom-resource-definition.yaml

kubectl create configmap -n passbolt-secrets passbolt-server \
    --from-literal fingerprint=8309EA52960DABD7B4BECEB93316619DAC9D2C81 \
    --from-literal url=https://passbolt.passbolt.svc.cluster.local

kubectl create secret generic -n passbolt-secrets passbolt-server \
    --from-file key=~/downloads/passbolt_private.txt \
    --from-literal password=asdf1234

kubectl create -f artifacts/controller-pod.yaml

# APP

kubectl create -f example/app/passbolt-secret.yaml

kubectl create -f example/app/pod.yaml

kubectl logs test