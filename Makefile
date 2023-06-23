PHONY: deploy-all
deploy-all:
	aws ecs update-service --cluster poltergeist --service poltergeist_core --force-new-deployment
