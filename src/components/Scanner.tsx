import { useEffect, useRef } from 'react'
import { Html5Qrcode, Html5QrcodeSupportedFormats } from 'html5-qrcode'

const BARCODE_FORMATS = [
  Html5QrcodeSupportedFormats.EAN_13,
  Html5QrcodeSupportedFormats.EAN_8,
  Html5QrcodeSupportedFormats.UPC_A,
  Html5QrcodeSupportedFormats.UPC_E,
  Html5QrcodeSupportedFormats.CODE_128,
  Html5QrcodeSupportedFormats.CODE_39,
]

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

    const scanner = new Html5Qrcode('scanner-container', {
      formatsToSupport: BARCODE_FORMATS,
      verbose: false,
      experimentalFeatures: { useBarCodeDetectorIfSupported: true },
    })
    scannerRef.current = scanner

    scanner.start(
      { facingMode: { exact: 'environment' } },
      {
        fps: 20,
        aspectRatio: 1.5,
        videoConstraints: {
          facingMode: { exact: 'environment' },
          width: { ideal: 1920 },
          height: { ideal: 1080 },
        },
      },
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
