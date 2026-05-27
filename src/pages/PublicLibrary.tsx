import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { getCoverColor } from '../lib/coverColors'
import BookCover from '../components/BookCover'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  genre: string
  publisher: string
  dewey_decimal: string
  lc_classification: string
  dewey_guess: string
  lc_guess: string
  cover_url: string
  publish_year: number | null
  loan_status: string
  patron_name?: string
}

type GroupBy = 'none' | 'dewey' | 'loc'
type DisplayMode = 'titles' | 'covers'

const DEWEY_CLASSES: Record<string, string> = {
  '0': '000 — Computer Science, Information & General Works',
  '1': '100 — Philosophy & Psychology',
  '2': '200 — Religion',
  '3': '300 — Social Sciences',
  '4': '400 — Language',
  '5': '500 — Science',
  '6': '600 — Technology',
  '7': '700 — Arts & Recreation',
  '8': '800 — Literature',
  '9': '900 — History & Geography',
}

const LC_CLASSES: Record<string, string> = {
  'A': 'A — General Works',
  'B': 'B — Philosophy, Psychology & Religion',
  'C': 'C — Auxiliary Sciences of History',
  'D': 'D — World History',
  'E': 'E — American History',
  'F': 'F — Americas (Local History)',
  'G': 'G — Geography, Anthropology & Recreation',
  'H': 'H — Social Sciences',
  'J': 'J — Political Science',
  'K': 'K — Law',
  'L': 'L — Education',
  'M': 'M — Music',
  'N': 'N — Fine Arts',
  'P': 'P — Language & Literature',
  'PA': 'PA — Greek & Latin Language & Literature',
  'PB': 'PB — Modern European Languages (Celtic)',
  'PC': 'PC — Romance Languages',
  'PD': 'PD — Germanic & Scandinavian Languages',
  'PE': 'PE — English Language',
  'PF': 'PF — West Germanic Languages',
  'PG': 'PG — Slavic, Baltic & Albanian Languages',
  'PH': 'PH — Uralic & Basque Languages',
  'PJ': 'PJ — Semitic Languages & Literatures',
  'PK': 'PK — Indo-Iranian Languages & Literatures',
  'PL': 'PL — East Asian, African & Oceanian Languages',
  'PM': 'PM — Indigenous American & Artificial Languages',
  'PN': 'PN — Literature (General)',
  'PQ': 'PQ — French, Italian, Spanish & Portuguese Literature',
  'PR': 'PR — English Literature',
  'PS': 'PS — American Literature',
  'PT': 'PT — German, Dutch & Scandinavian Literature',
  'PZ': 'PZ — Fiction & Juvenile Literature',
  'Q': 'Q — Science',
  'QA': 'QA — Mathematics',
  'QB': 'QB — Astronomy',
  'QC': 'QC — Physics',
  'QD': 'QD — Chemistry',
  'QE': 'QE — Geology',
  'QH': 'QH — Natural History & Biology',
  'QK': 'QK — Botany',
  'QL': 'QL — Zoology',
  'QM': 'QM — Human Anatomy',
  'QP': 'QP — Physiology',
  'QR': 'QR — Microbiology',
  'R': 'R — Medicine',
  'S': 'S — Agriculture',
  'T': 'T — Technology',
  'U': 'U — Military Science',
  'V': 'V — Naval Science',
  'Z': 'Z — Bibliography & Library Science',
}

function getDeweyGroup(code: string): string {
  const digit = code.charAt(0)
  return DEWEY_CLASSES[digit] || digit
}

function getLCGroup(code: string): string {
  const two = code.substring(0, 2).toUpperCase()
  if (LC_CLASSES[two]) return LC_CLASSES[two]
  const one = code.charAt(0).toUpperCase()
  return LC_CLASSES[one] || one
}

