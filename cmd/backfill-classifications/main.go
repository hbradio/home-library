package main

import (
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
	DeweyDecimal     string
	LCClassification string
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
		printUserStats(database, 0, 0, 0)
		return
	}

	fmt.Printf("Found %d books to process.\n\n", len(books))

	deweyFilled := 0
	locFilled := 0
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

		// Print results
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

		// Update DB
		if foundDewey != "" || foundLoC != "" {
			if !*dryRun {
				err := updateBook(database, book, foundDewey, foundLoC)
				if err != nil {
					fmt.Printf("        ERROR updating: %v\n", err)
				}
			}
			updated++
			if foundDewey != "" {
				deweyFilled++
			}
			if foundLoC != "" {
				locFilled++
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
	fmt.Printf("Done. Updated %d of %d books.\n", updated, len(books))
	fmt.Printf("  Dewey filled: %d\n", deweyFilled)
	fmt.Printf("  LoC filled:   %d\n", locFilled)
	fmt.Printf("  Still missing: %d\n", len(books)-updated)
	fmt.Println()

	printUserStats(database, len(books), deweyFilled, locFilled)
}

func queryMissingBooks(database *sql.DB) ([]bookRow, error) {
	rows, err := database.Query(`
		SELECT b.id, b.isbn, b.title, COALESCE(b.author, ''),
			   COALESCE(b.dewey_decimal, ''), COALESCE(b.lc_classification, ''),
			   COALESCE(u.email, '')
		FROM books b
		JOIN users u ON u.id = b.user_id
		WHERE (b.dewey_decimal IS NULL OR b.dewey_decimal = ''
			OR b.lc_classification IS NULL OR b.lc_classification = '')
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
		if err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.DeweyDecimal, &b.LCClassification, &b.Email); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func updateBook(database *sql.DB, book bookRow, dewey, loc string) error {
	newDewey := book.DeweyDecimal
	if dewey != "" {
		newDewey = dewey
	}
	newLoC := book.LCClassification
	if loc != "" {
		newLoC = loc
	}
	_, err := database.Exec(
		`UPDATE books SET dewey_decimal = $1, lc_classification = $2 WHERE id = $3`,
		newDewey, newLoC, book.ID,
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
	// Remove ".00000000" padding
	cleaned := strings.ReplaceAll(raw, ".00000000", "")
	// Remove dashes (e.g. "PG-3326" -> "PG3326")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	// Clean up any double spaces
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return strings.TrimSpace(cleaned)
}

// --- DNB (German National Library) SRU API ---

// MARC21 XML structures for parsing DNB responses
type sruResponse struct {
	XMLName xml.Name  `xml:"searchRetrieveResponse"`
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
	Tag       string          `xml:"tag,attr"`
	SubFields []marcSubField  `xml:"subfield"`
}

type marcSubField struct {
	Code  string `xml:"code,attr"`
	Value string `xml:",chardata"`
}

// lookupDNB fetches classification data from the German National Library's SRU API.
// It checks MARC tag 082 for Dewey and tag 050 for LoC classification.
// If those are missing, it falls back to tag 084 with indicator "sdnb"
// (Sachgruppen der DNB), which maps to broad Dewey classes.
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
		case "082": // Standard Dewey Decimal
			if dewey == "" {
				for _, sf := range df.SubFields {
					if sf.Code == "a" && sf.Value != "" {
						dewey = sf.Value
						break
					}
				}
			}
		case "050": // Library of Congress Classification
			if loc == "" {
				for _, sf := range df.SubFields {
					if sf.Code == "a" && sf.Value != "" {
						loc = sf.Value
						break
					}
				}
			}
		case "084": // Other classification (check for sdnb)
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

	// Fall back to sdnb code if no standard Dewey was found
	if dewey == "" && sdnbDewey != "" {
		dewey = sdnbDewey
	}

	return dewey, loc
}

func printUserStats(database *sql.DB, processed, newDewey, newLoC int) {
	// Print new-fills summary if we processed anything
	if processed > 0 {
		fmt.Println("New values added this run:")
		deweyPct := 0
		locPct := 0
		if processed > 0 {
			deweyPct = newDewey * 100 / processed
			locPct = newLoC * 100 / processed
		}
		fmt.Printf("  Dewey: %d/%d processed (%d%%)\n", newDewey, processed, deweyPct)
		fmt.Printf("  LoC:   %d/%d processed (%d%%)\n", newLoC, processed, locPct)
		fmt.Println()
	}

	rows, err := database.Query(`
		SELECT COALESCE(u.email, '(no email)'),
			   COUNT(*) AS total,
			   COUNT(NULLIF(b.dewey_decimal, '')) AS has_dewey,
			   COUNT(NULLIF(b.lc_classification, '')) AS has_loc
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
		var total, hasDewey, hasLoC int
		if err := rows.Scan(&email, &total, &hasDewey, &hasLoC); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan stats row: %v\n", err)
			continue
		}
		deweyPct := 0
		locPct := 0
		if total > 0 {
			deweyPct = hasDewey * 100 / total
			locPct = hasLoC * 100 / total
		}
		fmt.Printf("  %s (%d books total)\n", email, total)
		fmt.Printf("    Dewey Decimal:  %d/%d (%d%%)\n", hasDewey, total, deweyPct)
		fmt.Printf("    LoC:            %d/%d (%d%%)\n", hasLoC, total, locPct)
		fmt.Println()
	}
}
