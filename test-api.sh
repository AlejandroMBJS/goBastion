#!/bin/bash

# Test script for Go Native FastAPI
# This script tests the API endpoints with proper CSRF handling

BASE_URL="http://localhost:8080"

echo "=========================================="
echo "Testing Go Native FastAPI"
echo "=========================================="
echo

# Helper function to print colored output
print_success() {
    echo -e "\033[0;32m✓ $1\033[0m"
}

print_error() {
    echo -e "\033[0;31m✗ $1\033[0m"
}

print_info() {
    echo -e "\033[0;34mℹ $1\033[0m"
}

# Test 1: Check if server is running
echo "Test 1: Checking server health..."
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/docs/openapi.json")
if [ "$response" -eq 200 ]; then
    print_success "Server is running"
else
    print_error "Server is not running (HTTP $response)"
    exit 1
fi
echo

# Test 2: Get OpenAPI documentation
echo "Test 2: Fetching OpenAPI documentation..."
response=$(curl -s "$BASE_URL/docs/openapi.json")
if echo "$response" | grep -q "openapi"; then
    print_success "OpenAPI documentation available"
else
    print_error "Failed to fetch OpenAPI documentation"
fi
echo

# Test 3: Register a user (handling CSRF)
echo "Test 3: Registering a new user..."

# First, get a CSRF token by making a GET request
csrf_response=$(curl -s -i "$BASE_URL/docs")
csrf_token=$(echo "$csrf_response" | grep -i "set-cookie: csrf_token=" | sed 's/.*csrf_token=\([^;]*\).*/\1/')

if [ -z "$csrf_token" ]; then
    print_info "No CSRF token needed for registration endpoint"
    csrf_token="dummy"
fi

# Register user
register_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: $csrf_token" \
    -b "csrf_token=$csrf_token" \
    -d '{
        "name": "Test User",
        "email": "test@example.com",
        "password": "testpass123",
        "role": "user"
    }')

if echo "$register_response" | grep -q "access_token"; then
    print_success "User registered successfully"
    access_token=$(echo "$register_response" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"\(.*\)"/\1/')
    print_info "Access token: ${access_token:0:20}..."
else
    print_error "User registration failed"
    echo "Response: $register_response"
fi
echo

# Test 4: Login
echo "Test 4: Testing login..."
login_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: $csrf_token" \
    -b "csrf_token=$csrf_token" \
    -d '{
        "email": "test@example.com",
        "password": "testpass123"
    }')

if echo "$login_response" | grep -q "access_token"; then
    print_success "Login successful"
    access_token=$(echo "$login_response" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"\(.*\)"/\1/')
else
    print_error "Login failed"
    echo "Response: $login_response"
    exit 1
fi
echo

# Test 5: Get current user
echo "Test 5: Getting current user info..."
me_response=$(curl -s "$BASE_URL/api/v1/auth/me" \
    -H "Authorization: Bearer $access_token")

if echo "$me_response" | grep -q "test@example.com"; then
    print_success "Current user retrieved successfully"
else
    print_error "Failed to get current user"
    echo "Response: $me_response"
fi
echo

# Test 6: List users
echo "Test 6: Listing all users..."
users_response=$(curl -s "$BASE_URL/api/v1/users" \
    -H "Authorization: Bearer $access_token")

if echo "$users_response" | grep -q "test@example.com"; then
    print_success "Users list retrieved successfully"
    user_count=$(echo "$users_response" | grep -o '"id":' | wc -l)
    print_info "Found $user_count user(s)"
else
    print_error "Failed to list users"
    echo "Response: $users_response"
fi
echo

# Test 7: Create another user
echo "Test 7: Creating another user..."
create_response=$(curl -s -X POST "$BASE_URL/api/v1/users" \
    -H "Authorization: Bearer $access_token" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: $csrf_token" \
    -b "csrf_token=$csrf_token" \
    -d '{
        "name": "Another User",
        "email": "another@example.com",
        "password": "password123",
        "role": "user"
    }')

if echo "$create_response" | grep -q "another@example.com"; then
    print_success "User created successfully"
    new_user_id=$(echo "$create_response" | grep -o '"id":[0-9]*' | sed 's/"id":\(.*\)/\1/')
    print_info "New user ID: $new_user_id"
else
    print_error "Failed to create user"
    echo "Response: $create_response"
fi
echo

# Test 8: Get specific user
if [ ! -z "$new_user_id" ]; then
    echo "Test 8: Getting user by ID..."
    get_user_response=$(curl -s "$BASE_URL/api/v1/users/$new_user_id" \
        -H "Authorization: Bearer $access_token")

    if echo "$get_user_response" | grep -q "another@example.com"; then
        print_success "User retrieved by ID successfully"
    else
        print_error "Failed to get user by ID"
        echo "Response: $get_user_response"
    fi
    echo
fi

# Test 9: Update user
if [ ! -z "$new_user_id" ]; then
    echo "Test 9: Updating user..."
    update_response=$(curl -s -X PUT "$BASE_URL/api/v1/users/$new_user_id" \
        -H "Authorization: Bearer $access_token" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $csrf_token" \
        -b "csrf_token=$csrf_token" \
        -d '{
            "name": "Updated User",
            "email": "updated@example.com",
            "role": "user"
        }')

    if echo "$update_response" | grep -q "Updated User"; then
        print_success "User updated successfully"
    else
        print_error "Failed to update user"
        echo "Response: $update_response"
    fi
    echo
fi

# Test 10: Delete user
if [ ! -z "$new_user_id" ]; then
    echo "Test 10: Deleting user..."
    delete_response=$(curl -s -X DELETE "$BASE_URL/api/v1/users/$new_user_id" \
        -H "Authorization: Bearer $access_token" \
        -H "X-CSRF-Token: $csrf_token" \
        -b "csrf_token=$csrf_token" \
        -w "\nHTTP_STATUS:%{http_code}")

    if echo "$delete_response" | grep -q "HTTP_STATUS:204"; then
        print_success "User deleted successfully"
    else
        print_error "Failed to delete user"
        echo "Response: $delete_response"
    fi
    echo
fi

echo "=========================================="
echo "All tests completed!"
echo "=========================================="
echo
echo "Additional manual tests:"
echo "  - Swagger UI: $BASE_URL/docs"
echo "  - Admin Panel: $BASE_URL/admin (requires admin role)"
echo
