brew install kubectl
brew upgrade kubectl
brew link --overwrite kubernetes-cli
brew install helm
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add elastic https://helm.elastic.co

brew unlink minikube
brew install minikube
brew link minikube

minikube delete
minikube config set cpus 4
minikube config set memory 4096
minikube start
minikube addons enable default-storageclass
minikube addons enable storage-provisioner
minikube addons enable ingress

