import { useState, useEffect } from 'react'
import { useApi } from '../lib/api'
import BookCard from '../components/BookCard'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  genre: string
  publisher: string
  dewey_decimal: string
  lc_classification: string
  publish_year: number | null
  loan_status: string
  patron_name?: string
}

type GroupBy = 'none' | 'dewey' | 'loc'

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
  // Try 2-letter match first, then 1-letter
  const two = code.substring(0, 2).toUpperCase()
  if (LC_CLASSES[two]) return LC_CLASSES[two]
  const one = code.charAt(0).toUpperCase()
  return LC_CLASSES[one] || one
}

export default function Browse() {
  const { fetchWithAuth } = useApi()
  const [books, setBooks] = useState<Book[]>([])
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [genre, setGenre] = useState('')
  const [loading, setLoading] = useState(true)
  const [groupBy, setGroupBy] = useState<GroupBy>('loc')

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      const params = new URLSearchParams()
      if (title) params.set('title', title)
      if (author) params.set('author', author)
      if (genre) params.set('genre', genre)
      const resp = await fetchWithAuth(`/api/books?${params}`)
      if (resp.ok) setBooks(await resp.json())
      setLoading(false)
    }
    const timer = setTimeout(load, 300)
    return () => clearTimeout(timer)
  }, [title, author, genre, fetchWithAuth])

  const exportCsv = async () => {
    const resp = await fetchWithAuth('/api/books')
    if (!resp.ok) return
    const allBooks: Book[] = await resp.json()
    const header = 'ISBN,Title,Author,Genre,Publisher,Dewey Decimal,LoC,Year,Status'
    const escape = (s: string) => `"${(s || '').replace(/"/g, '""')}"`
    const rows = allBooks.map(b =>
      [escape(b.isbn), escape(b.title), escape(b.author), escape(b.genre), escape(b.publisher), escape(b.dewey_decimal), escape(b.lc_classification), b.publish_year ?? '', b.loan_status].join(',')
    )
    const csv = [header, ...rows].join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'library.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  const groupBooks = (books: Book[]): [string, Book[]][] => {
    if (groupBy === 'none') return [['', books]]
    const groups: Record<string, Book[]> = {}
    for (const book of books) {
      const raw = groupBy === 'dewey' ? book.dewey_decimal : book.lc_classification
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

  return (
    <div>
      <h2>Browse Library <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <div className="filters">
        <input placeholder="Filter by title..." value={title} onChange={(e) => setTitle(e.target.value)} />
        <input placeholder="Filter by author..." value={author} onChange={(e) => setAuthor(e.target.value)} />
        <input placeholder="Filter by genre..." value={genre} onChange={(e) => setGenre(e.target.value)} />
        <button onClick={exportCsv} disabled={loading || books.length === 0}>Export CSV</button>
      </div>
      <div className="group-toggle" style={{ margin: '0.5em 0 1em', display: 'flex', gap: '1em', alignItems: 'center', flexWrap: 'wrap' }}>
        <span style={{ color: '#8b7355', fontWeight: 600 }}>Group by:</span>
        <label><input type="radio" name="groupBy" checked={groupBy === 'none'} onChange={() => setGroupBy('none')} /> None</label>
        <label><input type="radio" name="groupBy" checked={groupBy === 'dewey'} onChange={() => setGroupBy('dewey')} /> Dewey Decimal</label>
        <label><input type="radio" name="groupBy" checked={groupBy === 'loc'} onChange={() => setGroupBy('loc')} /> Library of Congress</label>
      </div>
      {loading ? (
        <div className="loading">Loading books...</div>
      ) : books.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No books found. Add some from the home screen.</p>
      ) : (
        groupBooks(books).map(([group, groupedBooks]) => (
          <div key={group || '_all'}>
            {group && <h3 style={{ color: '#5c3d2e', borderBottom: '1px solid #d4c9b8', paddingBottom: '0.3em', marginTop: '1.5em' }}>{group}</h3>}
            {groupedBooks.map((book) => <BookCard key={book.id} book={book} />)}
          </div>
        ))
      )}
    </div>
  )
}
