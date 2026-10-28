#!/usr/bin/env python3
"""Plan a URL-only ECS deployment; use --apply to register and update the service.

Uses current running image digests, never mutable tags. AWS_PROFILE and the usual
AWS CLI credentials apply. No credentials or other environment values are logged.
The generated task definition exists only in a mode-0600 temporary file.
"""

import argparse
import copy
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
from urllib.parse import urlsplit


READ_ONLY_FIELDS = {
    "taskDefinitionArn", "revision", "status", "requiresAttributes",
    "compatibilities", "registeredAt", "registeredBy", "deregisteredAt",
}
URL_VARIABLE = "TRAVEL_ANGELS_WEB_URL"


class DeploymentError(Exception):
    pass


def web_origin(value):
    parsed = urlsplit(value)
    if (
        parsed.scheme != "https"
        or not parsed.hostname
        or parsed.username is not None
        or parsed.password is not None
        or parsed.path not in ("", "/")
        or parsed.query
        or parsed.fragment
        or any(character.isspace() for character in value)
    ):
        raise argparse.ArgumentTypeError("--web-url must be an HTTPS origin without a path, credentials, query, or fragment")
    try:
        parsed.port
    except ValueError:
        raise argparse.ArgumentTypeError("--web-url has an invalid port") from None
    return value.rstrip("/")


class AWS:
    def __init__(self, args):
        self.args = args

    def call(self, service, operation, *arguments):
        command = [self.args.aws_cli, service, operation, "--region", self.args.region,
                   "--output", "json", *arguments]
        environment = dict(os.environ, AWS_PAGER="", AWS_CLI_AUTO_PROMPT="off")
        try:
            result = subprocess.run(command, capture_output=True, text=True,
                                    env=environment, timeout=120, check=False)
        except (OSError, subprocess.TimeoutExpired):
            raise DeploymentError(f"AWS {service} {operation} could not complete; no response details logged") from None
        if result.returncode:
            # CLI validation errors can repeat secret-bearing request fields.
            code = re.search(r"An error occurred \(([A-Za-z][A-Za-z0-9_.-]{0,79})\) when calling", result.stderr)
            reason = f"{code.group(1)}; " if code else ""
            raise DeploymentError(f"AWS {service} {operation} failed ({reason}exit {result.returncode}); response suppressed")
        try:
            return json.loads(result.stdout)
        except ValueError:
            raise DeploymentError(f"AWS {service} {operation} returned invalid JSON; response suppressed") from None


def check_cli_model(aws):
    skeleton = aws.call("ecs", "register-task-definition", "--generate-cli-skeleton", "input")
    mappings = skeleton.get("containerDefinitions", [{}])[0].get("portMappings", [{}])[0]
    if not {"name", "appProtocol"}.issubset(mappings) or "enableFaultInjection" not in skeleton:
        raise DeploymentError("AWS CLI ECS model is outdated and could drop task fields. Install a current AWS CLI or supply --aws-cli PATH.")
    return set(skeleton)


def service_snapshot(aws):
    args = aws.args
    response = aws.call("ecs", "describe-services", "--cluster", args.cluster,
                        "--services", args.service)
    services = response.get("services", [])
    if response.get("failures") or len(services) != 1:
        raise DeploymentError("The requested ECS service was not found")
    service = services[0]
    deployments = service.get("deployments", [])
    if (
        service.get("status") != "ACTIVE"
        or service.get("desiredCount", 0) < 1
        or service.get("runningCount") != service.get("desiredCount")
        or service.get("pendingCount") != 0
        or len(deployments) != 1
        or deployments[0].get("status") != "PRIMARY"
        or deployments[0].get("rolloutState") != "COMPLETED"
        or deployments[0].get("taskDefinition") != service.get("taskDefinition")
        or service.get("deploymentController", {}).get("type", "ECS") != "ECS"
    ):
        raise DeploymentError("Service must have one completed, healthy ECS deployment before this change")
    return service


def running_digests(aws, service, definition):
    task_arns = aws.call("ecs", "list-tasks", "--cluster", aws.args.cluster,
                         "--service-name", aws.args.service, "--desired-status", "RUNNING").get("taskArns", [])
    if len(task_arns) != service["desiredCount"] or len(task_arns) > 100:
        raise DeploymentError("Running task count changed or exceeds this script's supported batch size")
    response = aws.call("ecs", "describe-tasks", "--cluster", aws.args.cluster, "--tasks", *task_arns)
    tasks = response.get("tasks", [])
    if response.get("failures") or len(tasks) != len(task_arns):
        raise DeploymentError("Could not inspect every running task")
    expected = {container["name"] for container in definition["containerDefinitions"]}
    digests = {}
    for task in tasks:
        if task.get("taskDefinitionArn") != service["taskDefinition"] or task.get("lastStatus") != "RUNNING" or task.get("healthStatus") == "UNHEALTHY":
            raise DeploymentError("Running task state differs from the completed deployment")
        containers = {container["name"]: container for container in task.get("containers", [])}
        for name in expected:
            digest = containers.get(name, {}).get("imageDigest", "")
            if not re.fullmatch(r"sha256:[0-9a-f]{64}", digest):
                raise DeploymentError(f"Running image digest is unavailable for {name}")
            if name in digests and digests[name] != digest:
                raise DeploymentError(f"Running tasks have different image digests for {name}")
            digests[name] = digest
    return digests, sorted(task_arns)


def image_repository(image):
    reference = image.split("@", 1)[0]
    slash = reference.rfind("/")
    colon = reference.rfind(":")
    return reference[:colon] if colon > slash else reference


