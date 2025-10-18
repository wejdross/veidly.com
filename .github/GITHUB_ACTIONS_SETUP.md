# GitHub Actions Deployment Setup Guide

This guide will help you configure GitHub Actions secrets for automated deployment to your Hetzner VPS.

## SSH Key Setup

The most common issue is SSH key format. Follow these steps carefully:

### Option 1: Generate a New Unencrypted Key (Recommended)

```bash
# Generate a new ED25519 key WITHOUT passphrase
ssh-keygen -t ed25519 -f ~/.ssh/veidly_deploy -N ""

# Display the private key (this goes in GitHub secret)
cat ~/.ssh/veidly_deploy

# Display the public key (this goes on your server)
cat ~/.ssh/veidly_deploy.pub
```

### Option 2: Convert Existing Encrypted Key

If you have an existing key with a passphrase:

```bash
# Remove passphrase from existing key
ssh-keygen -p -f ~/.ssh/id_rsa -N ""

# Or create a copy without passphrase
cp ~/.ssh/id_rsa ~/.ssh/veidly_deploy
ssh-keygen -p -f ~/.ssh/veidly_deploy -N ""
```

### Option 3: Convert New OpenSSH Format to PEM

If your key was generated with newer OpenSSH (starts with `-----BEGIN OPENSSH PRIVATE KEY-----`):

```bash
# Convert to PEM format
ssh-keygen -p -f ~/.ssh/id_rsa -m PEM -N ""
```

## Add Public Key to Server

```bash
# Copy public key to your Hetzner server
ssh-copy-id -i ~/.ssh/veidly_deploy.pub root@YOUR_SERVER_IP

# Or manually:
cat ~/.ssh/veidly_deploy.pub | ssh root@YOUR_SERVER_IP "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"

# Test SSH connection
ssh -i ~/.ssh/veidly_deploy root@YOUR_SERVER_IP
```

## GitHub Secrets Configuration

Go to your GitHub repository → Settings → Secrets and variables → Actions → New repository secret

### Required Secrets

| Secret Name | Description | Example Value |
|------------|-------------|---------------|
| `HETZNER_SSH_KEY` | Private SSH key (entire content) | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `HETZNER_HOST` | Server IP address | `157.180.34.185` |
| `HETZNER_USER` | SSH username (optional) | `root` (default) |
| `DOMAIN` | Your domain name (optional) | `veidly.com` (default) |
| `JWT_SECRET` | Random string for JWT | `your-secret-key-here` |
| `MAILGUN_DOMAIN` | Mailgun domain | `mg.veidly.com` |
| `MAILGUN_API_KEY` | Mailgun API key | `key-xxxxxxxxxxxxxxxx` |
| `MAILGUN_FROM_EMAIL` | From email address (optional) | `postmaster@mg.veidly.com` (default) |
| `ADMIN_EMAIL` | Admin email | `admin@veidly.com` |
| `ADMIN_PASSWORD` | Admin password | `your-secure-password` |

### How to Add HETZNER_SSH_KEY Secret

1. Copy the **entire private key** including headers:
   ```bash
   cat ~/.ssh/veidly_deploy
   ```

2. The output should look like:
   ```
   -----BEGIN OPENSSH PRIVATE KEY-----
   b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
   ... (many lines) ...
   -----END OPENSSH PRIVATE KEY-----
   ```

3. Copy **everything** from `-----BEGIN` to `-----END`

4. In GitHub:
   - Go to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `HETZNER_SSH_KEY`
   - Secret: Paste the entire key
   - Click "Add secret"

### Troubleshooting SSH Key Issues

If you see this error:
```
Load key "/home/runner/.ssh/deploy_key": error in libcrypto
Permission denied (publickey,password).
```

**Causes:**
1. ✅ Key has a passphrase (must be unencrypted)
2. ✅ Wrong key format (use PEM or standard OpenSSH)
3. ✅ Copied public key instead of private key
4. ✅ Public key not added to server's authorized_keys

**Solutions:**

**Check 1: Is your key encrypted?**
```bash
grep -q "ENCRYPTED" ~/.ssh/id_rsa && echo "Key is encrypted ❌" || echo "Key is not encrypted ✅"
```

**Check 2: Verify key format**
```bash
ssh-keygen -l -f ~/.ssh/id_rsa
# Should output key fingerprint, not an error
```

**Check 3: Test SSH locally**
```bash
ssh -i ~/.ssh/veidly_deploy -v root@YOUR_SERVER_IP
# Should connect without asking for password
```

