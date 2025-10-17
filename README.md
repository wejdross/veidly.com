# Veidly - Real People, Real Connections

A web application for people overwhelmed by social media's fake AI content. Veidly helps you discover and create real-world meetups with authentic people in your area.

## Tech Stack

**Backend:** Go 1.21+ | Gin | SQLite
**Frontend:** React 18 | TypeScript | Vite | React Leaflet
**Testing:** 62.1% coverage | Dedicated test database

## Quick Start

```bash
# Install dependencies
make install

# Start development servers (backend + frontend)
make dev
```

Backend: `http://localhost:8080`
Frontend: `http://localhost:5173`

### Seed Test Data

To populate the database with test users and events:

```bash
# Set admin password (backend must be running)
export ADMIN_PASSWORD=admin123

# Run seed script
ADMIN_PASSWORD=admin123 python3 seed_test_data.py
```

This creates:
- 5 test users with complete profiles
- 50 diverse events across Europe
- Event participation simulations
- Email verification for all users

Test user login: `anna.kowalski@example.com` / `SecurePass123`

## Common Commands

```bash
make help              # Show all available commands
make install           # Install all dependencies
make dev               # Run development servers
make build             # Build for production
make test              # Run tests (62.1% coverage)
make test-coverage     # View coverage in browser
make test-safety       # Verify production DB safety
make clean             # Clean build artifacts
make docs              # Build and serve documentation
```

## Documentation

**All comprehensive documentation is available via Antora.**

### View Documentation

```bash
# Build and serve documentation
make docs
```

Then open: **http://localhost:8000**

### Build Documentation Manually

```bash
# Using Docker (recommended)
docker run --rm -v $(pwd):/antora antora/antora:latest antora-playbook.yml
cd build/site && python3 -m http.server 8000

# Using local Antora
npm install -g @antora/cli @antora/site-generator
antora antora-playbook.yml
cd build/site && python3 -m http.server 8000
```

### Documentation Structure

All documentation is in `docs/modules/ROOT/pages/`:

- **index.adoc** - Project overview
- **guides/** - User guides (quickstart, features, implementation)
- **architecture/** - System architecture and design
- **testing/** - Test documentation and safety

## Testing

**Coverage: 62.1%** | **All tests passing** ✅

```bash
make test              # Run all tests
make test-coverage     # View coverage report
make test-safety       # Verify production DB safety
```

**Test Database:** Uses dedicated `veidly-test-suite.db` (auto-managed)
**Production Safety:** ✅ Verified - never touches `veidly.db`

## Deployment

### Free Tier AWS Deployment

Veidly is configured for AWS Free Tier deployment:

- **EC2**: t2.micro instance (750 hours/month free)
- **Storage**: 30 GB EBS (8 GB root + 20 GB data, gp2 free tier eligible)
- **Budget Monitoring**: Automatic cost alerts at 80%, 90%, and 100% of budget
- **Monthly Cost**: $0 (within free tier limits)

```bash
# See deployment guide
make docs
# Navigate to Deployment → Terraform Guide
```

### Building for Production

```bash
make build             # Build both backend and frontend
make deploy-check      # Verify deployment readiness
```

## Project Structure

```
veidly.com/
├── backend/                    # Backend (Go)
│   ├── main.go                # Server entry point
│   ├── handlers.go            # API handlers
│   ├── models.go              # Data models
│   ├── auth.go                # Authentication
│   ├── privacy.go             # Privacy controls
│   ├── ics.go                 # Calendar export
│   └── handlers_test.go       # Tests (62.1% coverage)
├── frontend/                   # Frontend (React + TypeScript)
├── terraform/                  # Infrastructure as Code
│   ├── modules/               # Reusable modules
│   │   ├── ec2/              # EC2 instance (t2.micro free tier)
│   │   ├── security/         # Security groups & IAM
│   │   ├── route53/          # DNS configuration
│   │   └── budget/           # Cost monitoring & alerts
│   └── environments/
│       └── production/       # Production environment
├── docs/antora/               # Antora documentation
├── .github/workflows/         # CI/CD pipelines
├── Makefile                   # Build automation
└── README.md                  # This file
```

## Features

- User authentication with JWT
- Event creation and management
- Event participation (join/leave with capacity limits)
- Multi-language support (10 languages)
- Advanced search and filtering
- Admin panel for moderation
- Interactive map with real-time events
- Shareable search URLs

## Prerequisites

- Go 1.21+
- Node.js 18+
- SQLite3

## API Endpoints

### Authentication
- `POST /api/register` - Register user
- `POST /api/login` - Login

### Events
- `GET /api/events` - List events (with filters)
- `GET /api/events/:id` - Get event
- `POST /api/events` - Create event
- `PUT /api/events/:id` - Update event
- `DELETE /api/events/:id` - Delete event

### Participation
- `POST /api/events/:id/join` - Join event
- `DELETE /api/events/:id/leave` - Leave event
- `GET /api/events/:id/participants` - Get participants

### Profile
- `GET /api/profile` - Get own profile
- `PUT /api/profile` - Update profile
- `GET /api/profile/:id` - View user profile

### Admin
- `GET /api/admin/users` - List users
- `PUT /api/admin/users/:id/block` - Block user
- `PUT /api/admin/users/:id/unblock` - Unblock user

**For complete API documentation, build the Antora docs:** `make docs`

## Development Notes

- Backend includes CORS support for dev and production
- SQLite database auto-created on first run
- Events filtered to show only upcoming events
- JWT tokens expire after 7 days
- Test database (`veidly-test-suite.db`) is auto-managed

## License

Open source - available for personal and commercial use.

---

**📚 For complete documentation including architecture, testing, deployment guides, and more, run:** `make docs`
