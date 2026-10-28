#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Build and deploy Travel Angels web to its dedicated Terraform-managed hosting.

Usage: deploy/travel-angels/deploy_web.sh [--help]
       make travel-angels/web-deploy

Requires applied terraform/travel-angels-web outputs: bucket_name,
distribution_id, website_url. This script never applies Terraform.

Environment:
  TRAVEL_ANGELS_API_URL  HTTPS API origin (default https://api.unclaimedstreets.com)
  FLUTTER_BIN           Flutter executable (default flutter)
  TERRAFORM_BIN         Terraform executable (default terraform)
  AWS_PROFILE/AWS_REGION  Standard AWS CLI configuration

Validates the target before building or uploading. Uploads assets before entry
files, keeps prior objects, and invalidates CloudFront. All uploaded files require
cache revalidation; offline service-worker caching is disabled. release.json
records the source revision, build time, and bundle checksum, never credentials.
USAGE
}

if [[ ${1:-} == --help || ${1:-} == -h ]]; then
  usage
  exit 0
fi
if [[ $# -ne 0 ]]; then
  usage >&2
  exit 2
fi

fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_root=$(cd -- "$script_dir/../.." && pwd)
terraform_dir="$repo_root/terraform/travel-angels-web"
app_dir="$repo_root/dart/travel_angels"
flutter_bin=${FLUTTER_BIN:-flutter}
terraform_bin=${TERRAFORM_BIN:-terraform}
api_url=${TRAVEL_ANGELS_API_URL:-https://api.unclaimedstreets.com}
export AWS_PAGER=""
for executable in "$flutter_bin" "$terraform_bin" aws python3 git; do
  command -v "$executable" >/dev/null 2>&1 || fail "Missing executable: $executable"
done
[[ -d $terraform_dir ]] || fail "Missing hosting configuration: $terraform_dir"
release_tmp=$(mktemp -d "${TMPDIR:-/tmp}/travel-angels-web.XXXXXXXX")
trap 'rm -rf -- "$release_tmp"' EXIT

"$terraform_bin" -chdir="$terraform_dir" output -json > "$release_tmp/outputs.json"
python3 - "$release_tmp/outputs.json" "$api_url" > "$release_tmp/target" <<'PY'
import json, re, sys
from urllib.parse import urlsplit

outputs = json.load(open(sys.argv[1]))
def value(name):
    result = outputs.get(name, {}).get('value')
    if not isinstance(result, str) or not result or '\n' in result:
        raise SystemExit(f'Missing or invalid Terraform output: {name}')
    return result
def origin(url):
    parsed = urlsplit(url)
    if (parsed.scheme != 'https' or not parsed.hostname or parsed.username
            or parsed.password or parsed.query or parsed.fragment
            or parsed.path not in ('', '/') or parsed.port not in (None, 443)):
        raise SystemExit('Web and API URLs must be HTTPS origins without credentials or paths')
    return url.rstrip('/')
bucket = value('bucket_name')
distribution = value('distribution_id')
if not re.fullmatch(r'[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]', bucket):
    raise SystemExit('Invalid S3 bucket name')
if not re.fullmatch(r'[A-Z0-9]{8,32}', distribution):
    raise SystemExit('Invalid CloudFront distribution ID')
print(bucket)
print(distribution)
print(origin(value('website_url')))
print(origin(sys.argv[2]))
PY
{
  IFS= read -r bucket
  IFS= read -r distribution
  IFS= read -r web_url
  IFS= read -r api_url
} < "$release_tmp/target"

# Read-only checks prevent uploading to an unrelated bucket or distribution.
aws sts get-caller-identity --output json > "$release_tmp/identity.json"
aws s3api head-bucket --bucket "$bucket"
aws cloudfront get-distribution --id "$distribution" --output json > "$release_tmp/distribution.json"
python3 - "$release_tmp/distribution.json" "$release_tmp/identity.json" "$bucket" "$web_url" <<'PY'
import json, re, sys
from urllib.parse import urlsplit

dist = json.load(open(sys.argv[1]))['Distribution']
account = json.load(open(sys.argv[2]))['Account']
bucket, web_url = sys.argv[3:]
config = dist['DistributionConfig']
allowed_hosts = [dist['DomainName'], *config.get('Aliases', {}).get('Items', [])]
if urlsplit(web_url).hostname not in allowed_hosts or not config.get('Enabled'):
    raise SystemExit('Website URL must belong to this enabled CloudFront distribution')
if dist['ARN'].split(':')[4] != account:
    raise SystemExit('AWS identity does not match the distribution account')
origin_id = config['DefaultCacheBehavior']['TargetOriginId']
origins = [item for item in config['Origins']['Items'] if item['Id'] == origin_id]
pattern = re.escape(bucket) + r'\.s3(?:[.-][a-z0-9-]+)?\.amazonaws\.com'
if (len(origins) != 1 or origins[0].get('OriginPath')
        or not re.fullmatch(pattern, origins[0]['DomainName'])):
    raise SystemExit('Default CloudFront origin must be the Terraform output S3 bucket root')
PY

printf 'Building Travel Angels for %s (API %s)\n' "$web_url" "$api_url"
build_args=(build web --release --base-href=/ --no-web-resources-cdn
  "--dart-define=TRAVEL_ANGELS_WEB_URL=$web_url"
  "--dart-define=TRAVEL_ANGELS_API_URL=$api_url")
flutter_help=$("$flutter_bin" build web --help)
if [[ $flutter_help == *--pwa-strategy* ]]; then
  build_args+=(--pwa-strategy=none)
fi
(
  cd -- "$app_dir"
  "$flutter_bin" "${build_args[@]}"
)

release_dir="$release_tmp/release"
mkdir -- "$release_dir"
cp -R -- "$app_dir/build/web/." "$release_dir/"
for entry in index.html main.dart.js flutter_bootstrap.js; do
  [[ -s $release_dir/$entry ]] || fail "Build is missing $entry"
done

# Retire any older Flutter worker without touching localStorage/account tokens.
# New builds do not need offline caching; an existing worker will discover this
# replacement on its next update and remove only Flutter's shell caches.
cat > "$release_dir/flutter_service_worker.js" <<'JS'
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', event => event.waitUntil((async () => {
  await Promise.all(['flutter-app-cache', 'flutter-temp-cache', 'flutter-app-manifest']
    .map(name => caches.delete(name)));
  await self.clients.claim();
  await self.registration.unregister();
})()));
JS
git_sha=$(git -C "$repo_root" rev-parse HEAD)
git_dirty=false
if [[ -n $(git -C "$repo_root" status --porcelain --untracked-files=no) ]]; then
  git_dirty=true
