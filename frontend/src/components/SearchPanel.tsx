import { useState } from 'react'
import { CATEGORIES } from '../types'
import { LANGUAGES } from '../languages'
import './SearchPanel.css'

interface SearchFilters {
  keyword: string
  location: string
  category: string
  languages: string
  smoking: string
  alcohol: string
  drugs: string
  gender: string
  age_min: string
  age_max: string
}

interface SearchPanelProps {
  onSearch: (filters: SearchFilters) => void
  onClear: () => void
  initialFilters?: Partial<SearchFilters>
}

function SearchPanel({ onSearch, onClear, initialFilters }: SearchPanelProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [filters, setFilters] = useState<SearchFilters>({
    keyword: initialFilters?.keyword || '',
    location: initialFilters?.location || '',
    category: initialFilters?.category || '',
    languages: initialFilters?.languages || '',
    smoking: initialFilters?.smoking || '',
    alcohol: initialFilters?.alcohol || '',
    drugs: initialFilters?.drugs || '',
    gender: initialFilters?.gender || 'any',
    age_min: initialFilters?.age_min || '',
    age_max: initialFilters?.age_max || '',
  })

  // Helper functions for multi-select language filtering
  const toggleLanguage = (languageCode: string) => {
    const currentLanguages = filters.languages ? filters.languages.split(',') : []
    const updatedLanguages = currentLanguages.includes(languageCode)
      ? currentLanguages.filter(code => code !== languageCode)
      : [...currentLanguages, languageCode]

    setFilters(prev => ({
      ...prev,
      languages: updatedLanguages.join(',')
    }))
  }

  const isLanguageSelected = (code: string) => {
    return filters.languages ? filters.languages.split(',').includes(code) : false
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target
    setFilters(prev => ({ ...prev, [name]: value }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSearch(filters)
    setIsOpen(false)
  }

  const handleClear = () => {
    setFilters({
      keyword: '',
      location: '',
      category: '',
      languages: '',
      smoking: '',
      alcohol: '',
      drugs: '',
      gender: 'any',
      age_min: '',
      age_max: '',
    })
    onClear()
  }

  const hasActiveFilters = Object.values(filters).some(v => v !== '' && v !== 'any')

  return (
    <>
      <button
        className={`search-toggle-button ${hasActiveFilters ? 'has-filters' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
      >
        🔍 {hasActiveFilters ? 'Filters Active' : 'Search & Filter'}
      </button>

      {isOpen && (
        <>
          <div className="search-overlay" onClick={() => setIsOpen(false)} />
          <div className="search-panel">
            <div className="search-header">
              <h2>Search & Filter Events</h2>
              <button className="close-button" onClick={() => setIsOpen(false)}>✕</button>
            </div>

            <form onSubmit={handleSubmit} className="search-form">
              <div className="form-section">
                <h3>🔍 Search</h3>
                <div className="form-group">
                  <label htmlFor="keyword">Keyword</label>
                  <input
                    type="text"
                    id="keyword"
                    name="keyword"
                    value={filters.keyword}
                    onChange={handleChange}
                    placeholder="Search in title or description..."
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="location">Location</label>
                  <input
                    type="text"
                    id="location"
                    name="location"
                    value={filters.location}
                    onChange={handleChange}
                    placeholder="City, place name..."
                  />
                </div>
              </div>

              <div className="form-section">
                <h3>📂 Category</h3>
                <div className="form-group">
                  <select
                    id="category"
                    name="category"
                    value={filters.category}
                    onChange={handleChange}
                  >
                    <option value="">All Categories</option>
                    {Object.entries(CATEGORIES).map(([key, label]) => (
                      <option key={key} value={key}>{label}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="form-section">
                <h3>🗣️ Event Languages</h3>
                <p className="filter-hint">Select one or more languages</p>
                <div className="language-filter-grid">
                  {LANGUAGES.map(lang => (
                    <button
                      key={lang.code}
                      type="button"
                      className={`language-filter-option ${isLanguageSelected(lang.code) ? 'selected' : ''}`}
                      onClick={() => toggleLanguage(lang.code)}
                    >
                      <span className="language-flag">{lang.flag}</span>
                      <span className="language-name">{lang.name}</span>
                    </button>
                  ))}
                </div>
                {filters.languages && (
                  <div className="selected-count">
                    {filters.languages.split(',').length} language(s) selected
                  </div>
                )}
              </div>

              <div className="form-section">
                <h3>🎯 Preferences</h3>
                <div className="form-group">
                  <label htmlFor="smoking">Smoking</label>
                  <select id="smoking" name="smoking" value={filters.smoking} onChange={handleChange}>
                    <option value="">Any</option>
                    <option value="true">Allowed</option>
                    <option value="false">Not Allowed</option>
                  </select>
                </div>

                <div className="form-group">
                  <label htmlFor="alcohol">Alcohol</label>
                  <select id="alcohol" name="alcohol" value={filters.alcohol} onChange={handleChange}>
                    <option value="">Any</option>
                    <option value="true">Allowed</option>
                    <option value="false">Not Allowed</option>
                  </select>
                </div>

                <div className="form-group">
                  <label htmlFor="drugs">Drugs</label>
                  <select id="drugs" name="drugs" value={filters.drugs} onChange={handleChange}>
                    <option value="">Any</option>
                    <option value="true">Allowed</option>
                    <option value="false">Not Allowed</option>
                  </select>
                </div>
              </div>

              <div className="form-section">
                <h3>👥 Demographics</h3>
                <div className="form-group">
                  <label htmlFor="gender">Gender</label>
                  <select id="gender" name="gender" value={filters.gender} onChange={handleChange}>
                    <option value="any">Any</option>
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                    <option value="non-binary">Non-binary</option>
                  </select>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label htmlFor="age_min">Min Age</label>
                    <input
                      type="number"
                      id="age_min"
                      name="age_min"
                      value={filters.age_min}
                      onChange={handleChange}
                      placeholder="0"
                      min="0"
                      max="99"
                    />
                  </div>

                  <div className="form-group">
                    <label htmlFor="age_max">Max Age</label>
                    <input
                      type="number"
                      id="age_max"
                      name="age_max"
                      value={filters.age_max}
                      onChange={handleChange}
                      placeholder="99"
                      min="0"
                      max="99"
                    />
                  </div>
                </div>
              </div>

              <div className="form-actions">
                <button type="button" className="clear-button" onClick={handleClear}>
                  Clear All
                </button>
                <button type="submit" className="apply-button">
                  Apply Filters
                </button>
              </div>
            </form>
          </div>
        </>
      )}
    </>
  )
}

export default SearchPanel
