#!/bin/bash

# Exit on error
set -e

# Check if parameters are empty:
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "please structure as such: ./add-cronjob.sh <parameter 1> <parameter 2>"
    echo "Aborting: One or more parameters empty"
    exit 1
fi

# Creating variables from parameters
lowerCopyFrom=${1,,}
upperCopyFrom=${lowerCopyFrom^}
bbJobFrom="bigbang$upperCopyFrom"

lowerJob=${2,,}
upperJob=${lowerJob^}
bbJob="bigbang$upperJob"

# Creating DIR and ROOT_DIR variables for later use
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$(dirname "$DIR")"
echo "Root Directory: $ROOT_DIR"
echo

# Read in name of old cronjob to copy from parameter 1
echo "Param 1: Name of existing cronjob to copy (exluding bigbang): $upperCopyFrom"
# Read in name of new cronjob from parameter 2
echo "Param 2: Name of new job name to create (excluding bigbang):  $upperJob"

# Read in name of old and new cronjobs
echo
echo "From:    Name of old cronjob created from (including bigbang): $bbJobFrom"
echo "To:      Name of new cronjob created into (including bigbang): $bbJob"

# Verify
echo
echo "Verify if above values are correct. Continue (y/n)?"
read yesNo
if [ -z "$yesNo" ]; then
    echo "Aborting: Input empty"
    exit 1
elif [ ${yesNo,,}  != "y" ]; then
    echo "Aborting: User verification refused"
    exit 2
fi

# Copy bigbangTemplate to given name of cronjob
echo
echo "cd $ROOT_DIR/chart/templates"
cd $ROOT_DIR/chart/templates
echo
echo "Creating $bbJob using $bbJobFrom as template"
cp -r $bbJobFrom $bbJob
newDir=$ROOT_DIR/chart/templates/$bbJob
cd $newDir

# Replacing all instances of old command name to new command name
echo
echo "Replacing all instances of $lowerCopyFrom with $lowerJob"
for filename in *;
do
    sed -i "s/$lowerCopyFrom/$lowerJob/g" $filename
    echo "$filename file checked and replaced all lowers"
    sed -i "s/$upperCopyFrom/$upperJob/g" $filename
    echo "$filename file checked and replaced all uppers"
done

# Adding new cronjob to values.yaml
echo
newValue=$(cat <<- NEW_VALUE

$bbJob:
  enabled: true
  schedule: "0 * * * *"
  bigbangReleaseName: "bigbang"
  bigbangReleaseNamespace: "bigbang"
  labels: {}
  config: {}
  podAnnotations: {}
  serviceAccount:
    # Specifies whether a service account should be created
    create: true
    # Annotations to add to the service account
    annotations: {}
    # The override name of the service account to use.
    # If not set and create is true, -bigbang-violations is appended to .Values.serviceAccount.name
    name: ""
NEW_VALUE
)

echo "cd $ROOT_DIR/chart/"
cd $ROOT_DIR/chart/
echo "Adding the following to values.yaml:"
echo "$newValue"
echo "$newValue" >> values.yaml
echo
echo "Complete! Please verify the following:"
echo "  - verify the created files in $ROOT_DIR/chart/templates/$bbJob "
echo "    to no longer have any instances of $lowerCopyFrom or $upperCopyFrom"
echo "  - verify that $ROOT_DIR/chart/values.yaml has a new entry with the desired cronjob definition"
echo "  - verify that $ROOT_DIR/chart/templates/$bbJob/_helpers.tpl has all the necessary arguments defined"
echo "  - adjust the schedule portion as needed in the new definition "
echo "    $ROOT_DIR/chart/values.yaml has a new entry with the desired cronjob"