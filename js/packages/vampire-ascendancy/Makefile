dev:
	npm run dev

deploy:
	yarn build && aws s3 sync dist s3://vampire-ascendancy
	aws cloudfront create-invalidation --distribution-id E35QTR6M5YEJDA --paths "/*"
