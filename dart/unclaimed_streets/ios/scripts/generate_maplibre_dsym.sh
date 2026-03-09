#!/bin/sh
set -euo pipefail

CONFIGURATION_NAME="${CONFIGURATION:-}"
if [ "$CONFIGURATION_NAME" != "Release" ] && [ "$CONFIGURATION_NAME" != "Profile" ]; then
  exit 0
fi

FRAMEWORK_BINARY="${TARGET_BUILD_DIR}/${FRAMEWORKS_FOLDER_PATH}/MapLibre.framework/MapLibre"
if [ ! -f "$FRAMEWORK_BINARY" ]; then
  echo "MapLibre.framework not found at ${FRAMEWORK_BINARY}; skipping dSYM generation."
  exit 0
fi

DSYM_OUTPUT="${DWARF_DSYM_FOLDER_PATH}/MapLibre.framework.dSYM"
if [ -d "$DSYM_OUTPUT" ]; then
  echo "MapLibre dSYM already exists at ${DSYM_OUTPUT}; skipping."
  exit 0
fi

echo "Generating dSYM for MapLibre.framework"
# dsymutil returns 0 even if debug info is minimal; we still want the UUID-matching dSYM.
dsymutil "$FRAMEWORK_BINARY" -o "$DSYM_OUTPUT" || {
  echo "warning: dsymutil failed for MapLibre.framework"
  exit 0
}
