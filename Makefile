PHONY: go/mod
go/mod:
	@for dir in $$(find go -name 'go.mod' -exec dirname {} \; | sort); do \
		echo "go mod tidy: $$dir"; \
		(cd $$dir && go mod tidy) || exit 1; \
	done
	@echo "go work sync"
	@go work sync

.PHONY: deploy-all
deploy-all:
	aws ecs update-service --cluster poltergeist --service poltergeist_core --force-new-deployment

deps:
	docker-compose -f deps.docker-compose.yml up -d

PHONY: trivai/build
trivai/build:
	docker build -f ./deploy/services/trivai/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/trivai:latest .

PHONY: trivai/ecr-push
trivai/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make trivai/build
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

PHONY: crystal-crisis-api/build
crystal-crisis-api/build:
	docker build -f ./deploy/services/crystal-crisis-api/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/crystal-crisis:latest .

PHONY: crystal-crisis-api/ecr-push
crystal-crisis-api/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make crystal-crisis-api/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/crystal-crisis:latest

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

PHONY: admin/build
admin/build:
	docker build -f ./deploy/services/admin/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/admin:latest .

PHONY: admin/ecr-push
admin/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make admin/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/admin:latest

PHONY: billing/build
billing/build:
	docker build -f ./deploy/services/billing/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/billing:latest .

PHONY: billing/ecr-push
billing/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make billing/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/billing:latest

PHONY: sonar/build
sonar/build:
	docker build -f ./deploy/services/sonar/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/sonar:latest .

PHONY: sonar/ecr-push
sonar/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make sonar/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/sonar:latest

PHONY: job-runner/build
job-runner/build:
	docker build -f ./deploy/services/job-runner/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/job-runner:latest .

PHONY: job-runner/ecr-push
job-runner/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make job-runner/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/job-runner:latest

PHONY: travel-angels/build
travel-angels/build:
	docker build -f ./deploy/services/travel-angels/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/travel-angels:latest .

PHONY: travel-angels/ecr-push
travel-angels/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make travel-angels/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/travel-angels:latest

PHONY: final-fete/build
final-fete/build:
	docker build -f ./deploy/services/final-fete/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/final-fete:latest .

PHONY: final-fete/ecr-push
final-fete/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make final-fete/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/final-fete:latest

PHONY: ethereum-transactor/build
ethereum-transactor/build:
	docker build -f ./deploy/services/ethereum-transactor/Dockerfile --platform x86_64 -t 872408892710.dkr.ecr.us-east-1.amazonaws.com/ethereum-transactor:latest .

PHONY: ethereum-transactor/ecr-push
ethereum-transactor/ecr-push:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 872408892710.dkr.ecr.us-east-1.amazonaws.com
	# Build the Docker image
	make ethereum-transactor/build
	# Push the Docker image to ECR
	docker push 872408892710.dkr.ecr.us-east-1.amazonaws.com/ethereum-transactor:latest