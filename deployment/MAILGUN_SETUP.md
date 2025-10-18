# Mailgun Configuration Guide

This guide will help you configure Mailgun for the Veidly application.

## Prerequisites

1. Mailgun account created
2. Domain `mg.veidly.com` added to Mailgun
3. API key from Mailgun dashboard

## Configuration Steps

### 1. Get Your Mailgun API Key

1. Log in to [Mailgun Dashboard](https://app.mailgun.com)
2. Go to **Settings → API Keys**
3. Copy your **Private API Key** (starts with `key-`)

### 2. Verify Your Domain (Required for Production)

For **sandbox domains** (like `mg.veidly.com`), you can send emails to authorized recipients only.

**Option A: Keep Sandbox Domain (Testing)**
- Go to: https://app.mailgun.com/app/sending/domains/mg.veidly.com/recipients
- Add authorized recipient emails (e.g., `lukaszwidera1993@gmail.com`)
- These are the only emails that can receive messages

**Option B: Verify Domain (Production)**
1. Go to: https://app.mailgun.com/app/sending/domains/mg.veidly.com/verify
2. Add the required DNS records to your domain:
   - **TXT records** for domain verification
   - **MX records** for receiving emails
   - **CNAME records** for tracking
3. Wait for DNS propagation (can take up to 48 hours)
4. Click "Verify DNS Settings" in Mailgun dashboard

### 3. Update Deployment Configuration

Edit `deployment/inventory` file:

```ini
[veidly_servers:vars]
# ... other settings ...

# Mailgun Configuration
mailgun_domain=mg.veidly.com
mailgun_api_key=key-YOUR_ACTUAL_API_KEY_HERE    # ⚠️ Replace this!
mailgun_from_email=postmaster@mg.veidly.com     # ⚠️ Must match sandbox domain
```

**Important Notes:**
- ✅ Use `mg.veidly.com` as the domain (your Mailgun sandbox)
- ✅ Use `postmaster@mg.veidly.com` as from email (required for sandbox)
- ✅ Replace `key-YOUR_ACTUAL_API_KEY_HERE` with your real API key
- ⚠️ **Never commit the actual API key to git!** (inventory file is gitignored)

### 4. Test Configuration Locally

Before deploying, test your Mailgun configuration:

```bash
cd backend

# Set your API key
export MAILGUN_API_KEY="key-your-actual-api-key"

# Optional: specify test email (defaults to lukaszwidera1993@gmail.com)
export TEST_EMAIL="your-email@example.com"

# Run test script
./test_mailgun.sh
```

Expected output:
```
✅ Domain is active (or ⚠️ Domain is unverified)
✅ Email sent successfully!
📬 Check your inbox at: your-email@example.com
```

### 5. Test with Your Go Code

You can also test using your own Go code:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/mailgun/mailgun-go/v4"
)

func main() {
    domain := "mg.veidly.com"
    apiKey := os.Getenv("MAILGUN_API_KEY")

    mg := mailgun.NewMailgun(domain, apiKey)
    mg.SetAPIBase("https://api.eu.mailgun.net/v3")

    m := mg.NewMessage(
        "Veidly <postmaster@mg.veidly.com>",
        "Test from Veidly",
        "This is a test email!",
        "lukaszwidera1993@gmail.com",
    )

    ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
    defer cancel()

    resp, id, err := mg.Send(ctx, m)
    fmt.Printf("Response: %s\nID: %s\nError: %v\n", resp, id, err)
}
```

Run with:
```bash
MAILGUN_API_KEY="key-your-key" go run your_test.go
```

### 6. Deploy to Production

Once tested locally:

```bash
cd deployment

# Make sure inventory has correct API key
ansible-playbook -i inventory deploy.yml
```

The deployment will:
1. Pull latest code (with EU API endpoint fix)
2. Set environment variables from inventory
3. Restart the backend service
4. Email service will be initialized with EU endpoint

### 7. Verify Emails Work in Production

After deployment, test the email features:

1. **Register a new account** on https://veidly.com
   - Should receive verification email

2. **Request password reset**
   - Should receive password reset email

3. **Check backend logs** for email confirmations:
   ```bash
   ansible -i inventory veidly_servers -m shell -a "tail -50 /var/log/veidly/backend.error.log | grep -i email" --become
   ```

Expected log messages:
```
✓ Email service initialized for domain: mg.veidly.com (EU endpoint)
✓ Verification email sent to user@example.com
✓ Password reset email sent to user@example.com
```

## Troubleshooting

### Error: "Domain not verified"
- Add DNS records from Mailgun dashboard
- Wait for DNS propagation
- Use sandbox mode with authorized recipients during testing

### Error: "Forbidden - Can't send to this recipient"
- For sandbox domains, add recipient to authorized list
- Or verify your domain for production use

### Error: "Invalid private key"
- Check that API key starts with `key-`
- Verify you're using the **Private API Key**, not Public Key
- Check for extra spaces or quotes in inventory file

### Error: "Connection timeout"
- Verify EU endpoint: `https://api.eu.mailgun.net/v3`
- Check firewall rules allow outbound HTTPS
- Verify server can reach Mailgun API

### No emails received
- Check spam folder
- Verify recipient is authorized (for sandbox domains)
- Check Mailgun logs: https://app.mailgun.com/app/logs
- Check backend logs for email sending confirmation

## Email Templates

The backend sends three types of emails:

1. **Verification Email** - Sent when user registers
2. **Password Reset Email** - Sent when user requests password reset
3. **Welcome Email** - Sent after email verification

All templates include:
- Professional HTML design
- Plain text fallback
- Secure token-based links
- Expiration times (24h for verification, 1h for password reset)

## Security Notes

- ✅ API key stored in inventory file (gitignored)
- ✅ Emails sent over HTTPS
- ✅ 30-second timeout prevents hanging
- ✅ Tokens are cryptographically secure (32 bytes)
- ⚠️ Use verified domain in production (not sandbox)
- ⚠️ Never commit API keys to version control

## Production Checklist

Before going live:

- [ ] Mailgun domain fully verified (all DNS records added)
- [ ] API key configured in production inventory
- [ ] Test registration email flow
- [ ] Test password reset email flow
- [ ] Monitor Mailgun dashboard for delivery rates
- [ ] Set up email bounce handling (optional)
- [ ] Configure DKIM and SPF for better deliverability

## Costs

Mailgun Free Tier:
- 5,000 emails/month for 3 months (trial)
- After trial: Pay as you go ($0.80/1000 emails)
- EU region supported

## Support

- Mailgun Documentation: https://documentation.mailgun.com/
- Mailgun Support: https://help.mailgun.com/
- Veidly Email Code: `backend/email.go`
