import { useRef, useEffect, useState } from 'react'

interface Props {
  onScan: (isbn: string) => void
  disabled?: boolean
}

export default function IsbnInput({ onScan, disabled }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [value, setValue] = useState('')

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  const handleSubmit = () => {
    const isbn = value.trim().replace(/-/g, '')
    if (isbn.length >= 10) {
      onScan(isbn)
      setValue('')
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }

  return (
    <input
      ref={inputRef}
      className="scan-input"
      type="text"
      placeholder="Scan or type ISBN..."
      value={value}
      disabled={disabled}
      onChange={(e) => {
        const v = e.target.value.replace(/-/g, '')
        setValue(e.target.value)
        // Auto-submit on 13-digit ISBN
        if (v.length === 13 && /^\d{13}$/.test(v)) {
          onScan(v)
          setValue('')
          setTimeout(() => inputRef.current?.focus(), 50)
        }
      }}
      onKeyDown={(e) => {
        if (e.key === 'Enter') {
          handleSubmit()
        }
      }}
    />
  )
}
