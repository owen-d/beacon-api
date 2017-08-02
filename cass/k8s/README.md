### Helm Chart

create a storageclass first for dynamic provisioning:
`kubectl create -f storageClass.yaml`

then install helm chart:
`helm install -f config.yaml incubator/cassandra`
