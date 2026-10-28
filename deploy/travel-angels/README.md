# Travel Angels web deployment

The canonical web origin is [Travel Angels](https://travel-angels.org),
with `https://www.travel-angels.org` as an alternate address and the API at
`https://api.unclaimedstreets.com`. Both custom domains were verified over HTTPS
on October 28, 2026, with an issued ACM certificate and deployed CloudFront
configuration. The existing [CloudFront address](https://d3hd0wcyslgumm.cloudfront.net)
remains available.

The custom-domain web release was also verified on October 28, 2026: both hosts
serve HTTPS, HTTP redirects to HTTPS, public A/AAAA records resolve, and the
served bundle checksum matches the local build. Browser checks passed for login
and the guest shared-trip access screen. The app's release metadata identifies
`https://travel-angels.org` as its web origin.
The API cutover completed in `sonar_core:52` with one healthy task and all seven
previous application image digests preserved; generated links use the same
canonical origin.

| Resource | Value |
| --- | --- |
| Canonical web origin | `https://travel-angels.org` |
| Alternate web origin | `https://www.travel-angels.org` |
| CloudFront fallback | `https://d3hd0wcyslgumm.cloudfront.net` |
| API origin | `https://api.unclaimedstreets.com` |
| CloudFront distribution | `E2EQ9RE874X8K6` |
| Private S3 bucket | `travel-angels-web-872408892710` |
| Hosting Terraform root | `terraform/travel-angels-web` |

The initial web release was published and verified at the CloudFront address on
October 28, 2026.
CloudFront deployment and invalidation completed, the served JavaScript matched
the local release checksum, and the fonts and CanvasKit loaded successfully.
Fresh desktop and phone browser sessions rendered the styled login screen with
no JavaScript errors. An invalid shared-trip link reached the guest access
screen without requiring login. No verification texts or invitations were sent
during these checks.

## Publish a web release

Run from the repository root with Flutter, Terraform, Python 3, Git, and a current
AWS CLI available, using AWS credentials for the hosting account:

```sh
make travel-angels/web-deploy
```

If Flutter is not on `PATH`, supply its executable explicitly:

```sh
FLUTTER_BIN=/Applications/flutter/bin/flutter make travel-angels/web-deploy
```

`deploy_web.sh` reads the bucket, distribution, and `website_url` output from the
isolated hosting state and checks their relationship before uploading. After the
domain configuration is applied, `website_url` is `https://travel-angels.org`;
the separate `cloudfront_url` output retains the fallback address. It builds the
current working tree, sets the live API and web URL build defines, uploads assets
before entry files, and invalidates CloudFront. `TRAVEL_ANGELS_API_URL` can override
the default API origin for an intentional environment change. `TERRAFORM_BIN` and
the normal `AWS_PROFILE`/`AWS_REGION` settings are also supported.

The script keeps older objects for existing browser tabs, requires cache
revalidation, and disables Flutter offline caching. `release.json` records the
source revision, build time, working-tree status, and bundle checksum. It contains
no credentials. Publishing web files does not apply Terraform, deploy a backend
image, or run migrations.

After an upload, verify CloudFront and the release before announcing completion:

```sh
aws cloudfront wait distribution-deployed --id E2EQ9RE874X8K6
aws cloudfront get-distribution --id E2EQ9RE874X8K6 \
  --query Distribution.Status --output text
travel_angels_web_url="$(terraform -chdir=terraform/travel-angels-web output -raw website_url)"
curl --fail --silent --show-error "$travel_angels_web_url/release.json"
```

The deploy command prints its invalidation ID. Check that invalidation completes,
then open the app in a fresh browser session and verify the expected UI, login,
and an authorized shared trip. Use the release metadata to distinguish a fresh
build from an old browser tab.

## Manage the isolated hosting state

Ordinary UI releases need only the deploy command above. For hosting changes, use
this independent root and review its saved plan:

```sh
terraform -chdir=terraform/travel-angels-web init
terraform -chdir=terraform/travel-angels-web validate
terraform -chdir=terraform/travel-angels-web plan -out=hosting.tfplan
terraform -chdir=terraform/travel-angels-web apply hosting.tfplan
terraform -chdir=terraform/travel-angels-web output
```

The local state is `terraform/travel-angels-web/terraform.tfstate`. Keep a secure
backup after infrastructure changes; it is intentionally ignored by Git along
with saved plans and `.terraform`. For example, replace the destination below with
an existing private backup directory:

```sh
umask 077
terraform -chdir=terraform/travel-angels-web state pull > /secure/backup/travel-angels-web.tfstate
```

Do not initialize an empty replacement state for resources that already exist;
restore the hosting state when working from another machine. The provider lock
file belongs in Git. The parent `terraform/` root manages separate application
infrastructure and secrets; do not apply it to deploy web files.

## Connect and verify the custom domain

The isolated hosting root connects the existing `travel-angels.org` hosted zone
to the same CloudFront distribution and S3 bucket. It adds an ACM certificate in
`us-east-1`, its DNS-validation CNAMEs, and A/AAAA aliases for both the apex and
`www`. Existing hosted-zone NS/SOA records and registrar name servers are
preserved; registration and delegation must finish before public DNS validation
can succeed. This does not create a second app or move the API.

Keep the ACM validation CNAMEs after the certificate is issued. ACM needs those
records for automatic certificate renewal. Preserve unrelated DNS records and
the existing hosted-zone delegation when making later changes.

Before switching generated links, verify that the certificate is issued, the
CloudFront update is deployed, and both `https://travel-angels.org` and
`https://www.travel-angels.org` resolve and load successfully over HTTPS. Verify
the CloudFront fallback as well. `www` serves the same app; generated links use
the apex origin. Publish a web release after the hosting outputs change so its
build defines use the canonical URL, then perform the targeted API update below.
The original CloudFront address remains available for existing links and
diagnostics.

## Set the API's public share-link origin

The API must also use `TRAVEL_ANGELS_WEB_URL` when creating invitation, calendar,
trip, and preference links. The initial deployment configured
`https://d3hd0wcyslgumm.cloudfront.net` in `sonar_core:51`, preserving all seven
existing application image digests. That is historical release information;
the custom-domain runtime change is a separate rollout. Once the domain's DNS
and HTTPS checks pass, preview the targeted ECS update:

```sh
python3 deploy/travel-angels/configure_web_url.py \
  --web-url https://travel-angels.org
```

Review the service, rollback task definition, image digests, and proposed URL.
Then apply that same URL-only change:

```sh
python3 deploy/travel-angels/configure_web_url.py \
  --web-url https://travel-angels.org --apply
```

This requires a current AWS CLI with the current ECS task-definition model. The
script rejects an outdated model; update the installed CLI or pass `--aws-cli`
with a durable executable path. It uses the usual AWS credential/profile settings.

The script targets `poltergeist` / `sonar_core` in `us-east-1`, preserves existing
task configuration, pins every application container to its current running image
digest, and changes only the core container's web URL environment setting. It
refuses to overwrite a deployment that changes while the update is being prepared.
It does not deploy newly pushed mutable image tags. After `--apply`, monitor the
ECS rollout and API health; the script returns after starting the deployment.

`terraform/travel_angels_web_url.tf` records the desired canonical URL as the parent root's
`travel_angels_web_url` default, and `terraform/ecs.tf` supplies the variable to
core, so a later reviewed infrastructure deployment retains the setting. That
configuration does not require an unrelated parent-root apply for a web release.
