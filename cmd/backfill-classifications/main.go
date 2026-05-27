package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"home-library/pkg/db"

	"github.com/joho/godotenv"
)

type bookRow struct {
	ID               string
	ISBN             string
	Title            string
	Author           string
	Genre            string
	Publisher        string
	DeweyDecimal     string
	LCClassification string
	DeweyGuess       string
	LCGuess          string
	Email            string
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Preview changes without writing to the database")
	flag.Parse()

	godotenv.Load()

	database, err := db.GetDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		fmt.Println("WARNING: GEMINI_API_KEY not set -- LLM guessing will be skipped.")
	}

	if *dryRun {
		fmt.Println("DRY RUN -- no changes will be written to the database.")
		fmt.Println()
	}

	fmt.Println("Backfilling classification numbers...")
	fmt.Println("Scanning books with missing Dewey Decimal or LoC numbers...")
	fmt.Println()

	books, err := queryMissingBooks(database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query books: %v\n", err)
		os.Exit(1)
	}

	if len(books) == 0 {
		fmt.Println("No books with missing classifications found.")
		printUserStats(database, 0, 0, 0, 0, 0)
		return
	}

	fmt.Printf("Found %d books to process.\n\n", len(books))

	deweyFilled := 0
	locFilled := 0
	deweyGuessed := 0
	locGuessed := 0
	updated := 0

	for i, book := range books {
		fmt.Printf("[%d/%d]  user: %s\n", i+1, len(books), book.Email)
		fmt.Printf("        %q (%s)\n", book.Title, book.ISBN)

		needDewey := book.DeweyDecimal == ""
		needLoC := book.LCClassification == ""
		foundDewey := ""
		foundLoC := ""
		deweySource := ""
		locSource := ""

		// API 1: Open Library Edition API
		d, l := lookupEdition(book.ISBN)
		if needDewey && d != "" {
			foundDewey = d
			deweySource = "Open Library edition"
			needDewey = false
		}
		if needLoC && l != "" {
			foundLoC = l
			locSource = "Open Library edition"
			needLoC = false
		}

		// API 2: Open Library Search by ISBN (if still missing)
		if needDewey || needLoC {
			time.Sleep(1 * time.Second)
			d, l = lookupSearchByISBN(book.ISBN)
			if needDewey && d != "" {
				foundDewey = d
				deweySource = "Open Library search (ISBN)"
				needDewey = false
			}
			if needLoC && l != "" {
				foundLoC = l
				locSource = "Open Library search (ISBN)"
				needLoC = false
			}
		}

		// API 3: Open Library Search by title+author (if still missing)
		if (needDewey || needLoC) && book.Title != "" && book.Author != "" {
			time.Sleep(1 * time.Second)
			d, l = lookupSearchByTitleAuthor(book.Title, book.Author)
			if needDewey && d != "" {
				foundDewey = d
				deweySource = "Open Library search (title+author)"
				needDewey = false
			}
			if needLoC && l != "" {
				foundLoC = l
				locSource = "Open Library search (title+author)"
				needLoC = false
			}
		}

		// API 4: German National Library (DNB) SRU API (if still missing)
		if needDewey || needLoC {
			time.Sleep(1 * time.Second)
			d, l = lookupDNB(book.ISBN)
			if needDewey && d != "" {
				foundDewey = d
				deweySource = "DNB (German National Library)"
				needDewey = false
			}
			if needLoC && l != "" {
				foundLoC = l
				locSource = "DNB (German National Library)"
				needLoC = false
			}
		}

		// Print authoritative results
		if foundDewey != "" {
			fmt.Printf("     \U0001F4D7 Dewey: %s (from %s)\n", foundDewey, deweySource)
		} else if book.DeweyDecimal != "" {
			fmt.Printf("        Dewey: %s (already set)\n", book.DeweyDecimal)
		} else {
			fmt.Printf("        Dewey: (not found)\n")
		}
		if foundLoC != "" {
			fmt.Printf("     \U0001F4D7 LoC:   %s (from %s)\n", foundLoC, locSource)
		} else if book.LCClassification != "" {
			fmt.Printf("        LoC:   %s (already set)\n", book.LCClassification)
		} else {
			fmt.Printf("        LoC:   (not found)\n")
		}

		// API 5: Gemini LLM guess (if still missing AND no existing guess)
		guessedDewey := ""
		guessedLoC := ""
		needDeweyGuess := (book.DeweyDecimal == "" && foundDewey == "" && book.DeweyGuess == "")
		needLoCGuess := (book.LCClassification == "" && foundLoC == "" && book.LCGuess == "")
		if (needDeweyGuess || needLoCGuess) && geminiKey != "" {
			fmt.Printf("        Asking Gemini to guess...\n")
			time.Sleep(4 * time.Second) // Gemini free tier: 15 RPM
			gd, gl := guessWithGemini(geminiKey, book.Title, book.Author, book.Genre, book.Publisher)
			if needDeweyGuess {
				if gd != "" {
					guessedDewey = gd
					fmt.Printf("     \U0001F916 Dewey (guess): %s (from Gemini)\n", guessedDewey)
				} else {
					fmt.Printf("        Dewey (guess): Gemini returned no result\n")
				}
			}
			if needLoCGuess {
				if gl != "" {
					guessedLoC = gl
					fmt.Printf("     \U0001F916 LoC (guess):   %s (from Gemini)\n", guessedLoC)
				} else {
					fmt.Printf("        LoC (guess):   Gemini returned no result\n")
				}
			}
		}

		// Update DB
		hasAuthUpdate := foundDewey != "" || foundLoC != ""
		hasGuessUpdate := guessedDewey != "" || guessedLoC != ""
		if hasAuthUpdate || hasGuessUpdate {
			if !*dryRun {
				err := updateBook(database, book, foundDewey, foundLoC, guessedDewey, guessedLoC)
				if err != nil {
					fmt.Printf("        ERROR updating: %v\n", err)
				}
			}
			if hasAuthUpdate {
				updated++
			}
			if foundDewey != "" {
				deweyFilled++
			}
			if foundLoC != "" {
				locFilled++
			}
			if guessedDewey != "" {
				deweyGuessed++
			}
			if guessedLoC != "" {
				locGuessed++
			}
		}

		fmt.Println()

		// Rate limit: wait before next API call
		if i < len(books)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// Global summary
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Done. Updated %d of %d books (authoritative).\n", updated, len(books))
	fmt.Printf("  Dewey filled: %d\n", deweyFilled)
	fmt.Printf("  LoC filled:   %d\n", locFilled)
	if geminiKey != "" {
		fmt.Printf("  Dewey guessed: %d\n", deweyGuessed)
		fmt.Printf("  LoC guessed:   %d\n", locGuessed)
	}
	fmt.Printf("  Still missing: %d\n", len(books)-updated)
	fmt.Println()

	printUserStats(database, len(books), deweyFilled, locFilled, deweyGuessed, locGuessed)
}

