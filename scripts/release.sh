#! /usr/bin/env bash

# Exit on error
set -e

# Run command function
run_command() {
  if [ "$DRY_RUN" != "false" ]; then
    echo "$@"
  else
    eval "$@"
  fi
}

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Check if dry-run is set
if [ "$DRY_RUN" != "false" ]; then
  echo "DRY_RUN is not set to 'false'. This will echo the write commands for git/gitlab instead of running them."
fi

# Set Project ID
PROJECT_ID=11320

# Set Version
VERSION="$(yq .BigBangCliVersion $PACKAGE_DIR/static/resources/constants.yaml)"

# Set URLs
BASE_REPO_URL="https://repo1.dso.mil"
PROJECT_URL="${BASE_REPO_URL}/bigbang/product/packages/bbctl"
PROJECT_API_URL="${BASE_REPO_URL}/api/v4/projects/$PROJECT_ID"

# Find MR type
# TODO: When moved into the pipes this will be switched to look at the MR changeset
CHANGESET=$(git diff --name-only main HEAD)
BINARY_CHANGESET="^bin/.*|^cmd/.*|^mocks/.*|^static/.*|^util/.*|^\.dockerignore|^\.mockery\.yaml|^Dockerfile|^go\.mod|^go\.sum|^main\.go"
GREP_EXPRESSIONS="-e $(echo "$BINARY_CHANGESET" | sed 's/|/ -e /g')"
# This allows for overriding the change type
if [ -z "$CHART_CHANGED" ]; then
  CHART_CHANGED=$(echo "$CHANGESET" | grep -E ^chart/.* | cat)
  if [ -n "$CHART_CHANGED" ]; then
    CHART_CHANGED="true"
  else
    CHART_CHANGED="false"
  fi
fi
# This allows for overriding the change type
if [ -z "$BINARY_CHANGED" ]; then
  BINARY_CHANGED=$(eval "echo \"$CHANGESET\" | grep $GREP_EXPRESSIONS" | cat)
  if [ -n "$BINARY_CHANGED" ]; then
    BINARY_CHANGED="true"
  else
    BINARY_CHANGED="false"
  fi
fi

# Check if both chart and binary are changed
if [ "$CHART_CHANGED" == "true" ] && [ "$BINARY_CHANGED" == "true" ]; then
  echo "Both chart and binary are changed. Please separate the changes into different MRs, run 'git diff --name-only main HEAD'" 1>&2
  exit 1
fi

# Check if neither chart nor binary are changed
if [ "$CHART_CHANGED" != "true" ] && [ "$BINARY_CHANGED" != "true" ]; then
  echo "Neither chart nor binary are changed. Please make changes to either the chart or the binary, run 'git diff --name-only main HEAD'" 1>&2
  exit 1
fi

if [ "$CHART_CHANGED" == "true" ]; then
  echo "Chart is changed"
  # Check if the Chart AppVersion is the same as the BigBangCliVersion
  CHART_APP_VERSION="$(yq .appVersion $PACKAGE_DIR/chart/Chart.yaml)"
  if [ "$CHART_APP_VERSION" != "$VERSION" ]; then
    echo "BigBangCliVersion ($VERSION) is not the same as the Chart appVersion ($CHART_APP_VERSION)" 1>&2
    exit 1
  fi

  CHART_VERSION="$(yq .version $PACKAGE_DIR/chart/Chart.yaml)"
  CHART_VERSION_APP_PART=$(echo "$CHART_VERSION" | awk -F'-' '{print $1}')
  # Check if the Chart Version is the same as the BigBangCliVersion
  if [ "$CHART_VERSION_APP_PART" != "${VERSION}" ]; then
    echo "BigBangCliVersion ($VERSION) is not the same as the Chart version \"app part\" ($CHART_VERSION_APP_PART), removing '\-bb\.[0-9]*'" 1>&2
    exit 1
  fi
  # Check if the Chart Version's BB part is the correct format
  CHART_VERSION_BB_PART=$(echo "$CHART_VERSION" | awk -F'-' '{print $2}')
  if ! [[ "$CHART_VERSION_BB_PART" =~ ^bb\.[0-9]*$ ]]; then
    echo "Chart version's BB part ($CHART_VERSION_BB_PART) is not the correct format" 1>&2
    exit 1
  fi
  # Check if the current Chart Version is different than the latest Chart Version
  # TODO: this will need to be commented out on the first release, because it won't work until the chart is on main
  LATEST_CHART_VERSION=$(curl -L https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/main/chart/Chart.yaml 2>/dev/null | yq .version)
  if [ "$CHART_VERSION" == "$LATEST_CHART_VERSION" ]; then
    echo "Current Chart Version ($CHART_VERSION) is the same as the latest Chart Version ($LATEST_CHART_VERSION)" 1>&2
    exit 1
  fi
  # Set the version to the Chart Version for the rest of the script
  VERSION="$CHART_VERSION"
