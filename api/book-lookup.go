package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"home-library/pkg/auth"
)

type BookLookupResult struct {
	ISBN        string `json:"isbn"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Genre       string `json:"genre"`
	PublishYear *int   `json:"publish_year"`
}

func BookLookup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	_, _, err := auth.ValidateRequest(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	isbn := r.URL.Query().Get("isbn")
	if isbn == "" {
		http.Error(w, `{"error":"isbn parameter required"}`, http.StatusBadRequest)
		return
	}

	result, err := lookupISBN(isbn)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func lookupISBN(isbn string) (*BookLookupResult, error) {
	// Fetch book data from Open Library
	resp, err := http.Get(fmt.Sprintf("https://openlibrary.org/isbn/%s.json", isbn))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book data")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("book not found for ISBN %s", isbn)
	}

	body, _ := io.ReadAll(resp.Body)
	var bookData map[string]interface{}
	json.Unmarshal(body, &bookData)

	result := &BookLookupResult{ISBN: isbn}

	// Title
	if title, ok := bookData["title"].(string); ok {
		result.Title = title
	}

	// Publish year
	if pubDate, ok := bookData["publish_date"].(string); ok {
		var year int
		if _, err := fmt.Sscanf(pubDate, "%d", &year); err == nil && year > 1000 {
			result.PublishYear = &year
		} else {
			// Try to find a 4-digit year in the string
			for _, part := range strings.Fields(pubDate) {
				part = strings.Trim(part, ",.")
				if len(part) == 4 {
					if _, err := fmt.Sscanf(part, "%d", &year); err == nil && year > 1000 {
						result.PublishYear = &year
						break
					}
				}
			}
		}
	}

	// Author - try multiple formats
	// Format 1: "authors" array with {key: "/authors/..."} objects
	if authors, ok := bookData["authors"].([]interface{}); ok && len(authors) > 0 {
		if authorObj, ok := authors[0].(map[string]interface{}); ok {
			if key, ok := authorObj["key"].(string); ok {
				authorResp, err := http.Get(fmt.Sprintf("https://openlibrary.org%s.json", key))
				if err == nil {
					defer authorResp.Body.Close()
					var authorData map[string]interface{}
					authorBody, _ := io.ReadAll(authorResp.Body)
					json.Unmarshal(authorBody, &authorData)
					if name, ok := authorData["name"].(string); ok {
						result.Author = name
					}
				}
			}
		}
	}
	// Format 2: "author" array of strings (e.g. ["Austen, Jane, 1775-1817."])
	if result.Author == "" {
		if authorList, ok := bookData["author"].([]interface{}); ok && len(authorList) > 0 {
			if name, ok := authorList[0].(string); ok {
				// Clean up: remove dates like ", 1775-1817." from the name
				parts := strings.Split(name, ",")
				if len(parts) >= 2 {
					// "Last, First, dates" -> "First Last"
					result.Author = strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
				} else {
					result.Author = strings.TrimRight(name, ".")
				}
			}
		}
	}
	// Format 3: try getting author from the works endpoint
	if result.Author == "" {
		if works, ok := bookData["works"].([]interface{}); ok && len(works) > 0 {
			if workObj, ok := works[0].(map[string]interface{}); ok {
				if key, ok := workObj["key"].(string); ok {
					workResp, err := http.Get(fmt.Sprintf("https://openlibrary.org%s.json", key))
					if err == nil {
						defer workResp.Body.Close()
						var workData map[string]interface{}
						workBody, _ := io.ReadAll(workResp.Body)
						json.Unmarshal(workBody, &workData)
						if wAuthors, ok := workData["authors"].([]interface{}); ok && len(wAuthors) > 0 {
							if aObj, ok := wAuthors[0].(map[string]interface{}); ok {
								if authorRef, ok := aObj["author"].(map[string]interface{}); ok {
									if aKey, ok := authorRef["key"].(string); ok {
										aResp, err := http.Get(fmt.Sprintf("https://openlibrary.org%s.json", aKey))
										if err == nil {
											defer aResp.Body.Close()
											var aData map[string]interface{}
											aBody, _ := io.ReadAll(aResp.Body)
											json.Unmarshal(aBody, &aData)
											if name, ok := aData["name"].(string); ok {
												result.Author = name
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Genre/subjects - follow works key
	if works, ok := bookData["works"].([]interface{}); ok && len(works) > 0 {
		if workObj, ok := works[0].(map[string]interface{}); ok {
			if key, ok := workObj["key"].(string); ok {
				workResp, err := http.Get(fmt.Sprintf("https://openlibrary.org%s.json", key))
				if err == nil {
					defer workResp.Body.Close()
					var workData map[string]interface{}
					workBody, _ := io.ReadAll(workResp.Body)
					json.Unmarshal(workBody, &workData)
					if subjects, ok := workData["subjects"].([]interface{}); ok && len(subjects) > 0 {
						var genreParts []string
						limit := 3
						if len(subjects) < limit {
							limit = len(subjects)
						}
						for _, s := range subjects[:limit] {
							if str, ok := s.(string); ok {
								genreParts = append(genreParts, str)
							}
						}
						result.Genre = strings.Join(genreParts, ", ")
					}
				}
			}
		}
	}

	if result.Title == "" {
		return nil, fmt.Errorf("could not parse book data for ISBN %s", isbn)
	}

	return result, nil
}
