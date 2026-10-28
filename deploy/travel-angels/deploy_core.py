#!/usr/bin/env python3
"""Plan a core-only ECS image release; register and deploy only with --apply.

The image must already exist in the account's core ECR repository. Every other
container must remain pinned to its currently running digest. Task configuration
and tags are cloned without changing environment variables, secrets, or URLs.
"""

import argparse
import copy
import json
import os
from pathlib import Path
import re
import sys
import tempfile

from configure_web_url import (
    AWS,
    DeploymentError,
    READ_ONLY_FIELDS,
    assert_unchanged,
    check_cli_model,
    image_repository,
    running_digests,
    service_snapshot,
)


def image_digest(value):
    if not re.fullmatch(r"sha256:[0-9a-f]{64}", value):
        raise argparse.ArgumentTypeError("--image-digest must be a SHA-256 image digest")
    return value


def registration_fields(response, supported_fields):
    original = response["taskDefinition"]
    if set(original) - READ_ONLY_FIELDS - supported_fields:
        raise DeploymentError("Task definition contains unsupported fields; update the AWS CLI before cloning")
    definition = {
        key: copy.deepcopy(value)
        for key, value in original.items()
        if key not in READ_ONLY_FIELDS
    }
    if response.get("tags"):
        definition["tags"] = copy.deepcopy(response["tags"])
    return definition


def build_definition(response, supported_fields, digests, repository, digest):
    definition = registration_fields(response, supported_fields)
    original = copy.deepcopy(definition)
    containers = definition["containerDefinitions"]
    names = [container["name"] for container in containers]
    if len(set(names)) != len(names) or names.count("core") != 1:
        raise DeploymentError("Expected unique container names and exactly one core container")
    for container in containers:
        name = container["name"]
        running_digest = digests.get(name, "")
        if not re.fullmatch(r"sha256:[0-9a-f]{64}", running_digest):
            raise DeploymentError("Every application container needs a verified running image digest")
        # Mutable source tags would make rollback or restarting other containers
        # unsafe. Require the already-pinned live release used by this service.
        if container["image"] != f"{image_repository(container['image'])}@{running_digest}":
            raise DeploymentError("All source container images must be pinned to their current running digests")
        if name == "core":
            if image_repository(container["image"]) != repository:
                raise DeploymentError("The running core image is not in the expected account/region core repository")
            container["image"] = f"{repository}@{digest}"

    # Verify a complete structural comparison, not just a list of known fields.
    # This includes current web URL, secrets, roles, volumes and container config.
    comparison = copy.deepcopy(definition)
    for container in comparison["containerDefinitions"]:
        if container["name"] == "core":
            container["image"] = next(
                item["image"] for item in original["containerDefinitions"]
                if item["name"] == "core"
            )
    if comparison != original:
        raise DeploymentError("Planned task definition changed fields other than core.image")
    return definition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--image-digest", required=True, type=image_digest)
    parser.add_argument("--account-id", default="872408892710")
    parser.add_argument("--region", default="us-east-1")
    parser.add_argument("--cluster", default="poltergeist")
    parser.add_argument("--service", default="sonar_core")
    parser.add_argument("--expected-task-definition", help="Optional exact ARN or family:revision to guard the reviewed source release")
    parser.add_argument("--aws-cli", default="aws")
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    if not re.fullmatch(r"\d{12}", args.account_id):
        raise DeploymentError("--account-id must be a 12-digit AWS account ID")
    aws = AWS(args)
    supported_fields = check_cli_model(aws)
    identity = aws.call("sts", "get-caller-identity")
    if identity.get("Account") != args.account_id:
        raise DeploymentError("AWS credentials do not match the expected account")
    service = service_snapshot(aws)
    original_arn = service["taskDefinition"]
    arn = original_arn.split(":", 5)
    if (
        len(arn) != 6
        or arn[:5] != ["arn", "aws", "ecs", args.region, args.account_id]
        or not arn[5].startswith("task-definition/")
    ):
        raise DeploymentError("Source task definition is not in the expected AWS account and region")
    if args.expected_task_definition and args.expected_task_definition not in (
        original_arn, arn[5].removeprefix("task-definition/")
    ):
        raise DeploymentError("Current task definition differs from --expected-task-definition")
    response = aws.call("ecs", "describe-task-definition", "--task-definition", original_arn, "--include", "TAGS")
    original = response["taskDefinition"]
    if original.get("taskDefinitionArn") != original_arn or original.get("status") != "ACTIVE":
        raise DeploymentError("Source task definition is no longer the expected active revision")
    digests, task_arns = running_digests(aws, service, original)
    image = aws.call("ecr", "describe-images", "--repository-name", "core", "--image-ids", f"imageDigest={args.image_digest}")
    details = image.get("imageDetails", [])
    if len(details) != 1 or details[0].get("imageDigest") != args.image_digest:
        raise DeploymentError("Requested digest was not found in the core ECR repository")
    repository = f"{args.account_id}.dkr.ecr.{args.region}.amazonaws.com/core"
    definition = build_definition(response, supported_fields, digests, repository, args.image_digest)
    changed = digests["core"] != args.image_digest
    plan = {
        "mode": "apply" if args.apply else "dry-run",
        "serviceArn": service["serviceArn"],
        "rollbackTaskDefinitionArn": original_arn,
        "oldCoreImage": f"{repository}@{digests['core']}",
        "newCoreImage": f"{repository}@{args.image_digest}",
        "preservedImages": {
            container["name"]: container["image"]
            for container in definition["containerDefinitions"]
            if container["name"] != "core"
        },
        "changes": ["Replace only core.image; preserve all other task fields and tags"] if changed else [],
    }
    print(json.dumps(plan, indent=2), flush=True)
    if not args.apply or not changed:
        print("No task definition registered and no service updated.")
        return

    assert_unchanged(aws, service, original, digests, task_arns)
    descriptor, temporary_path = tempfile.mkstemp(prefix="travel-angels-core-task-", suffix=".json")
    try:
        with os.fdopen(descriptor, "w") as output:
            json.dump(definition, output)
        registered = aws.call("ecs", "register-task-definition", "--cli-input-json", f"file://{temporary_path}")
    finally:
        Path(temporary_path).unlink(missing_ok=True)
    new_arn = registered["taskDefinition"]["taskDefinitionArn"]
    prefix = original_arn.rsplit(":", 1)[0] + ":"
    if not new_arn.startswith(prefix) or not new_arn[len(prefix):].isdigit() or new_arn == original_arn:
        raise DeploymentError("Registered task definition did not match the source family; service was not changed")
    print(json.dumps({"registeredTaskDefinitionArn": new_arn}), flush=True)
    # Concurrent deployments must never be overwritten, even after registration.
    assert_unchanged(aws, service, original, digests, task_arns)
    updated = aws.call("ecs", "update-service", "--cluster", args.cluster,
                       "--service", service["serviceArn"], "--task-definition", new_arn)
    if updated.get("service", {}).get("taskDefinition") != new_arn:
        raise DeploymentError("Update response did not confirm the requested revision; inspect service state")
    print(json.dumps({"deploymentStarted": True, "taskDefinitionArn": new_arn,
                      "rollbackTaskDefinitionArn": original_arn}), flush=True)
    print("Monitor ECS service stability, running digests and API health. This command does not wait for rollout.")


if __name__ == "__main__":
    try:
        main()
    except (DeploymentError, KeyError, TypeError) as error:
        message = str(error) if isinstance(error, DeploymentError) else "AWS returned an unexpected response shape; details suppressed"
        print(f"Error: {message}", file=sys.stderr)
        sys.exit(1)
