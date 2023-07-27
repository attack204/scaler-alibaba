#!/bin/zshrc

docker buildx build --platform linux/amd64 -t registry.cn-beijing.aliyuncs.com/gaojiliu-competition/scaler:$IMAGE_TAG . --push

kubectl delete job serverless-simulation

envsubst < ../manifest/serverless-simulaion.yaml | kubectl apply -f -