func queryMissingBooks(database *sql.DB) ([]bookRow, error) {
	rows, err := database.Query(`
		SELECT b.id, b.isbn, b.title, COALESCE(b.author, ''),
			   COALESCE(b.genre, ''), COALESCE(b.publisher, ''),
			   COALESCE(b.dewey_decimal, ''), COALESCE(b.lc_classification, ''),
			   COALESCE(b.dewey_guess, ''), COALESCE(b.lc_guess, ''),
			   COALESCE(u.email, '')
		FROM books b
		JOIN users u ON u.id = b.user_id
		WHERE (b.dewey_decimal IS NULL OR b.dewey_decimal = ''
			OR b.lc_classification IS NULL OR b.lc_classification = ''
			OR b.dewey_guess IS NULL OR b.dewey_guess = ''
			OR b.lc_guess IS NULL OR b.lc_guess = '')
		  AND b.isbn NOT LIKE 'MANUAL-%'
		ORDER BY u.email, b.title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []bookRow
	for rows.Next() {
		var b bookRow
		if err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.Publisher, &b.DeweyDecimal, &b.LCClassification, &b.DeweyGuess, &b.LCGuess, &b.Email); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func updateBook(database *sql.DB, book bookRow, dewey, loc, deweyGuess, locGuess string) error {
	newDewey := book.DeweyDecimal
	if dewey != "" {
		newDewey = dewey
	}
	newLoC := book.LCClassification
	if loc != "" {
		newLoC = loc
	}
	newDeweyGuess := book.DeweyGuess
	if deweyGuess != "" {
		newDeweyGuess = deweyGuess
	}
	newLCGuess := book.LCGuess
	if locGuess != "" {
		newLCGuess = locGuess
	}
	_, err := database.Exec(
		`UPDATE books SET dewey_decimal = $1, lc_classification = $2, dewey_guess = $3, lc_guess = $4 WHERE id = $5`,
		newDewey, newLoC, newDeweyGuess, newLCGuess, book.ID,
	)
	return err
}

// lookupEdition fetches Dewey and LoC from Open Library's Edition API.
func lookupEdition(isbn string) (dewey, loc string) {
	resp, err := http.Get(fmt.Sprintf("https://openlibrary.org/isbn/%s.json", isbn))
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", ""
	}

	if arr, ok := data["dewey_decimal_class"].([]interface{}); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			dewey = s
		}
	}
	if arr, ok := data["lc_classifications"].([]interface{}); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			loc = s
		}
	}
	return dewey, loc
}

// lookupSearchByISBN fetches Dewey and LoC from Open Library's Search API by ISBN,
// which aggregates classification data across all editions of a work.
func lookupSearchByISBN(isbn string) (dewey, loc string) {
	return lookupOLSearch(fmt.Sprintf("https://openlibrary.org/search.json?isbn=%s&fields=lcc,ddc&limit=1", isbn))
}

// lookupSearchByTitleAuthor fetches Dewey and LoC from Open Library's Search API
// by title+author, finding classification data from sibling editions of the same work.
func lookupSearchByTitleAuthor(title, author string) (dewey, loc string) {
	q := fmt.Sprintf("https://openlibrary.org/search.json?title=%s&author=%s&fields=lcc,ddc&limit=3",
		url.QueryEscape(title), url.QueryEscape(author))
	return lookupOLSearch(q)
}

func lookupOLSearch(searchURL string) (dewey, loc string) {
	resp, err := http.Get(searchURL)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ""
	}

	var result struct {
		Docs []struct {
			DDC []string `json:"ddc"`
			LCC []string `json:"lcc"`
		} `json:"docs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", ""
	}

	// Search across all returned docs for the first available values
	for _, doc := range result.Docs {
		if dewey == "" && len(doc.DDC) > 0 {
			dewey = cleanSearchClassification(doc.DDC[0])
		}
		if loc == "" && len(doc.LCC) > 0 {
			loc = cleanSearchClassification(doc.LCC[0])
		}
		if dewey != "" && loc != "" {
			break
		}
	}
	return dewey, loc
}

