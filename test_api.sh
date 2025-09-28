#!/bin/bash

# Chat App API Test Script
# Tests all REST API endpoints for the chat application

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 Starting Chat App API Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test results
print_test() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Test variables
USER1_EMAIL="user1@test.com"
USER1_PASSWORD="password123"
USER2_EMAIL="user2@test.com"
USER2_PASSWORD="password123"
ROOM_NAME="Test Room"

echo -e "\n${BLUE}=== 1. User Registration ===${NC}"

# Register User 1
echo "📝 Registering User 1..."
REGISTER1_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "user_email": "'${USER1_EMAIL}'",
    "user_password": "'${USER1_PASSWORD}'",
    "user_fullname": "Test User 1"
  }')

HTTP_CODE1=$(echo $REGISTER1_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE1" = "200" ] && echo 0 || echo 1) "User 1 registration (HTTP: $HTTP_CODE1)"

# Register User 2
echo "📝 Registering User 2..."
REGISTER2_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "user_email": "'${USER2_EMAIL}'",
    "user_password": "'${USER2_PASSWORD}'",
    "user_fullname": "Test User 2"
  }')

HTTP_CODE2=$(echo $REGISTER2_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE2" = "200" ] && echo 0 || echo 1) "User 2 registration (HTTP: $HTTP_CODE2)"

echo -e "\n${BLUE}=== 2. User Login ===${NC}"

# Login User 1
echo "🔑 Logging in User 1..."
LOGIN1_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "user_email": "'${USER1_EMAIL}'",
    "user_password": "'${USER1_PASSWORD}'"
  }')

