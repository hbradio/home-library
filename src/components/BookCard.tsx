import { useNavigate } from 'react-router-dom'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  loan_status: string
  patron_name?: string
}

export default function BookCard({ book }: { book: Book }) {
  const navigate = useNavigate()

  return (
    <div className="book-card" onClick={() => navigate(`/book/${book.id}`)} style={{ cursor: 'pointer' }}>
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
  )
}
