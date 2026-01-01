#!/bin/bash

# Test script for JWT verification
# This script tests the JWT verification flow between auth-service and score-service

set -e

AUTH_URL="${AUTH_SERVICE_URL:-http://localhost:3000}"
SCORE_URL="${SCORE_SERVICE_URL:-http://localhost:1323}"

echo "🧪 Testing JWT Verification"
echo "============================"
echo "Auth Service URL: $AUTH_URL"
echo "Score Service URL: $SCORE_URL"
echo ""

# Test 1: Check if JWKS endpoint is accessible
echo "1️⃣  Testing JWKS endpoint..."
JWKS_RESPONSE=$(curl -s "$AUTH_URL/api/auth/jwks" || echo "")
if [ -z "$JWKS_RESPONSE" ]; then
    echo "❌ JWKS endpoint is not accessible. Is the auth service running?"
    echo "   Start it with: cd auth-service && bun run dev"
    exit 1
fi

# Check if response contains "keys" (basic JSON validation)
if echo "$JWKS_RESPONSE" | grep -q '"keys"'; then
    echo "✅ JWKS endpoint is accessible and returns valid JSON"
    if command -v jq > /dev/null 2>&1; then
        echo "$JWKS_RESPONSE" | jq '.'
    else
        echo "$JWKS_RESPONSE"
    fi
else
    echo "❌ JWKS endpoint returned invalid JSON"
    echo "Response: $JWKS_RESPONSE"
    exit 1
fi

echo ""
echo "2️⃣  Testing score service health..."
SCORE_HEALTH=$(curl -s "$SCORE_URL/" || echo "")
if [ -z "$SCORE_HEALTH" ]; then
    echo "❌ Score service is not accessible. Is it running?"
    echo "   Start it with: cd score-service && go run ."
    exit 1
fi
echo "✅ Score service is accessible: $SCORE_HEALTH"

echo ""
echo "3️⃣  Testing protected endpoint without token..."
PROTECTED_RESPONSE=$(curl -s -w "\n%{http_code}" "$SCORE_URL/api/score" || echo "")
HTTP_CODE=$(echo "$PROTECTED_RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    echo "✅ Protected endpoint correctly returns 401 without token"
else
    echo "❌ Expected 401, got $HTTP_CODE"
    exit 1
fi

echo ""
echo "4️⃣  Testing protected endpoint with invalid token..."
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer invalid-token-12345" "$SCORE_URL/api/score" || echo "")
HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    echo "✅ Protected endpoint correctly returns 401 with invalid token"
else
    echo "❌ Expected 401, got $HTTP_CODE"
    exit 1
fi

echo ""
echo "✅ All tests passed!"
echo ""
echo "📝 To test with a real token:"
echo "   1. Sign up and sign in at $AUTH_URL/api/auth/sign-in"
echo "   2. Get a token from $AUTH_URL/api/auth/token (requires session cookie)"
echo "   3. Test with: curl -H \"Authorization: Bearer <token>\" $SCORE_URL/api/score"

