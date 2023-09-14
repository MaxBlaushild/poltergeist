.PHONY: dev
dev:
	go run ./cmd/server/main.go --config-name local

PHONY: ecr-push
ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/scorekeeper:latest .

	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/scorekeeper:latest
