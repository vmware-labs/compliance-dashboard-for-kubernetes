# compliance-dashboard-for-kubernetes

[![License](https://img.shields.io/badge/License-Apache%202.0-blue)](https://github.com/vmware-labs/compliance-dashboard-for-kubernetes/blob/main/LICENSE)

## Overview
A K8s compliance checker aggregator with a dashboard and analyzer of K8s compliance, as well as 3rd party scanners integration.

## How it works
The Compliance Dashboard for Kubernetes consists of:
* An agent to be deployed to the target kubernetes, and report details.
* A web portal based on Grafana to visualize the findings.
* An Elasticsearch backend for persist.
* An api server to connect all the above together.

![Dashboard](doc/images/screenshot-dashboard.png?raw=true)

### Plugins On Roadmap
* [kube-bench](https://github.com/aquasecurity/kube-bench)
* [kube-hunter](https://github.com/aquasecurity/kube-hunter)
* [kube-score](https://github.com/zegl/kube-score)
* [kube-linter](https://github.com/stackrox/kube-linter)
* Collie Analysis


## Try it out

### Prerequisites
On Mac:

* Approach 1 - Automated approach, try the [preparation script](deployment/helm-charts/prep.sh)

* Approach 2 - Manual installation:
  - Install homebrew: https://brew.sh/
  - Install/upgrade kubectl: https://formulae.brew.sh/formula/kubernetes-cli
    ``` 
    brew upgrade kubectl
    brew link --overwrite kubernetes-cli
    ```
  - Install/update minikube: https://minikube.sigs.k8s.io/docs/start/
    ```
    brew unlink minikube
    brew install minikube
    brew link minikube
    ```
  - Config and start minikube
    ```
    minikube config set cpus 4
    minikube config set memory 4096
    minikube start
    minikube addons enable default-storageclass
    minikube addons enable storage-provisioner
    minikube addons enable ingress
    ```
  - Install helm chart: https://helm.sh/docs/intro/quickstart/
    ```
    brew install helm
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add elastic https://helm.elastic.co
    helm repo update
    ```
### Build & Run

To run prebuilt images in local environment:

1. Identify local PC public IP, e.g. via ifconfig.
2. Add a DNS record "collie-dev.org" to that IP in /etc/hosts file
3. Run the deployment script, which deploys all components and forward ports to local host properly.
```
cd deployment/helm-charts

./deploy-all.sh
```
4. Open browser:
    ```
    http://collie-dev.org:8080/collie/portal/login
    ``` 
    ![Login](doc/images/screenshot-login.png?raw=true)
5. Copy agent installation script from the UI, and execute the script to install the againt. The script is a kubectl command to deploy the agent. You may run on any k8s that can connects to your pc.
   ![Agent Pairing](doc/images/screenshot-pairing.png?raw=true)
6. After the agent starts and paired, dashboard button is enabled on the UI page.
    ![Agent Paired](doc/images/screenshot-paired.png?raw=true)
7. Click the button to open the dashboard
    ![Dashboard](doc/images/screenshot-dashboard.png?raw=true)

    Note: this is a known issue with the default bootstrap, and by default for the first time you will get "Page not found" and "Unauthorized" notification. To workaround it, see the known issue section below.

### Known issue
To workaround the first-time "Page not found", a one-time operation is needed.
1. In the opened dashboard, click the right top "Sign in", and sign in using admin/admin
2. Click left top "Toggle Menu" -> "Administration" -> "Datasources" -> "es-collie-k8s-elastic" -> "Save & test"
3. Open [dashboard](http://collie-dev.org:3000/d/qIbLYbT4z/k8s-compliance-report?orgId=1) again

### Cleanup local dev environment
The following script will delete all the deployed k8s resources.
```
./delete-all.sh
```
or totally destroy the environment:
```
minikube delete
```

## Contributing

The compliance-dashboard-for-kubernetes project team welcomes contributions from the community. Before you start working with compliance-dashboard-for-kubernetes, please read and sign our [Contributor License Agreement CLA](CONTRIBUTING_CLA.md). If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will prompt you to do so when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ]([https://cla.vmware.com/faq](https://cla.vmware.com/faq)).

For more detailed information, refer to [CONTRIBUTING_CLA.md](CONTRIBUTING_CLA.md).

## License
Apache-2.0
