import { useNavigate } from 'react-router-dom'
import './LandingPage.css'

function LandingPage() {
  const navigate = useNavigate()

  return (
    <div className="landing-page">
      <nav className="navbar">
        <div className="logo-container">
          <img src="/logo.svg" alt="Veidly" className="logo-icon" />
          <span className="logo-text">Veidly</span>
        </div>
        <button className="cta-button" onClick={() => navigate('/map')}>
          Find Events
        </button>
      </nav>

      <section className="hero">
        <div className="hero-content">
          <h1>Real People, Real Connections</h1>
          <p className="hero-subtitle">
            A safe, welcoming space to discover authentic friendships and genuine experiences.
            No fake profiles. No algorithms. Just real people who share your interests.
          </p>
          <div className="hero-points">
            <div className="point">
              <span className="icon">✨</span>
              <span>Authentic communities</span>
            </div>
            <div className="point">
              <span className="icon">🛡️</span>
              <span>Privacy-focused & safe</span>
            </div>
            <div className="point">
              <span className="icon">🌸</span>
              <span>Respectful environment</span>
            </div>
          </div>
          <button className="hero-cta" onClick={() => navigate('/map')}>
            Discover Events Near You
          </button>
        </div>
      </section>

      <section className="features">
        <h2>Why People Love Veidly</h2>
        <div className="features-grid">
          <div className="feature-card">
            <div className="feature-icon">🗺️</div>
            <h3>Safe Event Discovery</h3>
            <p>Browse events on an interactive map. See who's organizing, read reviews, and choose what feels right for you.</p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">🔒</div>
            <h3>Privacy Controls</h3>
            <p>Control who sees your information. Hide your contact details until you join. Your safety, your rules.</p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">👥</div>
            <h3>Verified Participants</h3>
            <p>Email verification required. Know you're meeting real people, not bots or fake accounts.</p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">🎨</div>
            <h3>Diverse Activities</h3>
            <p>From coffee chats to art classes, hiking to book clubs. Find activities that match your vibe.</p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">⚙️</div>
            <h3>Your Preferences</h3>
            <p>Filter by interests, location, age groups, and comfort levels. Create events with your own boundaries.</p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">💬</div>
            <h3>Respectful Community</h3>
            <p>Report inappropriate behavior. Admins actively maintain a welcoming, harassment-free environment.</p>
          </div>
        </div>
      </section>

      <section className="event-types">
        <h2>What You Can Discover</h2>
        <div className="event-types-grid">
          <div className="event-type">
            <span className="emoji">🍻</span>
            <h4>Social & Drinks</h4>
            <p>Coffee dates, wine tastings, casual meetups</p>
          </div>
          <div className="event-type">
            <span className="emoji">🏃</span>
            <h4>Sports & Fitness</h4>
            <p>Yoga, hiking, running groups, dance classes</p>
          </div>
          <div className="event-type">
            <span className="emoji">🍕</span>
            <h4>Food & Dining</h4>
            <p>Cooking classes, restaurant tours, brunch clubs</p>
          </div>
          <div className="event-type">
            <span className="emoji">💼</span>
            <h4>Business & Networking</h4>
            <p>Professional meetups, networking events, co-working</p>
          </div>
          <div className="event-type">
            <span className="emoji">🎮</span>
            <h4>Gaming & Hobbies</h4>
            <p>Board games, crafts, photography walks</p>
          </div>
          <div className="event-type">
            <span className="emoji">📚</span>
            <h4>Learning & Skills</h4>
            <p>Book clubs, language exchange, workshops</p>
          </div>
          <div className="event-type">
            <span className="emoji">✈️</span>
            <h4>Adventure & Travel</h4>
            <p>Day trips, weekend getaways, explore together</p>
          </div>
          <div className="event-type">
            <span className="emoji">👶</span>
            <h4>Parents & Kids</h4>
            <p>Playdates, parent support groups, family outings</p>
          </div>
        </div>
      </section>

      <section className="about">
        <h2>Escape the Social Media Circus</h2>
        <div className="about-content">
          <p className="about-text">
            Tired of endless scrolling through fake AI content? Fed up with toxic comments and comparison culture?
            Exhausted by algorithms that prioritize engagement over your wellbeing?
          </p>
          <p className="about-text highlight">
            <strong>You deserve better.</strong>
          </p>
          <p className="about-text">
            Veidly is built for people who crave genuine human connection. We believe in face-to-face
            conversations, shared laughter, and real friendships. No filters. No facades. Just authentic you,
            meeting authentic others.
          </p>
          <p className="about-text">
            Whether you're new in town, exploring new hobbies, or simply want to expand your circle with
            like-minded people – you're in the right place.
          </p>
        </div>
      </section>

      <section className="opensource">
        <div className="opensource-content">
          <div className="opensource-badge">
            <span className="badge-icon">💝</span>
            <span className="badge-text">Free & Open Source</span>
          </div>
          <h2>Made with Love, Powered by Community</h2>
          <p className="opensource-description">
            Veidly is a <strong>passion project</strong> – built to create meaningful connections,
            not to extract profit from your data. Our code is transparent and community-driven.
          </p>
          <p className="opensource-description">
            We don't sell your data. We don't run ads. We don't have investors demanding growth at any cost.
            <strong> Veidly will always be free.</strong>
          </p>
          <div className="github-section">
            <a
              href="https://github.com/wejdross/veidly.com"
              target="_blank"
              rel="noopener noreferrer"
              className="github-button"
            >
              <svg className="github-icon" viewBox="0 0 16 16" width="20" height="20" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
              </svg>
              <span className="github-text">View on GitHub</span>
            </a>
            <p className="github-note">
              Star the project, report issues, or contribute to make Veidly even better!
            </p>
          </div>
          <div className="funding-callout">
            <h3>💌 Help Keep It Free</h3>
            <p>
              This project runs on coffee, code, and community support. If Veidly has helped you make
              meaningful connections, consider supporting our mission to keep it free and accessible for everyone.
            </p>
            <div className="funding-buttons">
              <a
                href="https://buycoffee.to/veidly"
                target="_blank"
                rel="noopener noreferrer"
                className="funding-button buycoffee"
              >
                <span className="funding-icon">☕</span>
                <span>Buy Me a Coffee</span>
              </a>
            </div>
            <p className="funding-note">
              Every contribution, no matter how small, helps cover hosting costs and keeps development going.
              You're directly supporting a healthier, more authentic alternative to mainstream social media.
            </p>
          </div>
        </div>
      </section>

      <section className="cta-section">
        <div className="cta-content">
          <h2>Ready to Connect?</h2>
          <p className="cta-subtitle">
            Join a community that values authenticity, respect, and genuine human connection.
            Your next great friendship might be just around the corner.
          </p>
          <button className="hero-cta" onClick={() => navigate('/map')}>
            Start Exploring Now
          </button>
        </div>
      </section>

      <footer className="footer">
        <div className="footer-content">
          <div className="footer-section">
            <img src="/logo.svg" alt="Veidly" className="footer-logo" />
            <p className="footer-tagline">Real People, Real Connections</p>
          </div>
          <div className="footer-section">
            <h4>Community</h4>
            <a href="https://github.com/anthropics/veidly" target="_blank" rel="noopener noreferrer">
              GitHub
            </a>
            <a href="mailto:hello@veidly.com">Contact</a>
          </div>
          <div className="footer-section">
            <h4>Support</h4>
            <a href="https://buycoffee.to/veidly" target="_blank" rel="noopener noreferrer">
              Buy Coffee
            </a>
          </div>
        </div>
        <div className="footer-bottom">
          <p>&copy; 2025 Veidly. Open source with 💜 · No ads, no tracking, no BS.</p>
        </div>
      </footer>
    </div>
  )
}

export default LandingPage
