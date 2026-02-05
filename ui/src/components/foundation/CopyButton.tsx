import { useState, useCallback } from 'react'

export interface CopyButtonProps {
  /** Text to copy to clipboard */
  text: string
  /** Optional button label (default: icon only) */
  label?: string
  /** Size variant */
  size?: 'sm' | 'md'
  /** Style variant */
  variant?: 'icon' | 'text' | 'outline'
  /** Callback after successful copy */
  onCopy?: () => void
  /** Custom class name */
  className?: string
}

const CheckIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
    <path d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z" />
  </svg>
)

const CopyIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
    <path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z" />
    <path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z" />
  </svg>
)

export function CopyButton({
  text,
  label,
  size = 'md',
  variant = 'icon',
  onCopy,
  className = '',
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      onCopy?.()
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }, [text, onCopy])

  const sizeClasses = size === 'sm' ? 'btn-sm' : ''
  const variantClasses = {
    icon: 'btn-link p-1',
    text: 'btn-secondary',
    outline: 'btn-outline-secondary',
  }[variant]

  return (
    <button
      type="button"
      className={`btn ${variantClasses} ${sizeClasses} d-inline-flex align-items-center gap-1 ${className}`}
      onClick={handleCopy}
      title={copied ? 'Copied!' : `Copy: ${text}`}
      aria-label={copied ? 'Copied to clipboard' : `Copy ${text} to clipboard`}
    >
      {copied ? <CheckIcon /> : <CopyIcon />}
      {label && <span>{copied ? 'Copied!' : label}</span>}
    </button>
  )
}
