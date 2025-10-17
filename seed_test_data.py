#!/usr/bin/env python3

"""
Veidly Test Data Seeding Script (Python)
Creates 5 users with 10 diverse events each across Europe
"""

import requests
import time
import sys
import os
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import random

# Configuration
API_URL = "http://localhost:8080/api"
VERBOSE = False

# Colors for terminal output
class Colors:
    BLUE = '\033[0;34m'
    GREEN = '\033[0;32m'
    RED = '\033[0;31m'
    YELLOW = '\033[1;33m'
    NC = '\033[0m'  # No Color

def log_info(message: str):
    print(f"{Colors.BLUE}ℹ {Colors.NC} {message}")

def log_success(message: str):
    print(f"{Colors.GREEN}✓{Colors.NC} {message}")

def log_error(message: str):
    print(f"{Colors.RED}✗{Colors.NC} {message}")

def log_section(message: str):
    print(f"\n{Colors.YELLOW}{'═' * 55}{Colors.NC}")
    print(f"{Colors.YELLOW}  {message}{Colors.NC}")
    print(f"{Colors.YELLOW}{'═' * 55}{Colors.NC}\n")

def api_call(method: str, endpoint: str, data: Optional[Dict] = None, token: Optional[str] = None) -> tuple:
    """Make API call and return (status_code, response_data)"""
    url = f"{API_URL}{endpoint}"
    headers = {"Content-Type": "application/json"}

    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        if method == "GET":
            response = requests.get(url, headers=headers)
        elif method == "POST":
            response = requests.post(url, json=data, headers=headers)
        elif method == "PUT":
            response = requests.put(url, json=data, headers=headers)
        elif method == "DELETE":
            response = requests.delete(url, headers=headers)
        else:
            return (0, {"error": "Invalid method"})

        if VERBOSE:
            print(f"→ {method} {endpoint} - Status: {response.status_code}")

        try:
            return (response.status_code, response.json())
        except:
            return (response.status_code, {})
    except requests.exceptions.RequestException as e:
        log_error(f"Request failed: {e}")
        return (0, {"error": str(e)})

def future_datetime(days: int, hour: int) -> str:
    """Generate future datetime in RFC3339 format"""
    future = datetime.now() + timedelta(days=days)
    future = future.replace(hour=hour, minute=0, second=0, microsecond=0)
    return future.strftime("%Y-%m-%dT%H:%M:%SZ")

def check_backend():
    """Check if backend is accessible"""
    log_info(f"API URL: {API_URL}")
    log_info("Checking if backend is accessible...")

    try:
        response = requests.get(f"{API_URL.replace('/api', '')}/health", timeout=5)
        if response.status_code == 200:
            log_success("Backend is accessible")
            return True
    except:
        pass

    # Try /events endpoint as fallback
    try:
        response = requests.get(f"{API_URL}/events", timeout=5)
        if response.status_code in [200, 401]:
            log_success("Backend is accessible")
            return True
    except:
        pass

    log_error(f"Backend is not accessible at {API_URL}")
    log_info("Please start the backend with: make dev")
    return False

# User definitions
USERS = [
    {
        "username": "anna",
        "email": "anna.kowalski@example.com",
        "password": "SecurePass123",
        "name": "Anna Kowalski",
        "bio": "Polish mom looking for playdates and coffee",
        "phone": "+48123456789",
        "threema": "ANNAPL",
        "languages": "pl,en"
    },
    {
        "username": "marco",
        "email": "marco.rossi@example.com",
        "password": "SecurePass123",
        "name": "Marco Rossi",
        "bio": "Italian sports enthusiast, love basketball and cycling",
        "phone": "+39334567890",
        "threema": "MARCOITA",
        "languages": "it,en"
    },
    {
        "username": "sophie",
        "email": "sophie.martin@example.com",
        "password": "SecurePass123",
        "name": "Sophie Martin",
        "bio": "French foodie and culture lover",
        "phone": "+33612345678",
        "threema": "SOPHIEFR",
        "languages": "fr,en"
    },
    {
        "username": "hans",
        "email": "hans.mueller@example.com",
        "password": "SecurePass123",
        "name": "Hans Müller",
        "bio": "German hiker and nature enthusiast",
        "phone": "+49151234567",
        "threema": "HANSDE",
        "languages": "de,en"
    },
    {
        "username": "elena",
        "email": "elena.garcia@example.com",
        "password": "SecurePass123",
        "name": "Elena García",
        "bio": "Spanish dancer and social butterfly",
        "phone": "+34612345678",
        "threema": "ELENAES",
        "languages": "es,en"
    }
]