// cleanSearchClassification converts the padded format from the Search API
// (e.g. "PG-3326.00000000.P7 2003") back to a standard form (e.g. "PG3326.P7 2003").
func cleanSearchClassification(raw string) string {
	cleaned := strings.ReplaceAll(raw, ".00000000", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return strings.TrimSpace(cleaned)
}

// --- DNB (German National Library) SRU API ---

type sruResponse struct {
	XMLName xml.Name    `xml:"searchRetrieveResponse"`
	Records []sruRecord `xml:"records>record"`
}

type sruRecord struct {
	RecordData sruRecordData `xml:"recordData"`
}

type sruRecordData struct {
	Record marcRecord `xml:"record"`
}

type marcRecord struct {
	DataFields []marcDataField `xml:"datafield"`
}

type marcDataField struct {
	Tag       string         `xml:"tag,attr"`
	SubFields []marcSubField `xml:"subfield"`
}

type marcSubField struct {
	Code  string `xml:"code,attr"`
	Value string `xml:",chardata"`
}

func lookupDNB(isbn string) (dewey, loc string) {
	dnbURL := fmt.Sprintf(
		"https://services.dnb.de/sru/dnb?version=1.1&operation=searchRetrieve&query=isbn%%3D%s&maximumRecords=1&recordSchema=MARC21-xml",
		isbn,
	)
	resp, err := http.Get(dnbURL)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ""
	}

	var sru sruResponse
	if err := xml.Unmarshal(body, &sru); err != nil {
		return "", ""
	}

	if len(sru.Records) == 0 {
		return "", ""
	}

	record := sru.Records[0].RecordData.Record
	sdnbDewey := ""

	for _, df := range record.DataFields {
		switch df.Tag {
		case "082":
			if dewey == "" {
				for _, sf := range df.SubFields {
					if sf.Code == "a" && sf.Value != "" {
						dewey = sf.Value
						break
					}
				}
			}
		case "050":
			if loc == "" {
				for _, sf := range df.SubFields {
					if sf.Code == "a" && sf.Value != "" {
						loc = sf.Value
						break
					}
				}
			}
		case "084":
			if sdnbDewey == "" {
				isSdnb := false
				firstValue := ""
				for _, sf := range df.SubFields {
					if sf.Code == "2" && sf.Value == "sdnb" {
						isSdnb = true
					}
					if sf.Code == "a" && firstValue == "" {
						firstValue = sf.Value
					}
				}
				if isSdnb && firstValue != "" {
					sdnbDewey = firstValue
				}
			}
		}
	}

	if dewey == "" && sdnbDewey != "" {
		dewey = sdnbDewey
	}

	return dewey, loc
}

