# Travel Angels web hosting

This is an independent Terraform root for the Flutter web app. It uses its own
local `terraform.tfstate`, separate from the parent directory's infrastructure
and secret resources. Back up this state securely after applying. State and
saved plans are ignored by Git; commit the provider lock file.

The configuration creates a private, encrypted, versioned S3 bucket and an HTTPS
CloudFront distribution. CloudFront signs origin requests with OAC; only that
distribution may read the bucket. The existing distribution serves
`travel-angels.org` and `www.travel-angels.org`, using a DNS-validated ACM
certificate in `us-east-1`. The `website_url` output uses the apex domain, while
`cloudfront_url` retains the original CloudFront address as a fallback.

The domain's existing public Route53 hosted zone is read as data. This root adds
certificate-validation CNAMEs and A/AAAA web aliases; it does not create or replace
the zone, alter registrar nameservers, or change unrelated records. Existing
records are not overwritten automatically. DNS delegation must already point to
the hosted zone for certificate validation to complete. Backend API and secret
resources remain outside this root.

From the repository root:

```sh
terraform -chdir=terraform/travel-angels-web init
terraform -chdir=terraform/travel-angels-web validate
terraform -chdir=terraform/travel-angels-web plan -out=hosting.tfplan
terraform -chdir=terraform/travel-angels-web apply hosting.tfplan
terraform -chdir=terraform/travel-angels-web output
```

The AWS provider is pinned to the parent root's currently installed version,
`6.25.0`, and restricted to account `872408892710`. CloudFront provisioning can
take several minutes. Apply returns while CloudFront propagates so assets can be
uploaded immediately; confirm the distribution reaches `Deployed` and verify the
`website_url` output before treating the deployment as complete. Initial custom
domain setup also waits for ACM DNS validation. Verify both apex and `www` HTTPS
addresses after the distribution finishes deploying; the existing bucket,
distribution ID, and original CloudFront hostname remain unchanged.

Upload the contents of `dart/travel_angels/build/web` to the output `bucket_name`,
then invalidate the output `distribution_id`. Upload assets before app
entrypoints. Set `Cache-Control: no-cache, max-age=0, must-revalidate` on
`index.html`, `flutter_bootstrap.js`, `flutter.js`, `main.dart.js`, and any
service-worker/version manifest so upgrades do not retain an old application.
The cache policy has a zero minimum TTL and honors these origin directives;
other assets default to five minutes and are capped at one day.

The app uses hash routing, so deep links retain the root document path. Missing
files keep their real error status; there is no catch-all 403/404-to-HTML rewrite.
The distribution adds no-referrer, noindex, nosniff, and HSTS response headers.
Build the app with its verified live API URL before uploading; hosting does not
change the backend deployment or run database migrations.
