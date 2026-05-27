const COVER_COLORS = [
  { bg: '#e8dcc8', title: '#5c3d2e', subtitle: '#7a6352' }, // warm parchment
  { bg: '#d5dce4', title: '#2e3d5c', subtitle: '#52637a' }, // dusty blue
  { bg: '#dce4d5', title: '#3d5c2e', subtitle: '#527a52' }, // sage green
  { bg: '#e4d5dc', title: '#5c2e4a', subtitle: '#7a526a' }, // muted rose
]

export function getCoverColor(bookId: string) {
  let hash = 0
  for (let i = 0; i < bookId.length; i++) {
    hash = ((hash << 5) - hash + bookId.charCodeAt(i)) | 0
  }
  return COVER_COLORS[Math.abs(hash) % COVER_COLORS.length]
}
