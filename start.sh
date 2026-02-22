#!/bin/bash

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   SkyHigh Core - Digital Check-In System ${NC}"
echo -e "${BLUE}   (Running fully on Docker)              ${NC}"
echo -e "${BLUE}=========================================${NC}"

# Check docker
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running or not accessible."
  exit 1
fi

echo -e "\n${GREEN}[1/1] Starting All Services (Db, Redis, Backend, Frontend)...${NC}"
# Use --build to ensure code changes are picked up
docker-compose up --build -d

if [ $? -ne 0 ]; then
    echo "Error starting docker-compose."
    exit 1
fi

echo -e "\n${BLUE}=========================================${NC}"
echo -e "${GREEN} System Running! ${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "Booking Backend:  http://localhost:8080"
echo -e "CheckIn Backend:  http://localhost:8081"
echo -e "Frontend:         http://localhost:5173"
echo -e "\nUse 'docker-compose logs -f' to see service logs."
echo -e "Use 'docker-compose down' to stop everything."

# Optional: tail logs
# docker-compose logs -f