# Event definitions for each user
ANNA_EVENTS = [
    {"location": "Warsaw, Poland", "lat": 52.2297, "lng": 21.0122, "title": "Playground Meetup for Toddlers", "desc": "Looking for other moms with toddlers (2-4 years) to meet at the playground. Let's chat while kids play!", "category": "parents_kids", "days": 2, "hour": 10, "max_p": 5, "gender": "any", "age_min": 25, "age_max": 45},
    {"location": "Krakow, Poland", "lat": 50.0647, "lng": 19.9450, "title": "Coffee Morning for New Moms", "desc": "New to motherhood? Join us for coffee and support. Share experiences and make friends.", "category": "social_drinks", "days": 3, "hour": 9, "max_p": 8, "gender": "female", "age_min": 20, "age_max": 50},
    {"location": "Gdansk, Poland", "lat": 54.3520, "lng": 18.6466, "title": "Baby Swimming Class", "desc": "Infant swimming session at local pool. Certified instructor, fun and safe environment.", "category": "sports_fitness", "days": 5, "hour": 11, "max_p": 10, "gender": "any", "age_min": 0, "age_max": 5},
    {"location": "Poznan, Poland", "lat": 52.4064, "lng": 16.9252, "title": "Kids Birthday Party Planning", "desc": "Planning a birthday party? Let's share ideas, vendors, and tips for amazing kids parties.", "category": "social_drinks", "days": 4, "hour": 14, "max_p": 6, "gender": "any", "age_min": 25, "age_max": 45},
    {"location": "Wroclaw, Poland", "lat": 51.1079, "lng": 17.0385, "title": "Stroller Fitness Walk", "desc": "Power walking with strollers! Get fit while bonding with other moms. All fitness levels welcome.", "category": "sports_fitness", "days": 6, "hour": 10, "max_p": 12, "gender": "female", "age_min": 20, "age_max": 45},
    {"location": "Lodz, Poland", "lat": 51.7592, "lng": 19.4560, "title": "Bilingual Playgroup (Polish-English)", "desc": "Playgroup for kids to practice English. Songs, games, and fun activities.", "category": "parents_kids", "days": 7, "hour": 15, "max_p": 8, "gender": "any", "age_min": 0, "age_max": 10},
    {"location": "Szczecin, Poland", "lat": 53.4285, "lng": 14.5528, "title": "Mom's Book Club", "desc": "Monthly book club for mothers. This month: modern parenting books. Bring wine!", "category": "social_drinks", "days": 8, "hour": 19, "max_p": 10, "gender": "female", "age_min": 25, "age_max": 55},
    {"location": "Lublin, Poland", "lat": 51.2465, "lng": 22.5684, "title": "Craft Afternoon for Kids", "desc": "Arts and crafts session for children 5-10. Materials provided. Parents stay and chat!", "category": "business_networking", "days": 9, "hour": 14, "max_p": 15, "gender": "any", "age_min": 5, "age_max": 12},
    {"location": "Bydgoszcz, Poland", "lat": 53.1235, "lng": 18.0084, "title": "Postpartum Support Group", "desc": "Safe space for new mothers to discuss challenges, share experiences, and support each other.", "category": "social_drinks", "days": 10, "hour": 17, "max_p": 8, "gender": "female", "age_min": 20, "age_max": 45},
    {"location": "Katowice, Poland", "lat": 50.2649, "lng": 19.0238, "title": "Family Picnic in the Park", "desc": "Bring your family for a relaxed picnic. Kids can play, adults can network.", "category": "social_drinks", "days": 11, "hour": 12, "max_p": 20, "gender": "any", "age_min": 0, "age_max": 99},
]

