#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

#  This function checks the formatting of a changelog file to ensure it follows the correct structure and conventions.
#  It verifies that the changelog starts with '# Changelog', each version has a change type header,
#  each change type has at least one comment, and the changelog ends with a new line.
changelog_format_check() {
  firstLine=1
  hasAtLeastOneVersion=0
  hasAtLeastOneTypeOfChange=0
  hasAtLeastOneComment=0
  exitFlag=0
  hasComment=1
  hasTypeOfChange=1
  nonstandardHeader=0

  # Adds a new line to end of changelog for proper parsing
  if [ "$(tail -c 1 ./CHANGELOG.md)" != "" ]; then
    echo "" >> ./CHANGELOG.md
    echo -e "\e[31mError: Changelog must end with a new line.\e[0m"
    exitFlag=1
  fi

  while IFS= read -r line; do
    if [[ $firstLine == 1 ]]; then
      # ensure first line says changelog
      if [[ ! "$line"  =~ ^\#[[:space:]]Changelog ]]; then
        echo -e "\e[31mError: Changelog must start with '# Changelog'. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
        exitFlag=1
      fi
      firstLine=0
    fi
    # Check for version/section header
    if [[ "$line" =~ ^\#\#[[:space:]].+ ]]; then
      if [[ "$line" =~ ^\#\#[[:space:]]\[[[0-9]+\.[0-9]+\.[0-9]+.*\].* ]]; then
        # version header
        if [[ $hasTypeOfChange == 0 ]]; then
          echo -e "\e[31mError: Changelog - version $prevVersion is missing a changetype header. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
          exitFlag=1
        fi
        hasTypeOfChange=0
        hasAtLeastOneVersion=1
        if [[ $nonstandardHeader == 0 ]]; then
          prevVersion=$line
        else
          # we had been ignoring the lines above this (IE malformed version) -- keep prevVersion the same
          nonstandardHeader=0
        fi
      elif [[ "$line" =~ ^\#\#[[:space:]]\[[a-zA-Z]+\] ]]; then
        # section header
        # don't want to count anything below section title
        nonstandardHeader=1
      else
        echo -e "\e[31mError: Changelog header $line is in the wrong format. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
        exitFlag=1
        # malformed header, set to make sure we don't count anything that comes after it
        nonstandardHeader=1
      fi
    fi
    # Check for changetype
    if [[ "$line" =~ ^\#\#\#[[:space:]]+ && $nonstandardHeader == 0 ]]; then
      hasAtLeastOneTypeOfChange=1
      hasTypeOfChange=1
      if [[ $hasComment == 0 ]]; then
        echo -e "\e[31mError: Changelog - version $prevVersion is missing a comment for the [$prevChangetype] changetype. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
        exitFlag=1
      fi
      hasComment=0
      prevChangetype=$line
    fi
    # Check for comment
    if [[ "$line" =~ ^[[:space:]]*-[[:space:]] && $nonstandardHeader == 0 ]]; then
      hasAtLeastOneComment=1
      hasComment=1
    fi
  done < ./CHANGELOG.md
  # check final section format
  if [[ $hasComment == 0 ]]; then
    echo -e "\e[31mError: Changelog - version $prevVersion is missing a comment. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
    exitFlag=1
  fi
  if [[ $hasTypeOfChange == 0 ]]; then
    echo -e "\e[31mError: Changelog - version $prevVersion is missing a changetype header. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
    exitFlag=1
  fi
  # check globally if sections are missing
  if [[ $hasAtLeastOneVersion == 0 ]]; then
    echo -e "\e[31mError: Changelog is missing the app version (IE '## [1.0.0]') or is formatted incorrectly. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
    exitFlag=1
  fi
  if [[ $hasAtLeastOneTypeOfChange == 0 ]]; then
    echo -e "\e[31mError: Changelog is missing the changetype (IE '### Added' or '### Changed') or is formatted incorrectly. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
    exitFlag=1
  fi
  if [[ $hasAtLeastOneComment == 0 ]]; then
    echo -e "\e[31mError: Changelog is missing comments or they are formatted incorrectly. For correct formatting, see https://keepachangelog.com/en/1.0.0/ \e[0m"
    exitFlag=1
  fi
  if [[ $exitFlag == 1 ]]; then
    exit 1
  else
    echo -e "Changelog is valid"
  fi
}

# Find MR type
CHANGESET=$(git diff --name-only $CI_MERGE_REQUEST_SOURCE_BRANCH_NAME $CI_MERGE_REQUEST_TARGET_BRANCH_NAME)
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
  echo "Both chart and binary are changed. Please separate the changes into different MRs" 1>&2
  exit 1
fi

NAME=""
VERSION=""
source_version="new"
target_version="old"
if [ "$CHART_CHANGED" == "true" ]; then
  echo "helm chart is changed"
  NAME="helm chart"
  source_version="$(curl -Ls "https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/$CI_MERGE_REQUEST_SOURCE_BRANCH_NAME/static/resources/constants.yaml" | yq '.BigBangCliVersion')"
  target_version="$(curl -Ls "https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME/static/resources/constants.yaml" | yq '.BigBangCliVersion')"
else
  echo "Binary is changed"
  NAME="bbctl application"
  source_version="$(curl -Ls "https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/$CI_MERGE_REQUEST_SOURCE_BRANCH_NAME/static/resources/constants.yaml" | yq '.version')"
  target_version="$(curl -Ls "https://repo1.dso.mil/big-bang/product/packages/bbctl/-/raw/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME/static/resources/constants.yaml" | yq '.version')"
fi

if [ "$source_version" == "$target_version" ]; then
  echo "$NAME version must be incremented!" 1>&2
  exit 1
else
  VERSION=$source_version
fi
echo $NAME $VERSION

changelog_format_check

# Check if the latest CHANGELOG.md entry is the same as the Version
CHANGELOG_VERSION=$(grep --after-context 2 '\-\-\-' "$PACKAGE_DIR/CHANGELOG.md" | grep -E '\[[0-9]*\.[0-9]*\.[0-9](\-bb\.[0-9]*)*\]' | sed 's/.*\[//g' | sed 's/\].*//g')
if [ "$CHANGELOG_VERSION" != "$VERSION" ]; then
  echo "You must add a changelog entry for version $VERSION of the $NAME. Last CHANGELOG.md entry is for $CHANGELOG_VERSION" 1>&2
  exit 1
fi
