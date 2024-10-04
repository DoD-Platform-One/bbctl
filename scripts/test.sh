#! /usr/bin/env bash

# Exit on error
set -e

# Coverage minimum quality thresholds
# Percent value times 100 since bash can't deal with floating point math
ERROR_THRESHOLD=8000    # 80%
WARNING_THRESHOLD=9000  # 90%

# Output colors
BOLD_RED='\033[1;31m'
BOLD_YELLOW='\033[1;33m'
BOLD_GREEN='\033[1;32m'
COLOR_RESET="\e[0m"

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Run tests
echo "Running tests in $PACKAGE_DIR..."
go test ./... -v --coverprofile=cover.txt

# Generate file-by-file coverage report
go tool cover -html=cover.txt -o coverage.html
# Parse HTML file output where the lines look like:
# <option value="file0">repo1.dso.mil/big-bang/product/packages/bbctl/cmd/cmd.go (97.6%)</option>
echo "Coverage Per File Summary:"
cat coverage.html |
    grep '<option value="file' |
    sed 's/ value="file/ /' |
    sed 's/">/ /' |
    sed 's/<\/option>/ /' |
    sed 's/repo1.dso.mil\/big-bang\/product\/packages\// /' |
    sed 's/(/ /' |
    sed 's/)/ /' |
    awk '{printf("%-90s %-5s\n", $3, $4)}' > output.txt
cat output.txt

# Scan for files with less than the desired code coverage
COVERAGE_FAILURES=()
COVERAGE_WARNINGS=()
while IFS= read -r line; do
    FILE="$(echo $line | awk '{print $1}')"
    COVERAGE="$(echo $line | awk '{print $2}')"
    CVG_FLOAT="$(printf "%.0f" "${COVERAGE::-1}e2")"
    # Exclude anything under /util/test/, /mocks/, and main.go
    if [[ ${FILE} != *"/util/test/"* ]] && [[ ${FILE} != *"/mocks/"* ]] && [[ ${FILE} != *"/main.go" ]]; then
        if [[ ${CVG_FLOAT} -lt ${ERROR_THRESHOLD} ]]; then
            COVERAGE_FAILURES+=("${FILE}: ${COVERAGE}")
        elif [[ ${CVG_FLOAT} -lt ${WARNING_THRESHOLD} ]]; then
            COVERAGE_WARNINGS+=("${FILE}: ${COVERAGE}")
        fi
    fi
done < output.txt

# Print all warnings regardless of pass/fail result
if [ ${#COVERAGE_WARNINGS[@]} -ne 0 ]; then
    echo -e "${BOLD_YELLOW}WARNING: The following files do meet the required minimum coverage of ${WARNING_THRESHOLD::-2}%! ${COLOR_RESET}"
    for warn in "${COVERAGE_WARNINGS[@]}"
    do
        echo "$warn"
    done
fi
# Print any errors and fail the pipeline with non-zero exit code
if [ ${#COVERAGE_FAILURES[@]} -ne 0 ]; then
    echo -e "${BOLD_RED}ERROR: The following files do meet the required minimum coverage of ${ERROR_THRESHOLD::-2}%! ${COLOR_RESET}"
    for err in "${COVERAGE_FAILURES[@]}"
    do
        echo "$err"
    done
    echo -e "${BOLD_RED}Pipeline failed due to lack of test coverage ${COLOR_RESET}"
    exit 1
else
    echo -e "${BOLD_GREEN}Passed minimum test coverage check ${COLOR_RESET}"
fi

# Exit successfully if no coverage quality errors were found
exit 0