MARCO_EVENTS = [
    {"location": "Rome, Italy", "lat": 41.9028, "lng": 12.4964, "title": "Pickup Basketball Game", "desc": "Looking for 5v5 basketball at outdoor court. All skill levels welcome. Bring water!", "category": "sports_fitness", "days": 2, "hour": 18, "max_p": 10, "gender": "any", "age_min": 18, "age_max": 45},
    {"location": "Milan, Italy", "lat": 45.4642, "lng": 9.1900, "title": "Morning Cycling Tour", "desc": "30km road cycling through city and countryside. Medium pace, coffee break included.", "category": "sports_fitness", "days": 3, "hour": 7, "max_p": 8, "gender": "any", "age_min": 20, "age_max": 55},
    {"location": "Florence, Italy", "lat": 43.7696, "lng": 11.2558, "title": "Football/Soccer Kickabout", "desc": "Casual soccer game in the park. Just for fun, no pressure. All ages and levels.", "category": "sports_fitness", "days": 4, "hour": 17, "max_p": 22, "gender": "any", "age_min": 16, "age_max": 50},
    {"location": "Venice, Italy", "lat": 45.4408, "lng": 12.3155, "title": "Beach Volleyball Tournament", "desc": "Beach volleyball at Lido. Form teams of 4. Prizes for winners! BBQ after.", "category": "sports_fitness", "days": 5, "hour": 15, "max_p": 16, "gender": "any", "age_min": 18, "age_max": 40},
    {"location": "Naples, Italy", "lat": 40.8518, "lng": 14.2681, "title": "Tennis Doubles Match", "desc": "Looking for tennis partners for doubles. Intermediate level. Courts reserved.", "category": "sports_fitness", "days": 6, "hour": 9, "max_p": 4, "gender": "any", "age_min": 25, "age_max": 50},
    {"location": "Turin, Italy", "lat": 45.0703, "lng": 7.6869, "title": "Mountain Biking Trail", "desc": "MTB ride through nearby trails. Moderate difficulty. Helmets required!", "category": "sports_fitness", "days": 7, "hour": 10, "max_p": 6, "gender": "any", "age_min": 20, "age_max": 45},
    {"location": "Bologna, Italy", "lat": 44.4949, "lng": 11.3426, "title": "Climbing Gym Session", "desc": "Indoor climbing at local gym. Beginners welcome, equipment available for rent.", "category": "sports_fitness", "days": 8, "hour": 19, "max_p": 8, "gender": "any", "age_min": 18, "age_max": 55},
    {"location": "Genoa, Italy", "lat": 44.4056, "lng": 8.9463, "title": "Running Club Meetup", "desc": "Weekly 10km run along the waterfront. All paces welcome. Stretching session after.", "category": "sports_fitness", "days": 9, "hour": 6, "max_p": 15, "gender": "any", "age_min": 18, "age_max": 60},
    {"location": "Verona, Italy", "lat": 45.4384, "lng": 10.9916, "title": "Yoga in the Park", "desc": "Outdoor yoga session at sunset. Bring your mat. Suitable for all levels.", "category": "sports_fitness", "days": 10, "hour": 18, "max_p": 20, "gender": "any", "age_min": 16, "age_max": 65},
    {"location": "Palermo, Italy", "lat": 38.1157, "lng": 13.3615, "title": "Swim Training Group", "desc": "Open water swimming practice. Lifeguard present. Intermediate swimmers.", "category": "sports_fitness", "days": 11, "hour": 8, "max_p": 12, "gender": "any", "age_min": 20, "age_max": 50},
]

