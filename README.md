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

### Problemy
- Każdy request przechodzi przez ten webhook, moga byc problemy wydajnosciowe
- single point of failure -> jezeli webhook nie bedzie dzialac, to nic na clustrze nie moze byc stworzone (z powodu "*")
- cluster-admin moze na milion sposobow uniknac dzialania webhooka 
   - zmienic konfiguracje webhooka, a zmiany validationwebhookconfiguration nie przechodza przez webhook
   - zmienic jakos konfiguracje deploymentu webhooka
   - stworzyc nowy service account z inna nazwa a ze wszystkimi uprawnieniami
- https://github.com/kubernetes/kubernetes/issues/85963#issuecomment-708403412    
- However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks
  from putting the cluster in a state which cannot be recovered from without completely
  disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called
  on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
