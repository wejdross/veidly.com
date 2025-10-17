# GitHub Secrets Setup Tutorial

## Overview

This guide shows you how to configure all required GitHub secrets for automated deployment to Hetzner VPS using the `gh` CLI.

**IMPORTANT:** Veidly uses **Mailgun API** for sending emails, NOT SMTP!

## Prerequisites

### 1. Install GitHub CLI

```bash
# macOS
brew install gh

# Linux (Debian/Ubuntu)
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh
```

### 2. Authenticate

```bash
gh auth login
# Follow the prompts to authenticate with GitHub
```

### 3. Navigate to Repository

```bash
cd /Users/lukaszwidera/repos/veidly.com
```

---

## Required Secrets

| Secret | Description | Example |
|--------|-------------|---------|
| `HETZNER_SSH_KEY` | SSH private key for server access | Contents of `~/.ssh/veidly_deploy` |
| `HETZNER_HOST` | Server IP or hostname | `157.180.34.185` |
| `HETZNER_USER` | SSH username | `root` |
| `DOMAIN` | Your domain name | `veidly.com` |
| `JWT_SECRET` | Secret for JWT tokens (min 43 chars) | Generated with openssl |
| `MAILGUN_DOMAIN` | Your Mailgun domain | `mg.veidly.com` |
| `MAILGUN_API_KEY` | Mailgun API key | `key-xxxxxxxxxxxxx` |
| `MAILGUN_FROM_EMAIL` | Sender email address | `noreply@veidly.com` |
| `ADMIN_EMAIL` | Admin contact email | `admin@veidly.com` |

---

## Step-by-Step Setup

### Step 1: Generate SSH Key

```bash
# Generate deployment key (no passphrase for automation)
ssh-keygen -t ed25519 -f ~/.ssh/veidly_deploy -N "" -C "github-actions@veidly.com"

# Copy public key to your server
ssh-copy-id -i ~/.ssh/veidly_deploy.pub root@YOUR_SERVER_IP

# Test connection
ssh -i ~/.ssh/veidly_deploy root@YOUR_SERVER_IP "echo 'SSH connection successful!'"
```

### Step 2: Generate JWT Secret

```bash
# Generate a 64-character base64 JWT secret
openssl rand -base64 64 | tr -d '\n' > /tmp/jwt_secret.txt

# View it
cat /tmp/jwt_secret.txt
```

### Step 3: Setup Mailgun Account

Mailgun is the email service provider used by Veidly for sending verification emails and password resets.

#### Create Account

1. Go to https://www.mailgun.com/
2. Sign up for a free account (allows 5,000 emails/month)

#### Add and Verify Your Domain

1. Go to **Sending** → **Domains** → **Add New Domain**
2. Add subdomain: `mg.veidly.com` (recommended) or use `veidly.com`
3. Configure DNS records in your domain registrar:

   **Required DNS Records:**
   ```
   # TXT record for SPF
   Name: mg.veidly.com
   Type: TXT
   Value: v=spf1 include:mailgun.org ~all

   # TXT records for DKIM (Mailgun provides these)
   Name: k1._domainkey.mg.veidly.com
   Type: TXT
   Value: k=rsa; p=MIGfMA0GCSqGSI... (from Mailgun)

   # CNAME for tracking (optional)
   Name: email.mg.veidly.com
   Type: CNAME
   Value: mailgun.org

   # MX records (if using root domain)
   Name: mg.veidly.com
   Type: MX
   Priority: 10
   Value: mxa.mailgun.org

   Name: mg.veidly.com
   Type: MX
   Priority: 10
   Value: mxb.mailgun.org
   ```

4. Wait for DNS propagation (can take 24-48 hours)
5. Mailgun will verify your domain automatically

#### Get API Credentials

1. Go to **Settings** → **API Keys**
2. Copy your **Private API key** (starts with `key-`)
3. Your domain will be: `mg.veidly.com`

#### Alternative: Use Mailgun Sandbox (for testing)