SOPHIE_EVENTS = [
    {"location": "Paris, France", "lat": 48.8566, "lng": 2.3522, "title": "Wine Tasting Evening", "desc": "Discover French wines! Expert sommelier guides us through 6 wines and cheeses.", "category": "food_dining", "days": 2, "hour": 19, "max_p": 12, "gender": "any", "age_min": 21, "age_max": 60},
    {"location": "Lyon, France", "lat": 45.7640, "lng": 4.8357, "title": "Cooking Class: French Pastries", "desc": "Learn to make croissants and pain au chocolat from scratch. Take home your creations!", "category": "food_dining", "days": 3, "hour": 14, "max_p": 8, "gender": "any", "age_min": 18, "age_max": 65},
    {"location": "Marseille, France", "lat": 43.2965, "lng": 5.3698, "title": "Food Market Tour", "desc": "Explore local markets, taste regional specialties. Lunch at hidden gem bistro included.", "category": "food_dining", "days": 4, "hour": 10, "max_p": 10, "gender": "any", "age_min": 25, "age_max": 70},
    {"location": "Nice, France", "lat": 43.7102, "lng": 7.2620, "title": "Picnic with French Delicacies", "desc": "Bring your favorite French dish to share. Wine, cheese, and conversation by the sea.", "category": "food_dining", "days": 5, "hour": 12, "max_p": 15, "gender": "any", "age_min": 20, "age_max": 55},
    {"location": "Bordeaux, France", "lat": 44.8378, "lng": -0.5792, "title": "Vineyard Visit & Lunch", "desc": "Day trip to Bordeaux vineyard. Wine tasting, tour, and gourmet lunch.", "category": "food_dining", "days": 6, "hour": 11, "max_p": 12, "gender": "any", "age_min": 21, "age_max": 65},
    {"location": "Toulouse, France", "lat": 43.6047, "lng": 1.4442, "title": "French Cinema Night", "desc": "Watch classic French film (with subtitles) followed by discussion and drinks.", "category": "business_networking", "days": 7, "hour": 20, "max_p": 20, "gender": "any", "age_min": 18, "age_max": 99},
    {"location": "Strasbourg, France", "lat": 48.5734, "lng": 7.7521, "title": "Museum & Gallery Hopping", "desc": "Visit 3 art museums in one afternoon. Share impressions over coffee after.", "category": "business_networking", "days": 8, "hour": 14, "max_p": 8, "gender": "any", "age_min": 22, "age_max": 70},
    {"location": "Nantes, France", "lat": 47.2184, "lng": -1.5536, "title": "Live Jazz & Dinner", "desc": "Evening at jazz club. Dinner reservations at 7pm, music starts at 9pm.", "category": "business_networking", "days": 9, "hour": 19, "max_p": 10, "gender": "any", "age_min": 25, "age_max": 60},
    {"location": "Lille, France", "lat": 50.6292, "lng": 3.0573, "title": "French Conversation Exchange", "desc": "Practice French! Native speakers welcome. Casual café setting, order your own.", "category": "learning_skills", "days": 10, "hour": 18, "max_p": 12, "gender": "any", "age_min": 18, "age_max": 99},
    {"location": "Montpellier, France", "lat": 43.6108, "lng": 3.8767, "title": "Chocolate Making Workshop", "desc": "Learn chocolate making from artisan chocolatier. Taste and take home samples!", "category": "food_dining", "days": 11, "hour": 15, "max_p": 10, "gender": "any", "age_min": 18, "age_max": 65},
]

