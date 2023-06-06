# Tutorial

### Onboard new K8S cluster, with Collie.eng.vmware.com

Get bootstrap command
```
curl -skH "Authorization: bc3dcc07f8263dda643d1703cd5254f2" https://collie.eng.vmware.com/collie/api/v1/onboarding/bootstrap-cmd
```

It will generate a command line like below. Make sure your current kubectl points to the taregt context, then execute the command.

```
curl -skH "Authorization: Token bc3dcc07f8263dda643d1703cd5254f2" "https://collie.eng.vmware.com/collie/api/v1/onboarding/agent.yaml?provider=AKS" | kubectl apply -f -
```

Navigate to dashboard
```
https://collie.eng.vmware.com/d/qIbLYbT4z/k8s-compliance-report?orgId=1
```

### Start local dev server (not using collie.eng.vmware.com)
```
cd collie/api-server
cp collie/helm-charts/.env.example .env
vi .env
source .env
go run .
```

### Deploy local cluster
```
cd collie/helm-charts
cp .env.example .env
vi .env
source .env
envsubst < api-server.yaml | kubectl apply -f -
```

####Onboard agent

```
curl http://localhost:8080/api/v1/onboarding/bootstrap-cmd
```