export default function PublicLibrary() {
  const { slug } = useParams()
  const [books, setBooks] = useState<Book[]>([])
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [genre, setGenre] = useState('')
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)
  const [groupBy, setGroupBy] = useState<GroupBy>('none')
  const [displayMode, setDisplayMode] = useState<DisplayMode>('covers')
  const [flippedCovers, setFlippedCovers] = useState<Set<string>>(new Set())
  const [validCovers, setValidCovers] = useState<Set<string>>(new Set())

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      const params = new URLSearchParams({ slug: slug! })
      if (title) params.set('title', title)
      if (author) params.set('author', author)
      if (genre) params.set('genre', genre)
      const resp = await fetch(`/api/public-library?${params}`)
      if (resp.ok) {
        setBooks(await resp.json())
        setNotFound(false)
      } else if (resp.status === 404) {
        setNotFound(true)
      }
      setLoading(false)
    }
    const timer = setTimeout(load, 300)
    return () => clearTimeout(timer)
  }, [slug, title, author, genre])

  const groupBooks = (books: Book[]): [string, Book[]][] => {
    if (groupBy === 'none') return [['', books]]
    const groups: Record<string, Book[]> = {}
    for (const book of books) {
      const raw = groupBy === 'dewey'
        ? (book.dewey_decimal || book.dewey_guess)
        : (book.lc_classification || book.lc_guess)
      const label = raw
        ? (groupBy === 'dewey' ? getDeweyGroup(raw) : getLCGroup(raw))
        : 'Unclassified'
      if (!groups[label]) groups[label] = []
      groups[label].push(book)
    }
    return Object.entries(groups).sort(([a], [b]) => {
      if (a === 'Unclassified') return 1
      if (b === 'Unclassified') return -1
      return a.localeCompare(b)
    })
  }

  if (notFound) {
    return (
      <div>
        <h2>Library not found</h2>
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>This library link doesn't exist or has been revoked.</p>
      </div>
    )
  }

  return (
    <div>
      <h2>{slug}</h2>
      <div className="filters">
        <input placeholder="Filter by title..." value={title} onChange={(e) => setTitle(e.target.value)} />
        <input placeholder="Filter by author..." value={author} onChange={(e) => setAuthor(e.target.value)} />
        <input placeholder="Filter by genre..." value={genre} onChange={(e) => setGenre(e.target.value)} />
      </div>
      <div className="group-toggle" style={{ margin: '0.5em 0 0.5em', display: 'flex', gap: '1em', alignItems: 'center', flexWrap: 'wrap' }}>
        <span style={{ color: '#8b7355', fontWeight: 600 }}>Display:</span>
        <label><input type="radio" name="displayMode" checked={displayMode === 'titles'} onChange={() => setDisplayMode('titles')} /> Titles</label>
        <label><input type="radio" name="displayMode" checked={displayMode === 'covers'} onChange={() => { setDisplayMode('covers'); setGroupBy('none') }} /> Covers</label>
      </div>
      <div className="group-toggle" style={{ margin: '0 0 1em', display: 'flex', gap: '1em', alignItems: 'center', flexWrap: 'wrap' }}>
        <span style={{ color: '#8b7355', fontWeight: 600 }}>Group by:</span>
        <label><input type="radio" name="groupBy" checked={groupBy === 'none'} onChange={() => setGroupBy('none')} /> None</label>
        <label><input type="radio" name="groupBy" checked={groupBy === 'dewey'} onChange={() => setGroupBy('dewey')} /> Dewey Decimal</label>
        <label><input type="radio" name="groupBy" checked={groupBy === 'loc'} onChange={() => setGroupBy('loc')} /> Library of Congress</label>
      </div>
      {loading ? (
        <div className="loading">Loading books...</div>
      ) : books.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No books found.</p>
      ) : (
        groupBooks(books).map(([group, groupedBooks]) => (
          <div key={group || '_all'}>
            {group && <h3 style={{ color: '#5c3d2e', borderBottom: '1px solid #d4c9b8', paddingBottom: '0.3em', marginTop: '1.5em' }}>{group}</h3>}
            {displayMode === 'titles'
              ? groupedBooks.map((book) => (
                <div key={book.id} className="book-card">
                  <div>
                    <strong>{book.title}</strong>
                    {book.author && <span> — {book.author}</span>}
                  </div>
                  <span className={`status-badge ${book.loan_status}`}>
                    {book.loan_status === 'checked_out'
                      ? `Out to ${book.patron_name || 'someone'}`
                      : 'Available'}
                  </span>
                </div>
              ))
              : (
                <div className="cover-grid">
                  {groupedBooks.map((book) => {
                    const isFlipped = flippedCovers.has(book.id)
                    const hasValidCover = validCovers.has(book.id)
                    const coverColor = getCoverColor(book.id)
                    const toggleFlip = () => {
                      if (!hasValidCover) return
                      setFlippedCovers(prev => {
                        const next = new Set(prev)
                        if (next.has(book.id)) next.delete(book.id)
                        else next.add(book.id)
                        return next
                      })
                    }
                    return (
                      <div
                        key={book.id}
                        className={`cover-grid-item${hasValidCover ? ' flippable' : ''}`}
                        title={book.isbn}
                        style={hasValidCover ? undefined : { cursor: 'default' }}
                        onClick={toggleFlip}
                      >
                        {isFlipped ? (
                          <div className="cover-placeholder" style={{ background: coverColor.bg }}>
                            <span className="cover-placeholder-title" style={{ color: coverColor.title }}>{book.title}</span>
                            {book.author && <span className="cover-placeholder-author" style={{ color: coverColor.subtitle }}>{book.author}</span>}
                            {book.publish_year && <span className="cover-placeholder-year" style={{ color: coverColor.subtitle }}>{book.publish_year}</span>}
                          </div>
                        ) : (
                          <BookCover
                            isbn={book.isbn}
                            bookId={book.id}
                            title={book.title}
                            author={book.author}
                            publishYear={book.publish_year}
                            alt={book.title}
                            coverUrl={book.cover_url || undefined}
                            size="M"
                            loading="lazy"
                            onValidCover={() => {
                              setValidCovers(prev => {
                                if (prev.has(book.id)) return prev
                                const next = new Set(prev)
                                next.add(book.id)
                                return next
                              })
                            }}
                          />
                        )}
                      </div>
                    )
                  })}
                </div>
              )
            }
          </div>
        ))
      )}
    </div>
  )
}
