# Veidly Deployment Guide

Automated deployment to Hetzner VPS using Ansible.

## Prerequisites

- **Hetzner VPS** with Ubuntu 22.04 LTS
- **Domain** pointed to your VPS IP address
- **Ansible** installed on your local machine (`pip install ansible`)
- **GitHub repository** with tagged releases
- **SMTP credentials** for email functionality

## Quick Start

### 1. Configure Inventory

```bash
cp inventory.example inventory
```

Edit `inventory` file with your server details:

```ini
[veidly_servers]
veidly-prod ansible_host=YOUR_SERVER_IP ansible_user=root

[veidly_servers:vars]
domain=yourdomain.com
release_tag=v1.0.0
github_repo=yourusername/veidly.com
jwt_secret=CHANGE_TO_RANDOM_64_CHAR_STRING
smtp_host=smtp.example.com
smtp_port=587
smtp_user=noreply@yourdomain.com
smtp_pass=YOUR_SMTP_PASSWORD
admin_email=admin@yourdomain.com
```

### 2. Run Deployment

```bash
ansible-playbook -i inventory deploy.yml
```

Or with specific release tag:

```bash
ansible-playbook -i inventory deploy.yml --extra-vars "release_tag=v1.0.1"
```

## What Gets Deployed

### System Setup
- ✅ System packages (Git, SQLite, etc.)
- ✅ Go 1.21+ compiler
- ✅ Node.js 20.x
- ✅ Nginx web server
- ✅ Certbot for Let's Encrypt

### Application
- ✅ Fetch specific GitHub release tag
- ✅ Build backend (Go)
- ✅ Build frontend (React + Vite)
- ✅ Systemd service configuration
- ✅ Environment variables

### Security & SSL
- ✅ Let's Encrypt SSL certificates
- ✅ Auto-renewal cron job
- ✅ HTTPS redirect
- ✅ Security headers

### Database & Backups
- ✅ SQLite database setup
- ✅ **Pre-deployment backup**
- ✅ Daily automated backups (2 AM)
- ✅ 30-day backup retention

### Logging
- ✅ Application logs preserved in `/var/log/veidly/`
- ✅ Automatic log rotation (30 days)
- ✅ Separate access and error logs

## Directory Structure (on server)

```
/home/veidly/
├── app/                    # Application code
│   ├── backend/
│   │   └── veidly-backend  # Compiled binary
│   └── frontend/
│       └── dist/           # Built static files
├── backups/                # Database backups
│   └── veidly_*.db.gz
└── backup-db.sh            # Backup script

/var/lib/veidly/
└── veidly.db              # SQLite database

/var/log/veidly/
├── backend.log            # Application logs
├── backend.error.log      # Error logs
├── nginx-access.log       # Nginx access logs
└── nginx-error.log        # Nginx error logs
```

## Post-Deployment Tasks

### 1. Verify Deployment

```bash
# Check backend service
ssh your-server "sudo systemctl status veidly"

# Check nginx
ssh your-server "sudo systemctl status nginx"

# Test API
curl https://yourdomain.com/api/events
```

### 2. Create Admin User

```bash
ssh your-server
sqlite3 /var/lib/veidly/veidly.db
UPDATE users SET is_admin = 1 WHERE email = 'admin@yourdomain.com';
.quit
```

### 3. Monitor Logs

```bash
# Follow backend logs
ssh your-server "sudo journalctl -u veidly -f"

# Check error logs
ssh your-server "tail -f /var/log/veidly/backend.error.log"
```

## Backup & Restore

### Manual Backup

```bash
ssh your-server "/home/veidly/backup-db.sh"
```

### Restore from Backup

```bash
# List available backups
ssh your-server "ls -lh /home/veidly/backups/"

# Restore specific backup
ssh your-server
sudo systemctl stop veidly
gunzip -c /home/veidly/backups/veidly_20241017_020000.db.gz > /var/lib/veidly/veidly.db
sudo systemctl start veidly
```

## Updating the Application

### Deploy New Release

```bash
ansible-playbook -i inventory deploy.yml --extra-vars "release_tag=v1.1.0"
```

This will:
1. ✅ Backup database automatically
2. ✅ Fetch new release from GitHub
3. ✅ Build backend and frontend
4. ✅ Restart service with zero downtime

### Rollback

If something goes wrong, rollback to previous release:

```bash
# Deploy previous version
ansible-playbook -i inventory deploy.yml --extra-vars "release_tag=v1.0.0"

# Or manually restore database backup
ssh your-server
sudo systemctl stop veidly
gunzip -c /home/veidly/backups/veidly_pre_deploy_*.db.gz > /var/lib/veidly/veidly.db
sudo systemctl start veidly
```

## Troubleshooting

### Backend Not Starting

```bash
# Check service status
sudo systemctl status veidly

# View logs
sudo journalctl -u veidly -n 100 --no-pager

# Check database permissions
ls -l /var/lib/veidly/veidly.db
```

### SSL Certificate Issues

```bash
# Renew certificate manually
sudo certbot renew --force-renewal

# Check certificate status
sudo certbot certificates
```

### Frontend Not Loading

```bash
# Check nginx configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx

# Check build output
ls -la /home/veidly/app/frontend/dist/
```

## Security Recommendations

1. **Change default JWT secret** in inventory file
2. **Use strong SMTP password**
3. **Enable firewall**:
   ```bash
   ufw allow 22/tcp
   ufw allow 80/tcp
   ufw allow 443/tcp
   ufw enable
   ```
4. **Regular updates**:
   ```bash
   apt update && apt upgrade -y
   ```
5. **Monitor logs** for suspicious activity

## Maintenance

### Daily (Automated)
- Database backup (2 AM)
- Log rotation (weekly)

### Weekly
- Check disk space
- Review error logs
- Monitor backup success

### Monthly
- Update system packages
- Review SSL certificate renewal
- Check application performance

## Support

For issues or questions:
- Check logs: `/var/log/veidly/`
- Review systemd status: `systemctl status veidly`
- Consult Antora documentation: `docs/antora/`