HANS_EVENTS = [
    {"location": "Munich, Germany", "lat": 48.1351, "lng": 11.5820, "title": "Alpine Hiking Adventure", "desc": "Full day hike in Bavarian Alps. 15km, moderate difficulty. Pack lunch and water.", "category": "adventure_travel", "days": 2, "hour": 8, "max_p": 8, "gender": "any", "age_min": 20, "age_max": 55},
    {"location": "Berlin, Germany", "lat": 52.5200, "lng": 13.4050, "title": "Urban Nature Walk", "desc": "Discover Berlin's hidden green spaces and parks. 2-hour easy walk with nature guide.", "category": "adventure_travel", "days": 3, "hour": 10, "max_p": 15, "gender": "any", "age_min": 16, "age_max": 70},
    {"location": "Hamburg, Germany", "lat": 53.5511, "lng": 9.9937, "title": "Birdwatching by the Lake", "desc": "Early morning birdwatching. Bring binoculars if you have them. Hot coffee provided!", "category": "adventure_travel", "days": 4, "hour": 6, "max_p": 10, "gender": "any", "age_min": 18, "age_max": 75},
    {"location": "Frankfurt, Germany", "lat": 50.1109, "lng": 8.6821, "title": "Forest Bathing (Shinrin-yoku)", "desc": "Mindful walk through forest. Reduce stress, connect with nature. Suitable for all.", "category": "adventure_travel", "days": 5, "hour": 14, "max_p": 12, "gender": "any", "age_min": 18, "age_max": 65},
    {"location": "Cologne, Germany", "lat": 50.9375, "lng": 6.9603, "title": "Rhine River Kayaking", "desc": "Kayaking trip on the Rhine. Equipment provided. Basic swimming skills required.", "category": "sports_fitness", "days": 6, "hour": 9, "max_p": 8, "gender": "any", "age_min": 18, "age_max": 50},
    {"location": "Stuttgart, Germany", "lat": 48.7758, "lng": 9.1829, "title": "Camping Weekend Prep", "desc": "Planning meeting for weekend camping trip. Discuss gear, location, and logistics.", "category": "adventure_travel", "days": 7, "hour": 19, "max_p": 10, "gender": "any", "age_min": 20, "age_max": 55},
    {"location": "Dresden, Germany", "lat": 51.0504, "lng": 13.7373, "title": "Rock Climbing in Saxon Switzerland", "desc": "Outdoor rock climbing for experienced climbers. Safety equipment required.", "category": "sports_fitness", "days": 8, "hour": 8, "max_p": 6, "gender": "any", "age_min": 21, "age_max": 50},
    {"location": "Heidelberg, Germany", "lat": 49.3988, "lng": 8.6724, "title": "Photography Walk in Nature", "desc": "Capture autumn colors! Bring your camera. Share tips and techniques.", "category": "business_networking", "days": 9, "hour": 15, "max_p": 12, "gender": "any", "age_min": 18, "age_max": 70},
    {"location": "Nuremberg, Germany", "lat": 49.4521, "lng": 11.0767, "title": "Wild Mushroom Foraging", "desc": "Learn to identify edible mushrooms with expert mycologist. Cook findings together!", "category": "adventure_travel", "days": 10, "hour": 9, "max_p": 8, "gender": "any", "age_min": 25, "age_max": 65},
    {"location": "Leipzig, Germany", "lat": 51.3397, "lng": 12.3731, "title": "Bike Tour Through Countryside", "desc": "40km easy cycling through villages and farmland. Stop at beer garden.", "category": "sports_fitness", "days": 11, "hour": 10, "max_p": 10, "gender": "any", "age_min": 18, "age_max": 60},
]

