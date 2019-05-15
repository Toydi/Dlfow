## How to deploy

```bash
$ cd deploy
```
```bash
$ helm install ./dlflow --name <name> --namespace <namespace> -f ./dlflow/values.yaml
```

## Cleanup

Using `helm` to remove dlflow from your cluster with commands as follows
```bash
$ helm del <name> --purge
```
## Run an example
```bash
$ cd deploy/example
$ kubectl create -f dfjob1.yaml -n <namespace>
```
