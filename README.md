```bash
helm install webhook ./charts/webook
``` 

```bash
kubectl apply -f ./showcase-resources
```

```bash
kubectl apply -f ./ok-pod.yaml --as=system:serviceaccount:default:test
``` 

```bash
kubectl apply -f ./no-label-pod.yaml --as=system:serviceaccount:default:test
``` 

```bash
kubectl create deployment nginx  --image nginx --as system:serviceaccount:default:test 
```