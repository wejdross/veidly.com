#!/bin/bash
#
# Mailgun Configuration Test Script
# This script helps you test your Mailgun configuration before deploying
#

set -e

echo "🔍 Mailgun Configuration Test"
echo "========================================"
echo ""

# Check if required variables are provided
if [ -z "$MAILGUN_API_KEY" ]; then
    echo "❌ Error: MAILGUN_API_KEY environment variable not set"
    echo ""
    echo "Usage: MAILGUN_API_KEY=your-key ./test_mailgun.sh"
    exit 1
fi

# Configuration
DOMAIN="mg.veidly.com"
FROM_EMAIL="postmaster@mg.veidly.com"
TO_EMAIL="${TEST_EMAIL:-lukaszwidera1993@gmail.com}"
API_BASE="https://api.eu.mailgun.net/v3"

echo "📧 Configuration:"
echo "   Domain: $DOMAIN"
echo "   From: $FROM_EMAIL"
echo "   To: $TO_EMAIL"
echo "   API Base: $API_BASE"
echo ""

# Test 1: Verify domain
echo "🔍 Test 1: Verifying domain configuration..."
DOMAIN_INFO=$(curl -s --user "api:$MAILGUN_API_KEY" \
    "$API_BASE/domains/$DOMAIN")

if echo "$DOMAIN_INFO" | grep -q '"state":"active"'; then
    echo "   ✅ Domain is active"
elif echo "$DOMAIN_INFO" | grep -q '"state":"unverified"'; then
    echo "   ⚠️  Domain is unverified - you need to add DNS records"
    echo "   Check: https://app.mailgun.com/app/sending/domains/$DOMAIN/verify"
else
    echo "   ❌ Domain check failed"
    echo "   Response: $DOMAIN_INFO"
fi
echo ""

# Test 2: Send test email
echo "📨 Test 2: Sending test email..."
RESPONSE=$(curl -s --user "api:$MAILGUN_API_KEY" \
    "$API_BASE/$DOMAIN/messages" \
    -F from="Veidly Test <$FROM_EMAIL>" \
    -F to="$TO_EMAIL" \
    -F subject="Veidly Mailgun Configuration Test" \
    -F text="This is a test email from Veidly. If you received this, your Mailgun configuration is working correctly!")

if echo "$RESPONSE" | grep -q '"id"'; then
    MESSAGE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "   ✅ Email sent successfully!"
    echo "   Message ID: $MESSAGE_ID"
    echo ""
    echo "   📬 Check your inbox at: $TO_EMAIL"
else
    echo "   ❌ Email sending failed"
    echo "   Response: $RESPONSE"
fi
echo ""

# Test 3: Check authorized recipients (sandbox limitation)
echo "🔍 Test 3: Checking authorized recipients (sandbox only)..."
RECIPIENTS=$(curl -s --user "api:$MAILGUN_API_KEY" \
    "$API_BASE/$DOMAIN/bounces")

if echo "$RECIPIENTS" | grep -q "lukaszwidera1993@gmail.com"; then
    echo "   ✅ lukaszwidera1993@gmail.com is authorized"
else
    echo "   ⚠️  Note: Sandbox domains can only send to authorized recipients"
    echo "   Add recipients here: https://app.mailgun.com/app/sending/domains/$DOMAIN/recipients"
fi
echo ""

echo "========================================"
echo "✅ Configuration test complete!"
echo ""
echo "Next steps:"
echo "1. If domain is unverified, add DNS records from Mailgun dashboard"
echo "2. For sandbox domains, authorize recipient emails"
echo "3. Update deployment/inventory with your API key"
echo "4. Deploy and test in production"
