import { useVersionCheck } from '../hooks/useVersionCheck'
import './VersionNotification.css'

/**
 * Component that displays a notification when a new app version is available
 * Automatically checks for updates every 5 minutes
 */
function VersionNotification() {
  const { newVersionAvailable, reloadApp } = useVersionCheck()

  if (!newVersionAvailable) {
    return null
  }

  return (
    <div className="version-notification">
      <div className="version-notification-content">
        <div className="version-notification-text">
          <strong>New version available!</strong>
          <span>A new version of Veidly is available. Reload to get the latest features and fixes.</span>
        </div>
        <button onClick={reloadApp} className="version-notification-button">
          Reload Now
        </button>
      </div>
    </div>
  )
}

export default VersionNotification