HTTP_CODE=$(echo $LOGIN1_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
LOGIN1_BODY=$(echo $LOGIN1_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
USER1_TOKEN=$(echo $LOGIN1_BODY | jq -r '.data.token')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "User 1 login (HTTP: $HTTP_CODE)"

# Login User 2
echo "🔑 Logging in User 2..."
LOGIN2_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "user_email": "'${USER2_EMAIL}'",
    "user_password": "'${USER2_PASSWORD}'"
  }')

HTTP_CODE=$(echo $LOGIN2_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
LOGIN2_BODY=$(echo $LOGIN2_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
USER2_TOKEN=$(echo $LOGIN2_BODY | jq -r '.data.token')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "User 2 login (HTTP: $HTTP_CODE)"

echo -e "\n${YELLOW}User 1 Token: ${USER1_TOKEN:0:50}...${NC}"
echo -e "${YELLOW}User 2 Token: ${USER2_TOKEN:0:50}...${NC}"

echo -e "\n${BLUE}=== 3. Room Management ===${NC}"

# Create Room
echo "🏠 Creating room..."
CREATE_ROOM_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/rooms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${USER1_TOKEN}" \
  -d '{
    "room_name": "'${ROOM_NAME}'",
    "is_direct_chat": false
  }')

HTTP_CODE=$(echo $CREATE_ROOM_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
CREATE_ROOM_BODY=$(echo $CREATE_ROOM_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
ROOM_ID=$(echo $CREATE_ROOM_BODY | jq -r '.data.room_id')
ROOM_CODE=$(echo $CREATE_ROOM_BODY | jq -r '.data.room_code')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Room creation (HTTP: $HTTP_CODE, Room ID: $ROOM_ID)"

# List User 1's Rooms
echo "📋 Listing User 1's rooms..."
LIST_ROOMS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/rooms" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $LIST_ROOMS_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "List rooms (HTTP: $HTTP_CODE)"

# User 2 joins room by code
echo "🚪 User 2 joining room by code..."
JOIN_ROOM_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/rooms/join-by-code" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${USER2_TOKEN}" \
  -d '{
    "room_code": "'${ROOM_CODE}'"
  }')

HTTP_CODE=$(echo $JOIN_ROOM_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "User 2 join room by code (HTTP: $HTTP_CODE)"

# User 2 joins room by ID (alternative method)
echo "🚪 User 2 joining room by ID..."
JOIN_ROOM_ID_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/rooms/${ROOM_ID}/join" \
  -H "Authorization: Bearer ${USER2_TOKEN}")

HTTP_CODE=$(echo $JOIN_ROOM_ID_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "User 2 join room by ID (HTTP: $HTTP_CODE)"

# Get room members
echo "👥 Getting room members..."
MEMBERS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/rooms/${ROOM_ID}/members" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $MEMBERS_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
MEMBERS_BODY=$(echo $MEMBERS_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
MEMBER_COUNT=$(echo $MEMBERS_BODY | jq -r '.data | length')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Get room members (HTTP: $HTTP_CODE, Members: $MEMBER_COUNT)"

echo -e "\n${BLUE}=== 4. Message Management ===${NC}"

# Send message via REST API (if available)
echo "💬 Sending test message..."
MESSAGE_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/rooms/${ROOM_ID}/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${USER1_TOKEN}" \
  -d '{
    "content": "Hello from REST API test!"
  }')

HTTP_CODE=$(echo $MESSAGE_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Send message (HTTP: $HTTP_CODE)"

# Get room messages
echo "📬 Getting room messages..."
MESSAGES_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/rooms/${ROOM_ID}/messages?limit=10&offset=0" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $MESSAGES_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
MESSAGES_BODY=$(echo $MESSAGES_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
MESSAGE_COUNT=$(echo $MESSAGES_BODY | jq -r '.data | length')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Get room messages (HTTP: $HTTP_CODE, Messages: $MESSAGE_COUNT)"

echo -e "\n${BLUE}=== 5. WebSocket Status ===${NC}"

# Check WebSocket room status
echo "📊 Checking WebSocket room status..."
WS_STATUS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/chat/rooms/${ROOM_ID}/status" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $WS_STATUS_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
WS_STATUS_BODY=$(echo $WS_STATUS_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
CLIENT_COUNT=$(echo $WS_STATUS_BODY | jq -r '.client_count')
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "WebSocket room status (HTTP: $HTTP_CODE, Clients: $CLIENT_COUNT)"

echo -e "\n${BLUE}=== 6. Leave Room Test ===${NC}"

# User 2 leaves room
echo "🚪 User 2 leaving room..."
LEAVE_ROOM_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
  "${BASE_URL}/rooms/${ROOM_ID}/leave" \
  -H "Authorization: Bearer ${USER2_TOKEN}")

HTTP_CODE=$(echo $LEAVE_ROOM_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "User 2 leave room (HTTP: $HTTP_CODE)"

# Check members after leave
echo "👥 Checking members after leave..."
MEMBERS_AFTER_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/rooms/${ROOM_ID}/members" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $MEMBERS_AFTER_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
MEMBERS_AFTER_BODY=$(echo $MEMBERS_AFTER_RESPONSE | sed 's/HTTPSTATUS:[0-9]*//g')
MEMBER_COUNT_AFTER=$(echo $MEMBERS_AFTER_BODY | jq -r '.data | length')
print_test $([ "$HTTP_CODE" = "200" ] && [ "$MEMBER_COUNT_AFTER" = "1" ] && echo 0 || echo 1) "Members after leave (HTTP: $HTTP_CODE, Members: $MEMBER_COUNT_AFTER)"

echo -e "\n${BLUE}=== 7. Admin Tests ===${NC}"

# Get all rooms (admin endpoint)
echo "🔐 Getting all rooms (admin)..."
ADMIN_ROOMS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/admin/rooms?limit=10&offset=0" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $ADMIN_ROOMS_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Admin get all rooms (HTTP: $HTTP_CODE)"

# Get all users (admin endpoint)
echo "👤 Getting all users (admin)..."
ADMIN_USERS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET \
  "${BASE_URL}/admin/users?limit=10&offset=0" \
  -H "Authorization: Bearer ${USER1_TOKEN}")

HTTP_CODE=$(echo $ADMIN_USERS_RESPONSE | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
print_test $([ "$HTTP_CODE" = "200" ] && echo 0 || echo 1) "Admin get all users (HTTP: $HTTP_CODE)"

echo -e "\n${GREEN}✅ API Tests Completed!${NC}"

echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo "Room ID: $ROOM_ID"
echo "Room Code: $ROOM_CODE"
echo "User 1 Token: ${USER1_TOKEN:0:50}..."
echo "User 2 Token: ${USER2_TOKEN:0:50}..."

echo -e "\n${BLUE}You can now test WebSocket with:${NC}"
echo "node test_websocket_enhanced.js"
echo -e "\n${BLUE}Or manually connect to WebSocket:${NC}"
echo "ws://localhost:8080/api/v1/chat/ws?token=${USER1_TOKEN}&room_id=${ROOM_ID}"