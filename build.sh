export DAPR_REGISTRY=registry.cn-hangzhou.aliyuncs.com/jkhe
export DAPR_TAG=hjk-1.00

make build-linux
make docker-build
make docker-push

helm uninstall dapr -n dapr-system
make docker-deploy-k8s