ELENA_EVENTS = [
    {"location": "Madrid, Spain", "lat": 40.4168, "lng": -3.7038, "title": "Salsa Dancing Night", "desc": "Salsa night at local club! Beginners welcome, free lesson at 8pm. Dance till midnight!", "category": "business_networking", "days": 2, "hour": 20, "max_p": 20, "gender": "any", "age_min": 18, "age_max": 45},
    {"location": "Barcelona, Spain", "lat": 41.3851, "lng": 2.1734, "title": "Beach Party & BBQ", "desc": "Sunset beach party at Barceloneta. Bring food to share. Music, dancing, swimming!", "category": "social_drinks", "days": 3, "hour": 18, "max_p": 25, "gender": "any", "age_min": 18, "age_max": 50},
    {"location": "Valencia, Spain", "lat": 39.4699, "lng": -0.3763, "title": "Paella Cooking Party", "desc": "Cook authentic Valencian paella together! Eat, drink wine, make friends.", "category": "food_dining", "days": 4, "hour": 17, "max_p": 12, "gender": "any", "age_min": 20, "age_max": 60},
    {"location": "Seville, Spain", "lat": 37.3891, "lng": -5.9845, "title": "Flamenco Show & Tapas", "desc": "Authentic flamenco performance followed by tapas bar crawl. Olé!", "category": "business_networking", "days": 5, "hour": 21, "max_p": 15, "gender": "any", "age_min": 21, "age_max": 99},
    {"location": "Malaga, Spain", "lat": 36.7213, "lng": -4.4214, "title": "Girls Night Out", "desc": "Ladies only! Dinner, drinks, dancing. Let's have fun and meet new friends!", "category": "social_drinks", "days": 6, "hour": 20, "max_p": 12, "gender": "female", "age_min": 21, "age_max": 45},
    {"location": "Bilbao, Spain", "lat": 43.2630, "lng": -2.9350, "title": "Language Exchange: Spanish-English", "desc": "Practice languages over coffee and pintxos. Native speakers of any language welcome!", "category": "learning_skills", "days": 7, "hour": 19, "max_p": 15, "gender": "any", "age_min": 18, "age_max": 99},
    {"location": "Granada, Spain", "lat": 37.1773, "lng": -3.5986, "title": "Sunset at Alhambra", "desc": "Watch sunset from Alhambra viewpoint. Bring wine and snacks. Magical experience!", "category": "social_drinks", "days": 8, "hour": 19, "max_p": 10, "gender": "any", "age_min": 18, "age_max": 65},
    {"location": "Zaragoza, Spain", "lat": 41.6488, "lng": -0.8891, "title": "Karaoke Night", "desc": "Sing your heart out! Private karaoke room reserved. All skill levels (no judgment!)", "category": "business_networking", "days": 9, "hour": 21, "max_p": 15, "gender": "any", "age_min": 18, "age_max": 99},
    {"location": "Ibiza, Spain", "lat": 38.9067, "lng": 1.4206, "title": "Yoga & Meditation Retreat", "desc": "Day retreat: yoga, meditation, healthy lunch. Find your zen by the sea.", "category": "sports_fitness", "days": 10, "hour": 9, "max_p": 20, "gender": "any", "age_min": 18, "age_max": 70},
    {"location": "San Sebastian, Spain", "lat": 43.3183, "lng": -1.9812, "title": "Pintxos Bar Hopping Tour", "desc": "Try the best pintxos in town! 5 bars, 5 pintxos, lots of laughs.", "category": "food_dining", "days": 11, "hour": 19, "max_p": 12, "gender": "any", "age_min": 21, "age_max": 60},
]

USER_EVENTS = {
    "anna": ("Anna Kowalski", "anna.kowalski@example.com", ANNA_EVENTS),
    "marco": ("Marco Rossi", "marco.rossi@example.com", MARCO_EVENTS),
    "sophie": ("Sophie Martin", "sophie.martin@example.com", SOPHIE_EVENTS),
    "hans": ("Hans Müller", "hans.mueller@example.com", HANS_EVENTS),
    "elena": ("Elena García", "elena.garcia@example.com", ELENA_EVENTS),
}

