import { useEffect, useRef } from 'react'
import { Html5Qrcode } from 'html5-qrcode'

interface Props {
  onScan: (isbn: string) => void
  active: boolean
}

export default function Scanner({ onScan, active }: Props) {
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const lastScanRef = useRef<string>('')

  useEffect(() => {
    if (!active || !containerRef.current) return

    const scanner = new Html5Qrcode('scanner-container')
    scannerRef.current = scanner

    scanner.start(
      { facingMode: 'environment' },
      { fps: 10, qrbox: { width: 250, height: 100 } },
      (text) => {
        const isbn = text.replace(/-/g, '')
        if (isbn !== lastScanRef.current && isbn.length >= 10) {
          lastScanRef.current = isbn
          onScan(isbn)
          setTimeout(() => { lastScanRef.current = '' }, 3000)
        }
      },
      () => {}
    ).catch(console.error)

    return () => {
      scanner.stop().catch(() => {})
    }
  }, [active, onScan])

  if (!active) return null

  return (
    <div style={{ margin: '1em 0', maxWidth: '400px' }}>
      <div id="scanner-container" ref={containerRef} />
    </div>
  )
}
