PHONY: deploy-all
deploy-all:
	aws ecs update-service --cluster poltergeist --service poltergeist_core --force-new-deployment

deps:
	docker-compose -f deps.docker-compose.yml up -d

PHONY: trivai/ecr-push
trivai/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/trivai/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/trivai:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/trivai:latest

PHONY: texter/ecr-push
texter/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/texter/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/texter:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/texter:latest

PHONY: scorekeeper/ecr-push
scorekeeper/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/scorekeeper/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/scorekeeper:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/scorekeeper:latest

PHONY: fount-of-erebos/ecr-push
fount-of-erebos/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/fount-of-erebos/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/fount-of-erebos:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/fount-of-erebos:latest

PHONY: crystal-crisis-api/ecr-push
crystal-crisis-api/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/crystal-crisis-api/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/crystal-crisis-api:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/crystal-crisis-api:latest

PHONY: core/ecr-push
core/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/core/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/core:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/core:latest

PHONY: authenticator/ecr-push
authenticator/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	docker build -f ./deploy/services/authenticator/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/authenticator:latest .
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/authenticator:latest