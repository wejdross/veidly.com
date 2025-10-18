#!/bin/bash
#
# SSH Key Fixer for GitHub Actions
# This script helps you create a properly formatted SSH key for GitHub Actions
#

set -e

echo "🔧 SSH Key Fixer for GitHub Actions"
echo "===================================="
echo ""

KEY_PATH="${1:-$HOME/.ssh/veidly_deploy}"
SERVER="${2:-157.180.34.185}"

echo "Configuration:"
echo "  Key path: $KEY_PATH"
echo "  Server: $SERVER"
echo ""

# Check if key already exists
if [ -f "$KEY_PATH" ]; then
    echo "⚠️  Key already exists at $KEY_PATH"
    read -p "Do you want to overwrite it? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting."
        exit 1
    fi
    rm -f "$KEY_PATH" "$KEY_PATH.pub"
fi

# Generate new key without passphrase
echo "📝 Generating new SSH key..."
ssh-keygen -t ed25519 -f "$KEY_PATH" -N "" -C "github-actions-deploy"

echo ""
echo "✅ SSH key generated successfully!"
echo ""

# Validate key format
echo "🔍 Validating key format..."
if ssh-keygen -l -f "$KEY_PATH" > /dev/null 2>&1; then
    echo "✅ Key format is valid"
else
    echo "❌ Key format is invalid"
    exit 1
fi

# Check if key is encrypted
if grep -q "ENCRYPTED" "$KEY_PATH"; then
    echo "❌ Key is encrypted (has passphrase) - this won't work with GitHub Actions"
    exit 1
else
    echo "✅ Key is not encrypted"
fi

echo ""
echo "📋 Next steps:"
echo ""
echo "1️⃣  Copy the PRIVATE key to GitHub Secret 'HETZNER_SSH_KEY':"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat "$KEY_PATH"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "2️⃣  Add the PUBLIC key to your server:"
echo ""

# Offer to copy public key to server
read -p "Do you want to copy the public key to $SERVER now? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Copying public key to $SERVER..."

    if ssh-copy-id -i "$KEY_PATH.pub" "root@$SERVER"; then
        echo "✅ Public key copied successfully!"
        echo ""
        echo "Testing SSH connection..."
        if ssh -i "$KEY_PATH" "root@$SERVER" "echo 'SSH connection works!' && exit"; then
            echo "✅ SSH connection successful!"
        else
            echo "❌ SSH connection failed"
            exit 1
        fi
    else
        echo "❌ Failed to copy public key"
        echo ""
        echo "Manual steps:"
        echo "1. Copy this public key:"
        echo ""
        cat "$KEY_PATH.pub"
        echo ""
        echo "2. SSH to your server: ssh root@$SERVER"
        echo "3. Add it to authorized_keys: echo 'PASTE_PUBLIC_KEY_HERE' >> ~/.ssh/authorized_keys"
    fi
else
    echo ""
    echo "Manual setup:"
    echo ""
    echo "Run this command to copy the key to your server:"
    echo "  ssh-copy-id -i $KEY_PATH.pub root@$SERVER"
    echo ""
    echo "Or manually:"
    echo "  cat $KEY_PATH.pub | ssh root@$SERVER \"mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys\""
fi

echo ""
echo "3️⃣  Test locally before using in GitHub Actions:"
echo "  ssh -i $KEY_PATH root@$SERVER"
echo ""
echo "4️⃣  Add to GitHub:"
echo "  - Go to: https://github.com/wejdross/veidly.com/settings/secrets/actions"
echo "  - Click 'New repository secret'"
echo "  - Name: HETZNER_SSH_KEY"
echo "  - Secret: Paste the ENTIRE private key (from step 1)"
echo ""
echo "5️⃣  Test GitHub Actions:"
echo "  - Go to Actions tab"
echo "  - Click 'Deploy Veidly'"
echo "  - Click 'Run workflow'"
echo "  - Select branch and environment"
echo "  - Click 'Run workflow'"
echo ""
echo "✅ Setup complete!"
