docker buildx build --platform "linux/arm64,linux/amd64" --tag docker.kannonfoundry.dev/whutbot3 --push .
kubectl rollout restart deployment --namespace bot whutbot