// --- Gemini LLM API ---

func guessWithGemini(apiKey, title, author, genre, publisher string) (dewey, loc string) {
	prompt := fmt.Sprintf(`You are a librarian. Given this book's metadata, suggest the most likely Dewey Decimal Classification number and Library of Congress Classification.
Return ONLY valid JSON with no markdown formatting: {"dewey": "...", "loc": "..."}
For dewey, provide a specific number like "823.914" not just a broad class.
For loc, provide a specific classification like "PR6068.O93" not just a letter.
If you cannot determine a value, use an empty string.

Title: %s
Author: %s
Genre: %s
Publisher: %s`, title, author, genre, publisher)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.1,
			"maxOutputTokens": 100,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", ""
	}

	geminiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey)

	var body []byte
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := http.Post(geminiURL, "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "        Gemini request error: %v\n", err)
			return "", ""
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			break
		}
		if resp.StatusCode == 429 {
			wait := 5 * (attempt + 1)
			fmt.Printf("        Gemini rate limited, waiting %ds...\n", wait)
			time.Sleep(time.Duration(wait) * time.Second)
			continue
		}
		fmt.Fprintf(os.Stderr, "        Gemini error (HTTP %d): %s\n", resp.StatusCode, string(body))
		return "", ""
	}

	if body == nil {
		return "", ""
	}

	// Parse Gemini response structure
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", ""
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", ""
	}

	text := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	// Strip markdown code fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result struct {
		Dewey string `json:"dewey"`
		Loc   string `json:"loc"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return "", ""
	}

	return result.Dewey, result.Loc
}

func printUserStats(database *sql.DB, processed, newDewey, newLoC, newDeweyGuess, newLoCGuess int) {
	if processed > 0 {
		fmt.Println("New values added this run:")
		fmt.Printf("  Dewey (authoritative): %d/%d (%d%%)\n", newDewey, processed, pct(newDewey, processed))
		fmt.Printf("  Dewey (guessed):       %d/%d (%d%%)\n", newDeweyGuess, processed, pct(newDeweyGuess, processed))
		fmt.Printf("  LoC (authoritative):   %d/%d (%d%%)\n", newLoC, processed, pct(newLoC, processed))
		fmt.Printf("  LoC (guessed):         %d/%d (%d%%)\n", newLoCGuess, processed, pct(newLoCGuess, processed))
		fmt.Println()
	}

	rows, err := database.Query(`
		SELECT COALESCE(u.email, '(no email)'),
			   COUNT(*) AS total,
			   COUNT(NULLIF(b.dewey_decimal, '')) AS has_dewey,
			   COUNT(NULLIF(b.lc_classification, '')) AS has_loc,
			   COUNT(NULLIF(b.dewey_guess, '')) AS has_dewey_guess,
			   COUNT(NULLIF(b.lc_guess, '')) AS has_lc_guess
		FROM books b
		JOIN users u ON u.id = b.user_id
		WHERE b.isbn NOT LIKE 'MANUAL-%'
		GROUP BY u.email
		ORDER BY u.email
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query user stats: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Per-user classification stats (overall library):")
	fmt.Println()
	for rows.Next() {
		var email string
		var total, hasDewey, hasLoC, hasDeweyGuess, hasLCGuess int
		if err := rows.Scan(&email, &total, &hasDewey, &hasLoC, &hasDeweyGuess, &hasLCGuess); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan stats row: %v\n", err)
			continue
		}
		deweyTotal := hasDewey + hasDeweyGuess
		if deweyTotal > total {
			deweyTotal = total
		}
		locTotal := hasLoC + hasLCGuess
		if locTotal > total {
			locTotal = total
		}
		fmt.Printf("  %s (%d books total)\n", email, total)
		fmt.Printf("    Dewey:  %d/%d (%d%%) authoritative, %d/%d (%d%%) with guesses\n",
			hasDewey, total, pct(hasDewey, total), deweyTotal, total, pct(deweyTotal, total))
		fmt.Printf("    LoC:    %d/%d (%d%%) authoritative, %d/%d (%d%%) with guesses\n",
			hasLoC, total, pct(hasLoC, total), locTotal, total, pct(locTotal, total))
		fmt.Println()
	}
}

func pct(n, total int) int {
	if total == 0 {
		return 0
	}
	return n * 100 / total
}
