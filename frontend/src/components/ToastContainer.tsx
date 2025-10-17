import { useState, useCallback, useEffect } from 'react'
import Toast, { ToastType } from './Toast'
import './Toast.css'

export interface ToastItem {
  id: number
  message: string
  type: ToastType
}

let toastId = 0

// Global toast manager
let globalShowToast: ((message: string, type: ToastType) => void) | null = null

export const showToast = (message: string, type: ToastType = 'info') => {
  if (globalShowToast) {
    globalShowToast(message, type)
  }
}

function ToastContainer() {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const addToast = useCallback((message: string, type: ToastType) => {
    const id = toastId++
    setToasts((prev) => [...prev, { id, message, type }])
  }, [])

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }, [])

  // Register global toast function with proper cleanup
  useEffect(() => {
    globalShowToast = addToast

    return () => {
      globalShowToast = null
    }
  }, [addToast])

  return (
    <div className="toast-container">
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          message={toast.message}
          type={toast.type}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </div>
  )
}

export default ToastContainer
