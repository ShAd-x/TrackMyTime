#!/bin/bash

# TrackMyTime Dashboard Test Script
# Tests all API endpoints and verifies web dashboard

echo "🧪 TrackMyTime Dashboard Test Suite v1.2.0"
echo "==========================================="
echo ""

API_BASE="http://localhost:8787"
PASS=0
FAIL=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local name=$1
    local endpoint=$2
    local expected_code=${3:-200}
    
    echo -n "Testing $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE$endpoint" 2>/dev/null)
    
    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $response)"
        ((PASS++))
    else
        echo -e "${RED}✗ FAIL${NC} (Expected HTTP $expected_code, got $response)"
        ((FAIL++))
    fi
}

# Check if agent is running
echo "🔍 Checking if TrackMyTime agent is running..."
if ! curl -s "$API_BASE/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ ERROR: Agent is not running!${NC}"
    echo "Please start the agent with: ./trackmytime"
    exit 1
fi
echo -e "${GREEN}✓ Agent is running${NC}"
echo ""

# Test Web Dashboard
echo "📊 Testing Web Dashboard"
echo "------------------------"
test_endpoint "Dashboard HTML" "/"
test_endpoint "CSS Stylesheet" "/static/css/style.css"
test_endpoint "JavaScript" "/static/js/app.js"
echo ""

# Test API Endpoints
echo "🔌 Testing API Endpoints"
echo "------------------------"
test_endpoint "Health Check" "/health"
test_endpoint "Stats Today" "/stats/today"
test_endpoint "Stats Week" "/stats/week"
test_endpoint "Current Activity" "/activity/current"
test_endpoint "Export CSV (today)" "/export/csv?period=today"
test_endpoint "Export Aggregated CSV" "/export/aggregated?period=today&format=csv"
test_endpoint "Export Aggregated JSON" "/export/aggregated?period=today&format=json"
test_endpoint "Hourly Stats (today)" "/api/stats/hourly?period=today"
test_endpoint "Hourly Stats (week)" "/api/stats/hourly?period=week"
echo ""

# Test API Response Content
echo "📝 Testing API Response Content"
echo "--------------------------------"

# Test health endpoint
echo -n "Health endpoint returns JSON... "
health_response=$(curl -s "$API_BASE/health")
if echo "$health_response" | grep -q '"status":"ok"'; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

# Test stats endpoint
echo -n "Stats endpoint has required fields... "
stats_response=$(curl -s "$API_BASE/stats/today")
if echo "$stats_response" | grep -q '"stats_by_app"' && echo "$stats_response" | grep -q '"total_active_seconds"'; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

# Test hourly endpoint
echo -n "Hourly endpoint returns array of 24... "
hourly_response=$(curl -s "$API_BASE/api/stats/hourly?period=today")
hourly_count=$(echo "$hourly_response" | grep -o ',' | wc -l)
if [ "$hourly_count" -ge 23 ]; then  # 24 elements = 23 commas
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

echo ""

# Test Dashboard Components
echo "🎨 Testing Dashboard Components"
echo "--------------------------------"

dashboard_html=$(curl -s "$API_BASE/")

echo -n "Dashboard contains Chart.js... "
if echo "$dashboard_html" | grep -q "chart.js"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

echo -n "Dashboard has donut chart canvas... "
if echo "$dashboard_html" | grep -q "donut-chart"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

echo -n "Dashboard has timeline chart canvas... "
if echo "$dashboard_html" | grep -q "timeline-chart"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

echo -n "Dashboard has top apps table... "
if echo "$dashboard_html" | grep -q "apps-table"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((FAIL++))
fi

echo ""

# Summary
echo "=========================================="
echo "📊 Test Results"
echo "=========================================="
echo -e "Total Tests: $((PASS + FAIL))"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✨ All tests passed! Dashboard is ready to use.${NC}"
    echo ""
    echo "🌐 Open your browser to:"
    echo "   http://localhost:8787/"
    echo ""
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed. Please check the output above.${NC}"
    exit 1
fi
