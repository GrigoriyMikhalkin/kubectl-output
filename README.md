# kubectl output

`kubectl-output` is a plugin for `kubectl` that allows users to set custom output format for specific resources/namespaces.
Custom output format is based on [custom-columns](https://kubernetes.io/docs/reference/kubectl/#custom-columns).

Example of how user can set custom output format for `Pod` resources in `test` namespace:
```shell
kubectl output set pods -n test -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,NAMESPACE:.metadata.namespace
```

The config is stored in `~/.kube-output/resource_tmpl_map.yaml` file. Which is later used to set custom output format for related requests made with `kubectl output get` command. For example:
```shell
kubectl output get pods -n test
```

### Examples:
```shell
alias ko="kubectl output"

# create default template
ko set pod --name=test -c=NAME:.metadata.name

# create template for kube-system namespace from file
cat pod.tmpl
# NAME          RSRC
# metadata.name metadata.resourceVersion

ko set pod --name=test-file --namespace=kube-system -c=./pod.tmpl

# list all templates for pod
ko template pod -a
# test
# test-file

# get pods with default template
ko get pods -A
# NAME
# pod1

# override default template
ko get pods -A -o test-file
# NAME                 RSRC
# pod1                 123
```