else
  echo "Binary is changed"
  # Check if the current BigBangCliVersion is different than the latest BigBangCliVersion
  LATEST_VERSION=$(curl -L https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/main/static/resources/constants.yaml 2>/dev/null | yq .BigBangCliVersion)
  if [ "$VERSION" == "$LATEST_VERSION" ]; then
    echo "Current BigBangCliVersion ($VERSION) is the same as the latest BigBangCliVersion ($LATEST_VERSION)" 1>&2
    exit 1
  fi
fi

# Check if the latest CHANGELOG.md entry is the same as the Version
CHANGELOG_VERSION=$(grep --after-context 2 '\-\-\-' "$PACKAGE_DIR/CHANGELOG.md" | grep -E '\[[0-9]*\.[0-9]*\.[0-9](\-bb\.[0-9]*)*\]' | sed 's/.*\[//g' | sed 's/\].*//g')
if [ "$CHANGELOG_VERSION" != "$VERSION" ]; then
  NAME="BigBangCliVersion"
  if [ "$CHART_CHANGED" == "true" ]; then
    NAME="ChartVersion"
  fi
  echo "$NAME ($VERSION) is not the same as the latest CHANGELOG.md entry ($CHANGELOG_VERSION)" 1>&2
  exit 1
fi

# Check if REPO1_TOKEN is set
if [ -z "$REPO1_TOKEN" ]; then
  echo "REPO1_TOKEN is not set" 1>&2
  exit 1
fi

# Check if there are any .tar.gz file in the bin directory
if [ "$(ls -1 "$PACKAGE_DIR/bin" | grep .tar.gz | wc -l)" -ne 0 ]; then
  echo "There should be exactly no .tar.gz file in the bin directory" 1>&2
  exit 1
fi

# Get the messages for the since the last tag
LAST_TAG=$(git describe --tags --abbrev=0)
# error if no tags found
if [ -z "$LAST_TAG" ]; then
  echo "No tags found" 1>&2
  exit 1