Mailgun provides a sandbox domain immediately for testing:
- Format: `sandboxXXXXXXXXXXXXXXX.mailgun.org`
- ⚠️ Can only send to authorized recipients
- Add test recipients in **Sending** → **Authorized Recipients**

```bash
# Example sandbox domain
MAILGUN_DOMAIN="sandbox1234567890abcdef.mailgun.org"
```

### Step 4: Set All Secrets

```bash
# 1. Hetzner Server Access
gh secret set HETZNER_SSH_KEY < ~/.ssh/veidly_deploy
gh secret set HETZNER_HOST --body "157.180.34.185"  # Replace with your server IP
gh secret set HETZNER_USER --body "root"

# 2. Domain Configuration
gh secret set DOMAIN --body "veidly.com"

# 3. JWT Secret
gh secret set JWT_SECRET < /tmp/jwt_secret.txt

# 4. Mailgun Configuration (NOT SMTP!)
gh secret set MAILGUN_DOMAIN --body "mg.veidly.com"  # Your verified domain
gh secret set MAILGUN_API_KEY --body "key-xxxxxxxxxxxxxxxxxxxxxxxx"  # From Mailgun dashboard
gh secret set MAILGUN_FROM_EMAIL --body "noreply@veidly.com"

# 5. Admin Email
gh secret set ADMIN_EMAIL --body "admin@veidly.com"

# Clean up temporary file
rm /tmp/jwt_secret.txt
```

### Step 5: Verify Secrets

```bash
# List all secrets (values are hidden)
gh secret list

# Expected output:
# ADMIN_EMAIL         Updated 2025-XX-XX
# DOMAIN              Updated 2025-XX-XX
# HETZNER_HOST        Updated 2025-XX-XX
# HETZNER_SSH_KEY     Updated 2025-XX-XX
# HETZNER_USER        Updated 2025-XX-XX
# JWT_SECRET          Updated 2025-XX-XX
# MAILGUN_API_KEY     Updated 2025-XX-XX
# MAILGUN_DOMAIN      Updated 2025-XX-XX
# MAILGUN_FROM_EMAIL  Updated 2025-XX-XX
```

---

## Quick Setup Script

Save this as `setup-secrets.sh`:

```bash
#!/bin/bash
set -e

echo "🔐 Veidly GitHub Secrets Setup"
echo "================================"

# Check prerequisites
command -v gh >/dev/null 2>&1 || { echo "❌ gh CLI not installed. Install with: brew install gh"; exit 1; }
command -v ssh-keygen >/dev/null 2>&1 || { echo "❌ ssh-keygen not found"; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "❌ openssl not found"; exit 1; }

# Get user input
read -p "Enter your Hetzner server IP: " HETZNER_HOST
read -p "Enter your domain (e.g., veidly.com): " DOMAIN
read -p "Enter your Mailgun domain (e.g., mg.veidly.com): " MAILGUN_DOMAIN
read -p "Enter your Mailgun API key: " MAILGUN_API_KEY
read -p "Enter admin email: " ADMIN_EMAIL

# Generate SSH key if it doesn't exist
if [ ! -f ~/.ssh/veidly_deploy ]; then
    echo "📝 Generating SSH key..."
    ssh-keygen -t ed25519 -f ~/.ssh/veidly_deploy -N "" -C "github-actions@veidly.com"
    echo "✅ SSH key generated"
    echo ""
    echo "⚠️  IMPORTANT: Copy public key to server:"
    echo "   ssh-copy-id -i ~/.ssh/veidly_deploy.pub root@$HETZNER_HOST"
    read -p "Press Enter after copying the key..."
fi

# Test SSH connection
echo "🔌 Testing SSH connection..."
if ssh -i ~/.ssh/veidly_deploy -o StrictHostKeyChecking=no -o ConnectTimeout=5 root@$HETZNER_HOST "echo 'Success'" 2>/dev/null; then
    echo "✅ SSH connection successful"
else
    echo "❌ SSH connection failed. Check your server IP and key."
    exit 1
fi

# Generate JWT secret
echo "🔑 Generating JWT secret..."
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')

# Set all secrets
echo "📤 Setting GitHub secrets..."
gh secret set HETZNER_SSH_KEY < ~/.ssh/veidly_deploy
gh secret set HETZNER_HOST --body "$HETZNER_HOST"
gh secret set HETZNER_USER --body "root"
gh secret set DOMAIN --body "$DOMAIN"
gh secret set JWT_SECRET --body "$JWT_SECRET"
gh secret set MAILGUN_DOMAIN --body "$MAILGUN_DOMAIN"
gh secret set MAILGUN_API_KEY --body "$MAILGUN_API_KEY"
gh secret set MAILGUN_FROM_EMAIL --body "noreply@$DOMAIN"
gh secret set ADMIN_EMAIL --body "$ADMIN_EMAIL"

echo ""
echo "✅ All secrets configured!"
echo ""
gh secret list

echo ""
echo "🚀 Next Steps:"
echo "1. Verify Mailgun domain DNS records"
echo "2. Deploy: git tag -a v1.0.0 -m 'Release v1.0.0' && git push origin v1.0.0"
echo "3. Monitor: gh run list --workflow=deploy.yml"
```

