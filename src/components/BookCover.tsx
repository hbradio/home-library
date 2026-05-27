import { useState, useEffect, useRef } from 'react'
import { getCoverColor } from '../lib/coverColors'

// Module-level cache shared across all BookCover instances.
// Maps ISBN -> Google Books thumbnail URL (or '' if none found).
const googleBooksCache: Record<string, string> = {}

async function fetchGoogleBooksCover(isbn: string): Promise<string> {
  if (isbn in googleBooksCache) return googleBooksCache[isbn]
  try {
    const resp = await fetch(
      `https://www.googleapis.com/books/v1/volumes?q=isbn:${isbn}&fields=items(volumeInfo/imageLinks)`
    )
    if (!resp.ok) {
      googleBooksCache[isbn] = ''
      return ''
    }
    const data = await resp.json()
    const imageLinks = data?.items?.[0]?.volumeInfo?.imageLinks
    // Prefer thumbnail, fall back to smallThumbnail
    const url = imageLinks?.thumbnail || imageLinks?.smallThumbnail || ''
    // Google returns http URLs; upgrade to https
    const secureUrl = url.replace(/^http:/, 'https:')
    googleBooksCache[isbn] = secureUrl
    return secureUrl
  } catch {
    googleBooksCache[isbn] = ''
    return ''
  }
}

interface BookCoverProps {
  isbn: string
  bookId: string
  title: string
  author?: string
  publishYear?: number | null
  alt: string
  /** 'M' for medium (grid), 'L' for large (detail page) */
  size?: 'M' | 'L'
  /** Additional className for the <img> element */
  className?: string
  loading?: 'lazy' | 'eager'
  /** Called when a valid cover image loads successfully (from either source) */
  onValidCover?: () => void
}

type CoverState = 'open-library' | 'google-books' | 'placeholder'

export default function BookCover({
  isbn,
  bookId,
  title,
  author,
  publishYear,
  alt,
  size = 'M',
  className,
  loading = 'lazy',
  onValidCover,
}: BookCoverProps) {
  const [state, setState] = useState<CoverState>('open-library')
  const [googleUrl, setGoogleUrl] = useState<string | null>(null)
  const imgRef = useRef<HTMLImageElement>(null)
  const coverColor = getCoverColor(bookId)

  const openLibraryUrl = `https://covers.openlibrary.org/b/isbn/${isbn}-${size}.jpg`

  // When Open Library fails, try Google Books
  useEffect(() => {
    if (state !== 'google-books') return
    let cancelled = false
    fetchGoogleBooksCover(isbn).then((url) => {
      if (cancelled) return
      if (url) {
        setGoogleUrl(url)
      } else {
        setState('placeholder')
      }
    })
    return () => { cancelled = true }
  }, [state, isbn])

  const handleOpenLibraryResult = (img: HTMLImageElement, success: boolean) => {
    if (!success || img.naturalWidth < 20 || img.naturalHeight < 20) {
      setState('google-books')
    } else {
      onValidCover?.()
    }
  }

  const handleGoogleBooksResult = (img: HTMLImageElement, success: boolean) => {
    if (!success || img.naturalWidth < 20 || img.naturalHeight < 20) {
      setState('placeholder')
    } else {
      onValidCover?.()
    }
  }

  if (state === 'placeholder') {
    return (
      <div className="cover-placeholder" style={{ background: coverColor.bg }}>
        <span className="cover-placeholder-title" style={{ color: coverColor.title }}>{title}</span>
        {author && <span className="cover-placeholder-author" style={{ color: coverColor.subtitle }}>{author}</span>}
        {publishYear && <span className="cover-placeholder-year" style={{ color: coverColor.subtitle }}>{publishYear}</span>}
      </div>
    )
  }

  if (state === 'google-books') {
    if (!googleUrl) {
      // Still loading the Google Books URL -- show nothing yet (brief flash)
      return null
    }
    return (
      <>
        <img
          ref={imgRef}
          className={className}
          src={googleUrl}
          alt={alt}
          loading={loading}
          onLoad={(e) => handleGoogleBooksResult(e.target as HTMLImageElement, true)}
          onError={(e) => handleGoogleBooksResult(e.target as HTMLImageElement, false)}
        />
        {/* Hidden placeholder ready to replace if Google also fails (handled by state change) */}
      </>
    )
  }

  // Default: Open Library
  return (
    <img
      ref={imgRef}
      className={className}
      src={openLibraryUrl}
      alt={alt}
      loading={loading}
      onLoad={(e) => handleOpenLibraryResult(e.target as HTMLImageElement, true)}
      onError={(e) => handleOpenLibraryResult(e.target as HTMLImageElement, false)}
    />
  )
}
