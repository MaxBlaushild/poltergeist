dev:
	npm run dev

deploy:
	npm run build && aws s3 sync dist s3://reef.forteus.tech
	aws cloudfront create-invalidation --distribution-id E2LKLM2OK4VTQ5 --paths "/*"
