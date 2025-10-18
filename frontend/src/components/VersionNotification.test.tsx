import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '../test/testUtils'
import VersionNotification from './VersionNotification'
import * as useVersionCheckHook from '../hooks/useVersionCheck'

// Mock the useVersionCheck hook
vi.mock('../hooks/useVersionCheck', () => ({
  useVersionCheck: vi.fn(),
}))

describe('VersionNotification Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should not render when no new version is available', () => {
    vi.mocked(useVersionCheckHook.useVersionCheck).mockReturnValue({
      newVersionAvailable: false,
      currentVersion: null,
      reloadApp: vi.fn(),
    })

    const { container } = render(<VersionNotification />)
    expect(container.firstChild).toBeNull()
  })

  it('should render notification when new version is available', () => {
    vi.mocked(useVersionCheckHook.useVersionCheck).mockReturnValue({
      newVersionAvailable: true,
      currentVersion: { buildTime: '2024-01-01T00:00:00Z', version: '0.2.0' },
      reloadApp: vi.fn(),
    })

    render(<VersionNotification />)

    expect(screen.getByText('New version available!')).toBeInTheDocument()
    expect(screen.getByText(/A new version of Veidly is available/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /reload now/i })).toBeInTheDocument()
  })

  it('should call reloadApp when button is clicked', () => {
    const mockReloadApp = vi.fn()
    vi.mocked(useVersionCheckHook.useVersionCheck).mockReturnValue({
      newVersionAvailable: true,
      currentVersion: { buildTime: '2024-01-01T00:00:00Z', version: '0.2.0' },
      reloadApp: mockReloadApp,
    })

    render(<VersionNotification />)

    const reloadButton = screen.getByRole('button', { name: /reload now/i })
    reloadButton.click()

    expect(mockReloadApp).toHaveBeenCalledTimes(1)
  })
})