def create_users() -> Dict[str, tuple]:
    """Create all users and return tokens with user IDs"""
    log_section("👥 Creating 5 Test Users")

    tokens = {}

    for user in USERS:
        log_info(f"Registering {user['name']} ({user['email']})...")

        status, body = api_call("POST", "/auth/register", {
            "email": user['email'],
            "password": user['password'],
            "name": user['name']
        })

        if status in [200, 201]:
            token = body.get("token")
            user_data = body.get("user", {})
            user_id = user_data.get("id")

            if token and user_id:
                tokens[user['username']] = (token, user_id)
                log_success(f"Registered {user['name']} (ID: {user_id})")

                # Update profile
                log_info(f"Updating profile for {user['name']}...")
                status, _ = api_call("PUT", "/profile", {
                    "name": user['name'],
                    "bio": user['bio'],
                    "phone": user['phone'],
                    "threema": user['threema'],
                    "languages": user['languages']
                }, token)

                if status == 200:
                    log_success(f"Profile updated for {user['name']}")
            else:
                log_error(f"No token or user_id in response for {user['name']}")
        else:
            log_error(f"Failed to register {user['name']} (Status: {status})")
            if VERBOSE and body:
                print(f"  Response: {body}")

    return tokens

def get_admin_token() -> Optional[str]:
    """Login as admin user (auto-created by backend) and return token"""
    log_section("🔑 Logging in as Admin")

    # Try to get admin password from environment or use default
    admin_password = os.getenv("ADMIN_PASSWORD", "admin123")

    log_info("Attempting to login as admin@veidly.com...")

    status, body = api_call("POST", "/auth/login", {
        "email": "admin@veidly.com",
        "password": admin_password
    })

    if status == 200:
        token = body.get("token")
        if token:
            log_success("Admin logged in successfully")
            return token
        else:
            log_error("No token in login response")
            return None
    else:
        log_error(f"Failed to login as admin (Status: {status})")
        if VERBOSE and body:
            print(f"  Response: {body}")
        log_info("💡 Hint: Set ADMIN_PASSWORD env var if using custom password")
        log_info("   Or check backend logs for the generated admin password")
        return None

def verify_users_manually(tokens: Dict[str, tuple], admin_token: str) -> bool:
    """Manually verify all user emails using admin endpoint"""
    log_section("✅ Manually Verifying User Emails")

    if not tokens:
        log_error("No users to verify")
        return False

    if not admin_token:
        log_error("No admin token provided")
        return False

    verified_count = 0

    for username, (token, user_id) in tokens.items():
        log_info(f"Verifying user ID {user_id} ({username})...")

        # Use admin endpoint to verify user email
        status, body = api_call("PUT", f"/admin/users/{user_id}/verify-email", None, admin_token)

        if status == 200:
            log_success(f"User {user_id} ({username}) email verified")
            verified_count += 1
        else:
            log_error(f"Failed to verify user {user_id} ({username}) - Status: {status}")
            if VERBOSE and body:
                print(f"  Response: {body}")

    log_success(f"Verified {verified_count}/{len(tokens)} users")
    return verified_count == len(tokens)

def create_events(tokens: Dict[str, tuple]) -> int:
    """Create events for all users and return count"""
    log_section("📍 Creating Diverse Events Across Europe")

    total_created = 0

    for username, (token, user_id) in tokens.items():
        creator_name, creator_email, events = USER_EVENTS[username]

        log_info(f"Creating events for {creator_name} (ID: {user_id})...")

        for event in events:
            event_data = {
                "title": event['title'],
                "description": event['desc'],
                "category": event['category'],
                "latitude": event['lat'],
                "longitude": event['lng'],
                "start_time": future_datetime(event['days'], event['hour']),
                "creator_name": creator_name,
                "creator_contact": creator_email,
                "max_participants": event['max_p'],
                "gender_restriction": event['gender'],
                "age_min": event['age_min'],
                "age_max": event['age_max']
            }

            status, body = api_call("POST", "/events", event_data, token)

            if status in [200, 201]:
                log_success(f"Created: {event['title']}")
                total_created += 1
            else:
                log_error(f"Failed: {event['title']} (Status: {status})")
                if VERBOSE and body:
                    print(f"  Response: {body}")

    return total_created

