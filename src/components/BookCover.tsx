import { useState, useEffect, useRef } from 'react'
import { getCoverColor } from '../lib/coverColors'

// Module-level caches shared across all BookCover instances.
const googleBooksCache: Record<string, string> = {}
const olCoverIdCache: Record<string, string> = {}

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
    const url = imageLinks?.thumbnail || imageLinks?.smallThumbnail || ''
    const secureUrl = url.replace(/^http:/, 'https:')
    googleBooksCache[isbn] = secureUrl
    return secureUrl
  } catch {
    googleBooksCache[isbn] = ''
    return ''
  }
}

async function fetchOLCoverByID(isbn: string, size: string): Promise<string> {
  if (isbn in olCoverIdCache) {
    const cached = olCoverIdCache[isbn]
    return cached ? `https://covers.openlibrary.org/b/id/${cached}-${size}.jpg` : ''
  }
  try {
    const resp = await fetch(
      `https://openlibrary.org/search.json?isbn=${isbn}&fields=cover_i&limit=1`
    )
    if (!resp.ok) {
      olCoverIdCache[isbn] = ''
      return ''
    }
    const data = await resp.json()
    const coverId = data?.docs?.[0]?.cover_i
    if (coverId) {
      olCoverIdCache[isbn] = String(coverId)
      return `https://covers.openlibrary.org/b/id/${coverId}-${size}.jpg`
    }
    olCoverIdCache[isbn] = ''
    return ''
  } catch {
    olCoverIdCache[isbn] = ''
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
  /** User-provided cover URL override */
  coverUrl?: string
  /** 'M' for medium (grid), 'L' for large (detail page) */
  size?: 'M' | 'L'
  /** Additional className for the <img> element */
  className?: string
  loading?: 'lazy' | 'eager'
  /** Called when a valid cover image loads successfully (from any source) */
  onValidCover?: () => void
}

type CoverState = 'cover-override' | 'open-library' | 'ol-cover-id' | 'google-books' | 'placeholder'

export default function BookCover({
  isbn,
  bookId,
  title,
  author,
  publishYear,
  alt,
  coverUrl,
  size = 'M',
  className,
  loading = 'lazy',
  onValidCover,
}: BookCoverProps) {
  const [state, setState] = useState<CoverState>(coverUrl ? 'cover-override' : 'open-library')
  const [olCoverIdUrl, setOlCoverIdUrl] = useState<string | null>(null)
  const [googleUrl, setGoogleUrl] = useState<string | null>(null)
  const imgRef = useRef<HTMLImageElement>(null)
  const coverColor = getCoverColor(bookId)

  const openLibraryUrl = `https://covers.openlibrary.org/b/isbn/${isbn}-${size}.jpg`

  // When Open Library by ISBN fails, try Open Library by cover ID
  useEffect(() => {
    if (state !== 'ol-cover-id') return
    let cancelled = false
    fetchOLCoverByID(isbn, size).then((url) => {
      if (cancelled) return
      if (url) {
        setOlCoverIdUrl(url)
      } else {
        setState('google-books')
      }
    })
    return () => { cancelled = true }
  }, [state, isbn, size])

  // When Open Library by cover ID fails, try Google Books
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

  const handleImageResult = (success: boolean, img: HTMLImageElement, nextState: CoverState) => {
    if (!success || img.naturalWidth < 20 || img.naturalHeight < 20) {
      setState(nextState)
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

  if (state === 'cover-override') {
    return (
      <img
        ref={imgRef}
        className={className}
        src={coverUrl}
        alt={alt}
        loading={loading}
        onLoad={(e) => handleImageResult(true, e.target as HTMLImageElement, 'open-library')}
        onError={() => setState('open-library')}
      />
    )
  }

  if (state === 'ol-cover-id') {
    if (!olCoverIdUrl) return null
    return (
      <img
        ref={imgRef}
        className={className}
        src={olCoverIdUrl}
        alt={alt}
        loading={loading}
        onLoad={(e) => handleImageResult(true, e.target as HTMLImageElement, 'google-books')}
        onError={() => setState('google-books')}
      />
    )
  }

  if (state === 'google-books') {
    if (!googleUrl) return null
    return (
      <img
        ref={imgRef}
        className={className}
        src={googleUrl}
        alt={alt}
        loading={loading}
        onLoad={(e) => handleImageResult(true, e.target as HTMLImageElement, 'placeholder')}
        onError={() => setState('placeholder')}
      />
    )
  }

  // Default: Open Library by ISBN
  return (
    <img
      ref={imgRef}
      className={className}
      src={openLibraryUrl}
      alt={alt}
      loading={loading}
      onLoad={(e) => handleImageResult(true, e.target as HTMLImageElement, 'ol-cover-id')}
      onError={() => setState('ol-cover-id')}
    />
  )
}
