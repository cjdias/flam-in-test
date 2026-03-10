#!/bin/sh

# Define some colors
YELLOW='\033[0;33m'
BOLD_YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD_GREEN='\033[1;32m'
NC='\033[0m' # No Color

# Print the message inside a box
echo "${CYAN}=======================================${NC}"
echo "${BOLD_YELLOW} Running pre-commit tests... ${NC}"
echo "${CYAN}=======================================${NC}"

echo # Add a blank line for spacing
echo "${YELLOW} Running code formatter ...${NC}"
echo # Add a blank line for spacing

task quality:format

if [ $? -ne 0 ]; then
    echo 
    echo "${RED}✗${NC} Code foramtting failed. Aborting commit."
    exit 1 # Exit the script with a non-zero code to indicate failure
fi

echo
echo "${GREEN} Formatting passed"
echo

echo # Add a blank line for spacing
echo "${YELLOW} Running static analyser (linter) ...${NC}"
echo # Add a blank line for spacing

task quality:lint

if [ $? -ne 0 ]; then
    echo 
    echo "${RED}✗${NC} Linting failed. Aborting commit."
    exit 1 # Exit the script with a non-zero code to indicate failure
fi

echo
echo "${GREEN} Linting passed"
echo

echo # Add a blank line for spacing
echo "${YELLOW} Running package tests ...${NC}"
echo # Add a blank line for spacing

task quality:tests-coverage

if [ $? -ne 0 ]; then
    echo 
    echo "${RED}✗${NC} Tests failed. Aborting commit."
    exit 1 # Exit the script with a non-zero code to indicate failure
fi

echo
echo "${GREEN} Tests passed"
echo

# Print the message inside a box
echo "${CYAN}=======================================${NC}"
echo "${BOLD_GREEN} All pre-commit tests passed ${NC}"
echo "${CYAN}=======================================${NC}"
echo # Add a blank line for spacing
