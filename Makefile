docker-build-image-prod:
	docker build -t k3sdemo . && \
	docker tag k3sdemo localhost:5000/k3sdemo

docker-push-image-prod:
	docker push localhost:5000/k3sdemo && \
	curl 127.0.0.1:5000/v2/_catalog
	

k3s-create-namespace: 
	kubectl create namespace k3sdemo

k3s-apply-yaml:
	kubectl apply -f ./deploy/development.yaml && \
	kubectl apply -f ./deploy/service.yaml && \
	kubectl apply -f ./deploy/ingress.yaml

k3s-replace-yaml:
	kubectl replace -f ./deploy/development.yaml && \
	kubectl replace -f ./deploy/service.yaml && \
	kubectl replace -f ./deploy/ingress.yaml

k3s-restart-deployment:
	kubectl rollout restart deployment k3sdemo-deployment -n k3sdemo
