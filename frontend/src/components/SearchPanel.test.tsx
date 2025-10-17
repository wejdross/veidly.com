import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import SearchPanel from './SearchPanel'

describe('SearchPanel Component', () => {
  const mockOnSearch = vi.fn()
  const mockOnClear = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Panel Toggle', () => {
    it('should render search toggle button', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      expect(screen.getByRole('button', { name: /search & filter/i })).toBeInTheDocument()
    })

    it('should open panel when toggle button is clicked', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      const toggleButton = screen.getByRole('button', { name: /search & filter/i })
      fireEvent.click(toggleButton)

      expect(screen.getByText('Search & Filter Events')).toBeInTheDocument()
    })

    it('should close panel when close button is clicked', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      // Open panel
      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Close panel
      const closeButton = screen.getByRole('button', { name: '✕' })
      fireEvent.click(closeButton)

      expect(screen.queryByText('Search & Filter Events')).not.toBeInTheDocument()
    })

    it('should close panel when overlay is clicked', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      // Open panel
      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Click overlay
      const overlay = document.querySelector('.search-overlay')
      if (overlay) {
        fireEvent.click(overlay)
      }

      expect(screen.queryByText('Search & Filter Events')).not.toBeInTheDocument()
    })
  })

  describe('Filter Inputs', () => {
    it('should handle keyword input changes', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      const keywordInput = screen.getByLabelText(/keyword/i) as HTMLInputElement
      fireEvent.change(keywordInput, { target: { value: 'coffee' } })

      expect(keywordInput.value).toBe('coffee')
    })

    it('should handle location input changes', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      const locationInput = screen.getByLabelText(/location/i) as HTMLInputElement
      fireEvent.change(locationInput, { target: { value: 'Warsaw' } })

      expect(locationInput.value).toBe('Warsaw')
    })
  })

  describe('Filter Submission', () => {
    it('should call onSearch with filter values when form is submitted', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Fill filters
      const keywordInput = screen.getByLabelText(/keyword/i)
      fireEvent.change(keywordInput, { target: { value: 'coffee' } })

      const selects = screen.getAllByRole('combobox')
      fireEvent.change(selects[0], { target: { value: 'social_drinks' } })

      // Submit
      fireEvent.click(screen.getByRole('button', { name: /apply filters/i }))

      expect(mockOnSearch).toHaveBeenCalledWith(
        expect.objectContaining({
          keyword: 'coffee',
          category: 'social_drinks',
        })
      )
    })

    it('should close panel after submitting filters', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Submit
      fireEvent.click(screen.getByRole('button', { name: /apply filters/i }))

      expect(screen.queryByText('Search & Filter Events')).not.toBeInTheDocument()
    })
  })

  describe('Clear Filters', () => {
    it('should call onClear when clear button is clicked', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Fill a filter
      const keywordInput = screen.getByLabelText(/keyword/i)
      fireEvent.change(keywordInput, { target: { value: 'coffee' } })

      // Clear
      fireEvent.click(screen.getByRole('button', { name: /clear all/i }))

      expect(mockOnClear).toHaveBeenCalled()
    })

    it('should reset all filter inputs when clear is clicked', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Fill filters
      const keywordInput = screen.getByLabelText(/keyword/i) as HTMLInputElement
      fireEvent.change(keywordInput, { target: { value: 'coffee' } })

      const selects = screen.getAllByRole('combobox')
      const categorySelect = selects[0] as HTMLSelectElement
      fireEvent.change(categorySelect, { target: { value: 'social_drinks' } })

      // Clear
      fireEvent.click(screen.getByRole('button', { name: /clear all/i }))

      expect(keywordInput.value).toBe('')
      expect(categorySelect.value).toBe('')
    })
  })

  describe('Initial Filters', () => {

    it('should show "Filters Active" when there are active filters', () => {
      const initialFilters = {
        keyword: 'coffee',
      }

      render(
        <SearchPanel
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          initialFilters={initialFilters}
        />
      )

      expect(screen.getByText(/filters active/i)).toBeInTheDocument()
    })

    it('should not show "Filters Active" when all filters are empty or "any"', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      expect(screen.getByText(/search & filter/i)).toBeInTheDocument()
      expect(screen.queryByText(/filters active/i)).not.toBeInTheDocument()
    })
  })

  describe('Filter Options', () => {
    it('should display all category options', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Check for some category options (without role due to multiple selects)
      expect(screen.getByText(/social & drinks/i)).toBeInTheDocument()
      expect(screen.getByText(/sports & fitness/i)).toBeInTheDocument()
    })

    it('should display language options', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      // Check for language options
      expect(screen.getByText('English')).toBeInTheDocument()
      expect(screen.getByText('German')).toBeInTheDocument()
    })

    it('should display gender select', () => {
      render(<SearchPanel onSearch={mockOnSearch} onClear={mockOnClear} />)

      fireEvent.click(screen.getByRole('button', { name: /search & filter/i }))

      const genderSelect = screen.getByLabelText(/gender/i)
      expect(genderSelect).toBeInTheDocument()
    })
  })
})
