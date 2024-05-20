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

# Set OS/ARCH
SUPPORTED_OS_LIST=("linux" "darwin" "windows")
SUPPORTED_ARCH_LIST=("amd64" "arm64")

# Set Project ID
PROJECT_ID=11320

# Set Version
VERSION="$(yq .BigBangCliVersion $PACKAGE_DIR/static/resources/constants.yaml)"

# Set URLs
BASE_REPO_URL="https://repo1.dso.mil"
PROJECT_URL="${BASE_REPO_URL}/bigbang/product/packages/bbctl"
PROJECT_API_URL="${BASE_REPO_URL}/api/v4/projects/$PROJECT_ID"

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
run_command "git tag -a \"$VERSION\" -m \"bbctl Release $VERSION\""
run_command "git push origin \"$VERSION\""

# Build and push
for OS in "${SUPPORTED_OS_LIST[@]}"; do
  for ARCH in "${SUPPORTED_ARCH_LIST[@]}"; do
    echo "---"
    BINARY_NAME="$PACKAGE_NAME-$VERSION-$OS-$ARCH"
    echo "Building in $PACKAGE_DIR for $BINARY_NAME..."
    if [ "$SKIP_BUILD" == "true" ]; then
      echo "Skipping build..."
      continue
    fi
    GOOS=$OS GOARCH=$ARCH go build -o "$PACKAGE_DIR/bin/$BINARY_NAME"
    echo "Creating tarball..."
    TARBALL="$BINARY_NAME.tar.gz"
    TARBALL_PATH="$PACKAGE_DIR/bin/$TARBALL"
    tar -czf "$TARBALL_PATH" -C "$PACKAGE_DIR/bin" "$BINARY_NAME"
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
  done
done

# Create Release
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
RELEASE_JSON='{
  "name": "'$VERSION'",
  "tag_name": "'$VERSION'", 
  "description": "BigBangCli Release '$VERSION'\n\n---\n\n'$LOG_MESSAGES'",
  "assets": { "links": '$ALL_LINKS' }
}'
# https://docs.gitlab.com/ee/api/releases/#create-a-release
run_command "curl --header 'Content-Type: application/json' --header \"PRIVATE-TOKEN: $REPO1_TOKEN\" \
     --data '$RELEASE_JSON' \
     --request POST \"${PROJECT_API_URL}/releases\""