def simulate_participation(tokens: Dict[str, tuple]):
    """Users join random events"""
    log_section("🤝 Simulating Event Participation")

    log_info("Fetching all events...")
    status, body = api_call("GET", "/events", None)

    if status != 200 or not isinstance(body, list):
        log_error("Failed to fetch events")
        return

    event_ids = [event['id'] for event in body[:20]]  # First 20 events
    log_success(f"Found {len(event_ids)} events, simulating participation...")

    user_names = ["Anna", "Marco", "Sophie", "Hans", "Elena"]

    for i, (username, (token, user_id)) in enumerate(tokens.items()):
        join_count = 0

        for event_id in event_ids:
            # Random decision to join (40% chance)
            if random.random() < 0.4 and join_count < 4:
                status, _ = api_call("POST", f"/events/{event_id}/join", None, token)

                if status in [200, 201]:
                    log_success(f"{user_names[i]} joined event #{event_id}")
                    join_count += 1

            if join_count >= 4:
                break

def test_profiles(tokens: Dict[str, tuple]):
    """Test profile endpoints"""
    log_section("👤 Testing Profile Endpoints")

    user_names = ["Anna", "Marco", "Sophie", "Hans", "Elena"]

    for i, (username, (token, user_id)) in enumerate(tokens.items()):
        log_info(f"Fetching profile for {user_names[i]}...")
        status, body = api_call("GET", "/profile", None, token)

        if status == 200:
            log_success(f"Profile retrieved for {user_names[i]}")
        else:
            log_error(f"Failed to get profile for {user_names[i]} (Status: {status})")

def test_filters():
    """Test search and filter endpoints"""
    log_section("📊 Testing Search and Filters")

    # Test category filter
    log_info("Testing category filter (sports)...")
    status, _ = api_call("GET", "/events?category=activity_sports", None)
    if status == 200:
        log_success("Category filter works")
    else:
        log_error("Category filter failed")

    # Test gender filter
    log_info("Testing gender filter (female only)...")
    status, _ = api_call("GET", "/events?gender=female", None)
    if status == 200:
        log_success("Gender filter works")
    else:
        log_error("Gender filter failed")

    # Test age filter
    log_info("Testing age filter (18-30)...")
    status, _ = api_call("GET", "/events?age=25", None)
    if status == 200:
        log_success("Age filter works")
    else:
        log_error("Age filter failed")

def print_summary(event_count: int):
    """Print summary"""
    log_section("🎯 Summary")

    log_info("Total users created: 5")
    log_info(f"Total events created: {event_count}")
    log_info("Events spread across: Poland, Italy, France, Germany, Spain")
    log_info("Categories covered: sports, food, social, outdoor, culture, music, kids")

    print()
    log_success("Test data seeding completed successfully!")
    print()
    log_info("You can now:")
    print("  • Visit http://localhost:3000 to see all events on the map")
    print("  • Login with any user (password: SecurePass123):")
    print("    - anna.kowalski@example.com (Polish mom)")
    print("    - marco.rossi@example.com (Italian sports enthusiast)")
    print("    - sophie.martin@example.com (French foodie)")
    print("    - hans.mueller@example.com (German hiker)")
    print("    - elena.garcia@example.com (Spanish dancer)")
    print()
    log_section("✨ Done!")

def main():
    global VERBOSE

    # Check for flags
    for arg in sys.argv[1:]:
        if arg == "-v":
            VERBOSE = True

    log_section("🚀 Starting Veidly Test Data Seeding")

    # Check backend
    if not check_backend():
        sys.exit(1)

    # Login as admin (auto-created by backend)
    admin_token = get_admin_token()

    if not admin_token:
        log_error("Cannot continue without admin access")
        sys.exit(1)

    # Create users
    tokens = create_users()
    if len(tokens) < 5:
        log_error(f"Only {len(tokens)} users created. Expected 5.")
        log_info("Continuing anyway...")

    # Verify users manually using admin endpoint
    verify_users_manually(tokens, admin_token)

    # Create events
    event_count = create_events(tokens)

    # Simulate participation
    simulate_participation(tokens)

    # Test profiles
    test_profiles(tokens)

    # Test filters
    test_filters()

    # Print summary
    print_summary(event_count)

if __name__ == "__main__":
    main()
