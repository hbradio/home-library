import { useState, useEffect } from 'react'
import { useApi } from '../lib/api'
import BookCard from '../components/BookCard'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  genre: string
  publish_year: number | null
  loan_status: string
  patron_name?: string
}

export default function Browse() {
  const { fetchWithAuth } = useApi()
  const [books, setBooks] = useState<Book[]>([])
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [genre, setGenre] = useState('')
  const [loading, setLoading] = useState(true)

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
    const header = 'ISBN,Title,Author,Genre,Year,Status'
    const escape = (s: string) => `"${(s || '').replace(/"/g, '""')}"`
    const rows = allBooks.map(b =>
      [escape(b.isbn), escape(b.title), escape(b.author), escape(b.genre), b.publish_year ?? '', b.loan_status].join(',')
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

  return (
    <div>
      <h2>Browse Library <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <div className="filters">
        <input placeholder="Filter by title..." value={title} onChange={(e) => setTitle(e.target.value)} />
        <input placeholder="Filter by author..." value={author} onChange={(e) => setAuthor(e.target.value)} />
        <input placeholder="Filter by genre..." value={genre} onChange={(e) => setGenre(e.target.value)} />
        <button onClick={exportCsv} disabled={loading || books.length === 0}>Export CSV</button>
      </div>
      {loading ? (
        <div className="loading">Loading books...</div>
      ) : books.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No books found. Add some from the home screen.</p>
      ) : (
        books.map((book) => <BookCard key={book.id} book={book} />)
      )}
    </div>
  )
}
