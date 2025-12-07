#!/bin/bash
set -e

# Setup colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Relia E2E Demo ===${NC}"

# Ensure Relia is built (using local dev env)
RELIA_CMD="./.venv/bin/python -m relia.main"

echo -e "\n${BLUE}[1] Estimating Costs for demo/main.tf...${NC}"
$RELIA_CMD estimate demo/main.tf

echo -e "\n${BLUE}[2] Checking Budget (Fail Case)...${NC}"
echo -e "Running: RELIA_BUDGET=50 relia check demo/main.tf"
# We expect this to fail (exit code 1)
set +e
RELIA_BUDGET=50 $RELIA_CMD check demo/main.tf
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -eq 1 ]; then
    echo -e "${GREEN}✓ Properly failed (Budget exceeded) as expected!${NC}"
else
    echo -e "${RED}✗ Failed to enforce budget!${NC}"
    exit 1
fi

echo -e "\n${BLUE}[3] Checking Budget (Pass Case)...${NC}"
echo -e "Running: RELIA_BUDGET=200 relia check demo/main.tf"
# t3.large (60) + m5.large (70) = 130. Limit 200 should pass.
RELIA_BUDGET=200 $RELIA_CMD check demo/main.tf

echo -e "\n${GREEN}=== E2E Demo Complete: SUCCESS ===${NC}"
