import { useEffect, useState } from 'react'

interface VersionInfo {
  buildTime: string
  version: string
}

/**
 * Hook to check if a new version of the app is available
 * Polls version.json every 5 minutes and notifies when version changes
 */
export const useVersionCheck = () => {
  const [newVersionAvailable, setNewVersionAvailable] = useState(false)
  const [currentVersion, setCurrentVersion] = useState<VersionInfo | null>(null)

  useEffect(() => {
    let initialVersion: VersionInfo | null = null

    const checkVersion = async () => {
      try {
        // Add cache-busting query parameter
        const response = await fetch(`/version.json?t=${Date.now()}`)
        if (!response.ok) {
          console.warn('Could not fetch version info')
          return
        }

        const versionInfo: VersionInfo = await response.json()

        // Store initial version on first check
        if (!initialVersion) {
          initialVersion = versionInfo
          setCurrentVersion(versionInfo)
          return
        }

        // Check if version has changed
        if (
          initialVersion.buildTime !== versionInfo.buildTime ||
          initialVersion.version !== versionInfo.version
        ) {
          console.log('New version detected:', versionInfo)
          setNewVersionAvailable(true)
          setCurrentVersion(versionInfo)
        }
      } catch (error) {
        console.warn('Error checking version:', error)
      }
    }

    // Check immediately on mount
    checkVersion()

    // Then check every 5 minutes
    const interval = setInterval(checkVersion, 5 * 60 * 1000)

    return () => clearInterval(interval)
  }, [])

  const reloadApp = () => {
    window.location.reload()
  }

  return {
    newVersionAvailable,
    currentVersion,
    reloadApp,
  }
}
