### Helm Chart

create a storageclass first for dynamic provisioning:
`kubectl create -f storageClass.yaml`

then install helm chart:
`helm install --namespace cassandra -n cass -f config.yaml incubator/cassandra`
