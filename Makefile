all: docker k8s-delete k8s-apply

docker:
	docker build -t sinth-chaos-poc .
	docker tag sinth-chaos-poc localhost:5000/sinth-chaos-poc:latest
	kind load docker-image localhost:5000/sinth-chaos-poc:latest

k8s-apply:
	kubectl apply -f ./k8s/pod.yaml

k8s-delete:
	kubectl delete -f ./k8s/pod.yaml
