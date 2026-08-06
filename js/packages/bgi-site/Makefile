dev:
	npm run dev

# distribution-id filled in once trays.forteus.tech's CloudFront
# distribution is provisioned (mirrors js/packages/reef-site/Makefile's
# same pattern for reef.forteus.tech — that infra isn't in terraform
# either, see go/bgi-site/PLATFORM_FINDINGS.md).
deploy:
	npm run build && aws s3 sync dist s3://trays.forteus.tech
	aws cloudfront create-invalidation --distribution-id E3JTS0VS6ITN70 --paths "/*"
