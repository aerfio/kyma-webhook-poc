## TODO:
- proper logging - done
- securityContexts etc - done
- istio-sidecarem ? - doesn't work, don't know why, all other webhooks we have do not have sidecar
- production/eval profiles ? - still not done
- readiness/liveness probes - done
- prometheus metrics - provided out of the box, done
- exempt serviceaccounts from denied namespaces (from iteration review)
- serviceacounts -> handled already
- groups/users -> see `gke-user.png`

- kubectl exec -> denied, special verb for that action is CONNECT (see webhooks[].rules.operations)
  - available verbs are:
    - CONNECT
    - CREATE
    - UPDATE
    - DELETE
- kubectl logs -> allowed
- kubectl get,list,watch -> allowed (- do not block reading (from iteration review))
  
- Which kubeconfig is used to install Kyma - Still don't know


```bash
kubectl -n kyma-system run busybox --image busybox --as=system:serviceaccount:default:test-deny -- sh -c "echo something; sleep 10000"
``` 