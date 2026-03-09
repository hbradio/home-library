# Home Library

A personal book library management app. Log in with Google or email, scan books by ISBN (keyboard or camera), browse your collection, and loan/return books to friends.

**Live:** https://home-library-ten.vercel.app

## Tech Stack

- **Frontend:** React + TypeScript (Vite)
- **Backend:** Go serverless functions (Vercel)
- **Database:** CockroachDB Cloud (PostgreSQL-compatible)
- **Auth:** Auth0 (Google social + email/password)
- **Book Data:** Open Library API

## Features

- **ISBN Scanning** - USB barcode scanner support (auto-focused input) + phone camera via html5-qrcode
- **Book Lookup** - Automatic metadata from Open Library (title, author, genre, cover image)
- **Loan/Return** - Scan a book to auto-detect checkout vs. return. Event-sourced loan history
- **Patron Management** - Track who has your books
- **Filtering** - Browse by title, author, or genre
- **Multi-user** - Each user has their own isolated library

## Local Development

```bash
# Prerequisites: Docker, Go 1.21+, Node 18+

# Start everything (Docker PostgreSQL + Go API + Vite)
npm run dev:all

# Or run individually:
npm run dev:docker    # Start PostgreSQL in Docker
npm run dev:api       # Start Go API on :8089
npm run dev           # Start Vite on :5179
```

Create a `.env` file:
```
VITE_AUTH0_DOMAIN=your-auth0-domain
VITE_AUTH0_CLIENT_ID=your-spa-client-id
VITE_AUTH0_AUDIENCE=your-api-identifier
AUTH0_DOMAIN=your-auth0-domain
AUTH0_AUDIENCE=your-api-identifier
DATABASE_URL=postgresql://user:pass@localhost:5433/home_library?sslmode=disable
```

## API Endpoints

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/api/health` | GET | Health check |
| `/api/user` | GET | Get/create authenticated user |
| `/api/book-lookup?isbn=X` | GET | Look up book metadata from Open Library |
| `/api/books` | GET, POST, DELETE | CRUD for books (with filter params) |
| `/api/patrons` | GET, POST, DELETE | CRUD for patrons |
| `/api/loans` | GET, POST | Loan history + smart checkout/return |

## Keyboard Shortcuts

- **ESC** - Return to home screen from any page
- **Enter** - Submit ISBN in scan input