def build_definition(response, supported_fields, digests, web_url):
    original = response["taskDefinition"]
    unknown = set(original) - READ_ONLY_FIELDS - supported_fields
    if unknown:
        raise DeploymentError("Task definition contains unsupported fields; update the AWS CLI before cloning")
    definition = {key: copy.deepcopy(value) for key, value in original.items() if key not in READ_ONLY_FIELDS}
    # ECS rejects an explicit empty tags list even though its input shape allows it.
    if response.get("tags"):
        definition["tags"] = copy.deepcopy(response["tags"])
    core = [container for container in definition["containerDefinitions"] if container["name"] == "core"]
    if len(core) != 1:
        raise DeploymentError("Expected exactly one core container")
    if any(secret.get("name") == URL_VARIABLE for secret in core[0].get("secrets", [])):
        raise DeploymentError("WEB_URL is configured as a secret; refusing to create a conflicting environment entry")
    for container in definition["containerDefinitions"]:
        container["image"] = f"{image_repository(container['image'])}@{digests[container['name']]}"
    environment = core[0].setdefault("environment", [])
    environment[:] = [entry for entry in environment if entry.get("name") != URL_VARIABLE]
    environment.append({"name": URL_VARIABLE, "value": web_url})
    return definition


def assert_unchanged(aws, initial, definition, digests, task_arns):
    current = service_snapshot(aws)
    if current["serviceArn"] != initial["serviceArn"] or current["taskDefinition"] != initial["taskDefinition"]:
        raise DeploymentError("Service task definition changed during planning; restart the command")
    current_digests, current_tasks = running_digests(aws, current, definition)
    if current_digests != digests or current_tasks != task_arns:
        raise DeploymentError("Running task identities or image digests changed during planning; restart the command")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--web-url", required=True, type=web_origin)
    parser.add_argument("--region", default="us-east-1")
    parser.add_argument("--cluster", default="poltergeist")
    parser.add_argument("--service", default="sonar_core")
    parser.add_argument("--aws-cli", default="aws", help="Path to a current AWS CLI executable")
    parser.add_argument("--apply", action="store_true", help="Register the planned revision and update this service")
    args = parser.parse_args()
    aws = AWS(args)
    supported_fields = check_cli_model(aws)
    service = service_snapshot(aws)
    original_arn = service["taskDefinition"]
    identity = aws.call("sts", "get-caller-identity")
    arn = original_arn.split(":", 5)
    if len(arn) != 6 or arn[2:5] != ["ecs", args.region, identity["Account"]] or not arn[5].startswith("task-definition/"):
        raise DeploymentError("Task definition ARN does not match the selected AWS account and region")
    response = aws.call("ecs", "describe-task-definition", "--task-definition", original_arn, "--include", "TAGS")
    original = response["taskDefinition"]
    if original.get("taskDefinitionArn") != original_arn or original.get("status") != "ACTIVE":
        raise DeploymentError("Source task definition is no longer the expected active revision")
    digests, task_arns = running_digests(aws, service, original)
    definition = build_definition(response, supported_fields, digests, args.web_url)
    plan = {
        "mode": "apply" if args.apply else "dry-run",
        "serviceArn": service["serviceArn"],
        "rollbackTaskDefinitionArn": original_arn,
        "webUrl": args.web_url,
        "pinnedImages": {container["name"]: container["image"] for container in definition["containerDefinitions"]},
        "changes": ["Pin all application images to their current running digests", "Set only core.TRAVEL_ANGELS_WEB_URL"],
    }
    print(json.dumps(plan, indent=2), flush=True)
    if not args.apply:
        print("Dry run only. No task definition registered and no service updated.")
        return
    assert_unchanged(aws, service, original, digests, task_arns)
    # mkstemp creates the file with 0600, before any task configuration is written.
    descriptor, temporary_path = tempfile.mkstemp(prefix="travel-angels-task-", suffix=".json")
    try:
        with os.fdopen(descriptor, "w") as output:
            json.dump(definition, output)
        registered = aws.call("ecs", "register-task-definition", "--cli-input-json", f"file://{temporary_path}")
    finally:
        Path(temporary_path).unlink(missing_ok=True)
    new_arn = registered["taskDefinition"]["taskDefinitionArn"]
    prefix = original_arn.rsplit(":", 1)[0] + ":"
    if not new_arn.startswith(prefix) or not new_arn[len(prefix):].isdigit() or new_arn == original_arn:
        raise DeploymentError("Registered task ARN did not match the expected account, region, and family; service was not changed")
    print(json.dumps({"registeredTaskDefinitionArn": new_arn}), flush=True)
    # A concurrent deploy after registration must not be overwritten.
    assert_unchanged(aws, service, original, digests, task_arns)
    updated = aws.call("ecs", "update-service", "--cluster", args.cluster,
                       "--service", service["serviceArn"], "--task-definition", new_arn)
    if updated.get("service", {}).get("taskDefinition") != new_arn:
        raise DeploymentError("Update response did not confirm the requested revision; inspect service state")
    print(json.dumps({"deploymentStarted": True, "taskDefinitionArn": new_arn,
                      "rollbackTaskDefinitionArn": original_arn}), flush=True)
    print("Monitor ECS service stability and API health. This command does not wait for rollout.")


if __name__ == "__main__":
    try:
        main()
    except (DeploymentError, KeyError, TypeError) as error:
        message = str(error) if isinstance(error, DeploymentError) else "AWS returned an unexpected response shape; details suppressed"
        print(f"Error: {message}", file=sys.stderr)
        sys.exit(1)