fi
python3 - "$release_dir" "$git_sha" "$git_dirty" "$web_url" "$api_url" <<'PY'
import datetime, hashlib, json, pathlib, sys
root = pathlib.Path(sys.argv[1])
release = {
    'gitSha': sys.argv[2],
    'trackedWorkingTreeDirty': sys.argv[3] == 'true',
    'builtAt': datetime.datetime.now(datetime.timezone.utc).isoformat().replace('+00:00', 'Z'),
    'mainDartJsSha256': hashlib.sha256((root / 'main.dart.js').read_bytes()).hexdigest(),
    'webUrl': sys.argv[4],
    'apiUrl': sys.argv[5],
}
(root / 'release.json').write_text(json.dumps(release, indent=2) + '\n')
PY

# Unversioned Flutter filenames must revalidate. Do not delete previous objects:
# an already-open tab may still request an asset from the previous release.
cache_control='no-cache, max-age=0, must-revalidate'
entries=(main.dart.js flutter.js flutter_bootstrap.js version.json manifest.json release.json index.html flutter_service_worker.js)
asset_args=(s3 cp "$release_dir/" "s3://$bucket/" --recursive
  --cache-control "$cache_control" --no-progress --only-show-errors --exclude '*.wasm')
for entry in "${entries[@]}"; do
  asset_args+=(--exclude "$entry")
done
printf 'Uploading assets to %s\n' "$bucket"
aws "${asset_args[@]}"
while IFS= read -r -d '' wasm_file; do
  relative_path=${wasm_file#"$release_dir/"}
  aws s3 cp "$wasm_file" "s3://$bucket/$relative_path" --content-type application/wasm \
    --cache-control "$cache_control" --no-progress --only-show-errors
done < <(find "$release_dir" -type f -name '*.wasm' -print0)
for entry in "${entries[@]}"; do
  if [[ -f $release_dir/$entry ]]; then
    aws s3 cp "$release_dir/$entry" "s3://$bucket/$entry" \
      --cache-control "$cache_control" --no-progress --only-show-errors
  fi
done
invalidation=$(aws cloudfront create-invalidation --distribution-id "$distribution" \
  --paths '/*' --query Invalidation.Id --output text)
printf 'CloudFront invalidation: %s\nTravel Angels: %s\nRelease metadata: %s/release.json\n' \
  "$invalidation" "$web_url" "$web_url"