Make it executable:

```bash
chmod +x setup-secrets.sh
./setup-secrets.sh
```

---

## Testing Deployment

### Trigger via Tag (Recommended)

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Manual Trigger

```bash
gh workflow run deploy.yml \
  --field environment=production \
  --field release_tag=main

# Check status
gh run list --workflow=deploy.yml
gh run view --log
```

---

## Troubleshooting

### Mailgun Emails Not Sending

1. **Check Domain Verification**
   ```bash
   dig TXT mg.veidly.com
   dig MX mg.veidly.com
   ```

2. **Test API Key**
   ```bash
   curl -s --user "api:YOUR_API_KEY" \
     https://api.mailgun.net/v3/domains/mg.veidly.com
   ```

3. **Send Test Email**
   ```bash
   curl -s --user "api:YOUR_API_KEY" \
     https://api.mailgun.net/v3/mg.veidly.com/messages \
     -F from="noreply@veidly.com" \
     -F to="test@example.com" \
     -F subject="Test" \
     -F text="Testing Mailgun"
   ```

4. **Check Backend Logs**
   ```bash
   ssh root@YOUR_SERVER "tail -f /var/log/veidly/backend.log | grep -i mail"
   ```

### SSH Connection Issues

```bash
# Test with verbose output
ssh -v -i ~/.ssh/veidly_deploy root@YOUR_SERVER_IP

# Fix permissions
chmod 600 ~/.ssh/veidly_deploy
chmod 644 ~/.ssh/veidly_deploy.pub
```

---

## Complete Example

```bash
# 1. Generate keys
ssh-keygen -t ed25519 -f ~/.ssh/veidly_deploy -N ""
JWT=$(openssl rand -base64 64 | tr -d '\n')

# 2. Copy SSH key to server
ssh-copy-id -i ~/.ssh/veidly_deploy.pub root@157.180.34.185

# 3. Set all secrets
gh secret set HETZNER_SSH_KEY < ~/.ssh/veidly_deploy
gh secret set HETZNER_HOST --body "157.180.34.185"
gh secret set HETZNER_USER --body "root"
gh secret set DOMAIN --body "veidly.com"
gh secret set JWT_SECRET --body "$JWT"
gh secret set MAILGUN_DOMAIN --body "mg.veidly.com"
gh secret set MAILGUN_API_KEY --body "key-1234567890abcdef"
gh secret set MAILGUN_FROM_EMAIL --body "noreply@veidly.com"
gh secret set ADMIN_EMAIL --body "admin@veidly.com"

# 4. Verify
gh secret list

# 5. Deploy
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

---

**You're now ready to deploy Veidly! 🚀**

For support, check:
- Workflow logs: `gh run view --log`
- Backend logs: `ssh root@YOUR_SERVER tail -f /var/log/veidly/backend.log`
- Mailgun dashboard: https://app.mailgun.com/