**Check 4: Verify public key on server**
```bash
ssh root@YOUR_SERVER_IP "cat ~/.ssh/authorized_keys | grep -q 'YOUR_KEY_COMMENT' && echo 'Key found ✅' || echo 'Key not found ❌'"
```

## Generate JWT Secret

```bash
# Generate a secure random JWT secret
openssl rand -base64 32
```

Copy the output and add it as `JWT_SECRET` secret in GitHub.

## Testing the Workflow

### Option 1: Manual Workflow Trigger

1. Go to Actions → Deploy Veidly → Run workflow
2. Select branch: `master`
3. Environment: `production`
4. Release tag: leave empty (will use current code)
5. Click "Run workflow"

### Option 2: Create a Git Tag

```bash
# Create and push a new version tag
git tag v0.4.0
git push origin v0.4.0

# This will automatically trigger the deployment workflow
```

## Workflow Behavior

The deployment workflow will:

1. ✅ Run backend tests
2. ✅ Run frontend tests
3. ✅ Create GitHub release (if triggered by tag)
4. ✅ Deploy to Hetzner VPS using Ansible:
   - Backup database before deployment
   - Pull latest code
   - Build backend and frontend
   - Update systemd services
   - Configure nginx with SSL
5. ✅ Verify deployment (check API health)
6. ✅ Rollback automatically if deployment fails

## Monitoring Deployments

### View Workflow Runs

- GitHub → Actions → Click on the workflow run
- See real-time logs for each step
- Check deployment summary at the bottom

### Check Deployment on Server

```bash
# SSH to server
ssh -i ~/.ssh/veidly_deploy root@YOUR_SERVER_IP

# Check backend service
systemctl status veidly

# Check backend logs
tail -50 /var/log/veidly/backend.error.log

# Check nginx
systemctl status nginx

# View recent deployments
ls -lh /home/veidly/backups/
```

### API Health Check

```bash
# Check API is responding
curl -I https://veidly.com/api/events

# Should return: HTTP/2 200
```

## Rollback Procedure

If deployment fails, the workflow automatically attempts rollback to the previous version.

### Manual Rollback

```bash
# Option 1: Via GitHub Actions
# Go to Actions → Deploy Veidly → Run workflow
# Select previous release tag (e.g., v0.3.0)

# Option 2: Via Ansible locally
cd deployment
ansible-playbook -i inventory deploy.yml --extra-vars "release_tag=v0.3.0"

# Option 3: Restore database backup manually
ssh root@YOUR_SERVER_IP
/home/veidly/backup-db.sh restore
```

## Security Best Practices

1. ✅ **Never commit** the `inventory` file with real secrets
2. ✅ **Use GitHub Secrets** for all sensitive data
3. ✅ **Rotate secrets** regularly (especially JWT_SECRET)
4. ✅ **Use unencrypted SSH keys** for CI/CD (store securely in GitHub)
5. ✅ **Enable 2FA** on GitHub account
6. ✅ **Limit deploy key permissions** (use deploy keys, not personal SSH keys)
7. ✅ **Monitor deployment logs** for suspicious activity

## Environment-Specific Deployments

### Production
- Triggered by: version tags (e.g., `v1.0.0`)
- Domain: `veidly.com`
- Uses: `HETZNER_*` secrets

### Staging
- Triggered by: RC tags (e.g., `v1.0.0-rc.1`) or manual workflow
- Domain: `staging.veidly.com`
- Uses: `STAGING_*` secrets

## Common Issues

### Issue: "Permission denied (publickey)"
**Solution**: Verify public key is in server's authorized_keys

### Issue: "error in libcrypto"
**Solution**: Remove passphrase from SSH key or convert to PEM format

### Issue: "Deployment failed - API health check failed"
**Solution**: Check backend logs, verify database, check nginx configuration

### Issue: "Failed to connect to the host"
**Solution**: Verify `HETZNER_HOST` secret is correct IP address

### Issue: "Module 'ping' not found"
**Solution**: Python3 should be installed on server (Ansible requirement)

## Next Steps

1. ✅ Add all required secrets to GitHub
2. ✅ Test SSH connection locally
3. ✅ Run workflow manually to test
4. ✅ Monitor first deployment carefully
5. ✅ Set up monitoring/alerting (optional)

## Support

- GitHub Actions Docs: https://docs.github.com/en/actions
- Ansible Docs: https://docs.ansible.com/
- Issue Tracker: https://github.com/wejdross/veidly.com/issues