fi
# error if last tag isn't semver
if ! [[ "$LAST_TAG" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Last tag is not a semver tag" 1>&2
  exit 1
fi
LOG_MESSAGES=$(git log "$LAST_TAG"..HEAD --pretty="%B%n---%n" | sed "s/'/\\\"/g" | sed 's/"/\\"/g' | sed 's/$/\\n/' | tr -d '\n' | sed 's/\\n$//g')

# Tag the release
TAG_COMMENT="bbctl Release $VERSION"
if [ "$CHART_CHANGED" == "true" ]; then
  TAG_COMMENT="bbctl Chart Release $VERSION"
fi
run_command "git tag -a \"$VERSION\" -m \"$TAG_COMMENT\""
run_command "git push origin \"$VERSION\""

# Build and push
if [ "$CHART_CHANGED" == "true" ]; then
  # TODO: this needs to run the regular chart stuff when we upgrade the pipes
  echo "Creating chart tarball..."
  CHART_NAME="$PACKAGE_NAME-$VERSION"
  TARBALL="$CHART_NAME.tar.gz"
  TARBALL_PATH="bin/$TARBALL" # $PACKAGE_DIR/
  FILES_TO_IGNORE=($(echo "$BINARY_CHANGESET" | sed -e 's/\^//g' -e 's/\*|/ /g' -e 's/\\\././g' | tr '|' ' '))
  EXCLUDE_ARGS=""
  for FILE in "${FILES_TO_IGNORE[@]}"; do
    FILE_ADJUSTED="$FILE"
    if [[ "$FILE" =~ /\.+$ ]]; then
      FILE_ADJUSTED="./$(echo "$FILE" | sed 's/\/\.*$//g')"
    fi
    EXCLUDE_ARGS="$EXCLUDE_ARGS--exclude='${FILE_ADJUSTED}' " # ${PACKAGE_DIR}/
  done
  EXCLUDE_ARGS="$EXCLUDE_ARGS--exclude='./.git'"
  TAR_COMMAND="tar $EXCLUDE_ARGS -czf '$TARBALL_PATH' -C '$PACKAGE_DIR' --transform 's/\./$CHART_NAME/' ."
  eval "$TAR_COMMAND"
  echo "Pushing to gitlab..."
  if [ "$(ls -1 "$PACKAGE_DIR/bin" | grep .tar.gz | wc -l)" -ne 1 ]; then
    echo "There should be exactly one .tar.gz file in the bin directory" 1>&2
    exit 1
  fi
  # https://docs.gitlab.com/ee/user/packages/generic_packages/#publish-a-package-file
  run_command "curl --header \"PRIVATE-TOKEN: $REPO1_TOKEN\" \
        --upload-file \"$TARBALL_PATH\" \
        \"${PROJECT_API_URL}/packages/generic/$PACKAGE_NAME/$VERSION/$TARBALL\""
  rm "$TARBALL_PATH"
else
  if [ "$DRY_RUN" != "false" ]; then
     goreleaser release --snapshot --clean
  else
     goreleaser release --clean
  fi
fi

# Create Helm Chart Release
if [ "$DRY_RUN" != "false" ]; then
  ALL_LINKS='[{"url": "https://repo1.dso.mil/bigbang/product/packages/bbctl/-/package_files/672/download", "name": "All-bbctl-0.0.0", "link_type": "other"}]'
  FULL_PACKAGE_URL="https://repo1.dso.mil/big-bang/product/packages/bbctl/-/packages/520"
else
  ## Get the package ID
  # https://docs.gitlab.com/ee/api/packages.html#list-packages
  PACKAGE_JSON=$(curl --header "PRIVATE-TOKEN: $REPO1_TOKEN" "${PROJECT_API_URL}/packages?package_name=$PACKAGE_NAME&package_version=$VERSION")
  PACKAGE_WEB_PATH=$(echo $PACKAGE_JSON | jq -r '.[0]._links.web_path')
  PACKAGE_ID=$(echo $PACKAGE_JSON | jq -r '.[0].id')
  FULL_PACKAGE_URL="${BASE_REPO_URL}$PACKAGE_WEB_PATH"

  ## Get all the package files
  # https://docs.gitlab.com/ee/api/packages.html#list-package-files
  ALL_PACKAGES_FULL_JSON="$(curl --header "PRIVATE-TOKEN: $REPO1_TOKEN" "${PROJECT_API_URL}/packages/$PACKAGE_ID/package_files")"
  ALL_LINKS="$(echo $ALL_PACKAGES_FULL_JSON | jq -r '[.[] | {url: "'$PROJECT_URL'/-/package_files/\(.id)/download", name: .file_name, link_type: "other"}]')"
fi
ALL_LINKS="$(echo $ALL_LINKS | jq -r '. += [{url: "'$FULL_PACKAGE_URL'", name: "'All-$PACKAGE_NAME-$VERSION'", link_type: "other"}]')"


## Create the release
if [ "$CHART_CHANGED" == "true" ]; then
  RELEASE_STRING="BigBangCli Chart Release"
  RELEASE_JSON='{
    "name": "'$VERSION'",
    "tag_name": "'$VERSION'", 
    "description": "'$RELEASE_STRING' '$VERSION'\n\n---\n\n'$LOG_MESSAGES'",
    "assets": { "links": '$ALL_LINKS' }
  }'
  # https://docs.gitlab.com/ee/api/releases/#create-a-release
  run_command "curl --header 'Content-Type: application/json' --header \"PRIVATE-TOKEN: $REPO1_TOKEN\" \
      --data '$RELEASE_JSON' \
      --request POST \"${PROJECT_API_URL}/releases\""
fi
