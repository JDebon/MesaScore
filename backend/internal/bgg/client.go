package bgg

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://boardgamegeek.com/xmlapi2"

type SearchResult struct {
	BggID        int     `json:"bgg_id"`
	Name         string  `json:"name"`
	Year         *int    `json:"year"`
	ThumbnailURL *string `json:"thumbnail_url"`
}

type GameDetail struct {
	BggID         int      `json:"bgg_id"`
	Name          string   `json:"name"`
	Year          *int     `json:"year"`
	Description   *string  `json:"description"`
	CoverImageURL *string  `json:"cover_image_url"`
	MinPlayers    *int     `json:"min_players"`
	MaxPlayers    *int     `json:"max_players"`
	BggRating     *float64 `json:"bgg_rating"`
}

// XML structures for BGG API
type searchResponse struct {
	XMLName xml.Name     `xml:"items"`
	Items   []searchItem `xml:"item"`
}

type searchItem struct {
	ID   int `xml:"id,attr"`
	Name struct {
		Value string `xml:"value,attr"`
	} `xml:"name"`
	YearPublished struct {
		Value string `xml:"value,attr"`
	} `xml:"yearpublished"`
}

type thingResponse struct {
	XMLName xml.Name    `xml:"items"`
	Items   []thingItem `xml:"item"`
}

type thingItem struct {
	ID        int    `xml:"id,attr"`
	Thumbnail string `xml:"thumbnail"`
	Image     string `xml:"image"`
	Names     []struct {
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"name"`
	Description   string `xml:"description"`
	YearPublished struct {
		Value string `xml:"value,attr"`
	} `xml:"yearpublished"`
	MinPlayers  struct {
		Value string `xml:"value,attr"`
	} `xml:"minplayers"`
	MaxPlayers struct {
		Value string `xml:"value,attr"`
	} `xml:"maxplayers"`
	Statistics struct {
		Ratings struct {
			Average struct {
				Value string `xml:"value,attr"`
			} `xml:"average"`
		} `xml:"ratings"`
	} `xml:"statistics"`
}

var apiToken string

// Init sets the BGG API bearer token.
func Init(token string) {
	apiToken = token
}

func get(reqURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+apiToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("BGG API returned 401 — check BGG_API_TOKEN")
	}
	return resp, nil
}

func Search(query string) ([]SearchResult, error) {
	resp, err := get(fmt.Sprintf("%s/search?query=%s&type=boardgame", baseURL, url.QueryEscape(query)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sr searchResponse
	if err := xml.Unmarshal(body, &sr); err != nil {
		return nil, err
	}

	// Limit to first 20 results (BGG /thing accepts max 20 IDs per call)
	items := sr.Items
	if len(items) > 20 {
		items = items[:20]
	}

	results := make([]SearchResult, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		r := SearchResult{
			BggID: item.ID,
			Name:  item.Name.Value,
		}
		if item.YearPublished.Value != "" {
			if y, err := strconv.Atoi(item.YearPublished.Value); err == nil {
				r.Year = &y
			}
		}
		results = append(results, r)
		ids = append(ids, strconv.Itoa(item.ID))
	}

	// Batch-fetch thumbnails from /thing
	if len(ids) > 0 {
		thumbnails := fetchThumbnails(ids)
		for i := range results {
			if u, ok := thumbnails[results[i].BggID]; ok {
				results[i].ThumbnailURL = &u
			}
		}
	}

	// Server-side ranking per spec F6.1
	lq := strings.ToLower(strings.TrimSpace(query))
	tier := func(name string) int {
		lower := strings.ToLower(name)
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(query)) {
			return 0
		}
		if strings.HasPrefix(lower, lq) && !strings.Contains(name, ":") {
			return 1
		}
		if strings.Contains(lower, lq) && !strings.Contains(name, ":") {
			return 2
		}
		return 3
	}
	sort.SliceStable(results, func(i, j int) bool {
		ti, tj := tier(results[i].Name), tier(results[j].Name)
		if ti != tj {
			return ti < tj
		}
		// Within tier: year ascending, nil last
		yi, yj := results[i].Year, results[j].Year
		if yi == nil && yj == nil {
			return false
		}
		if yi == nil {
			return false
		}
		if yj == nil {
			return true
		}
		return *yi < *yj
	})

	return results, nil
}

// fetchThumbnails calls /thing for a batch of IDs and returns a map of bgg_id → thumbnail URL.
func fetchThumbnails(ids []string) map[int]string {
	batchURL := fmt.Sprintf("%s/thing?id=%s", baseURL, strings.Join(ids, ","))
	resp, err := get(batchURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var tr thingResponse
	if err := xml.Unmarshal(body, &tr); err != nil {
		return nil
	}

	m := make(map[int]string, len(tr.Items))
	for _, item := range tr.Items {
		if item.Thumbnail != "" {
			m[item.ID] = item.Thumbnail
		}
	}
	return m
}

func GetThing(bggID int) (*GameDetail, error) {
	fetchURL := fmt.Sprintf("%s/thing?id=%d&stats=1", baseURL, bggID)

	resp, err := get(fetchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// BGG may return 202 on first call — retry once after 2 seconds
	if resp.StatusCode == http.StatusAccepted {
		resp.Body.Close()
		time.Sleep(2 * time.Second)
		resp, err = get(fetchURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BGG returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tr thingResponse
	if err := xml.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	if len(tr.Items) == 0 {
		return nil, fmt.Errorf("game not found on BGG")
	}

	item := tr.Items[0]
	gd := &GameDetail{
		BggID: item.ID,
	}

	for _, n := range item.Names {
		if n.Type == "primary" {
			gd.Name = n.Value
			break
		}
	}

	if item.YearPublished.Value != "" {
		if y, err := strconv.Atoi(item.YearPublished.Value); err == nil {
			gd.Year = &y
		}
	}
	if item.Description != "" {
		gd.Description = &item.Description
	}
	if item.Image != "" {
		gd.CoverImageURL = &item.Image
	}

	if v, err := strconv.Atoi(item.MinPlayers.Value); err == nil {
		gd.MinPlayers = &v
	}
	if v, err := strconv.Atoi(item.MaxPlayers.Value); err == nil {
		gd.MaxPlayers = &v
	}
	if v, err := strconv.ParseFloat(item.Statistics.Ratings.Average.Value, 64); err == nil {
		gd.BggRating = &v
	}

	return gd, nil
}
