package tabs

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	FaviconCacheDir = "/tmp/favicons"
)

type FaviconProcessor struct {
	db *sql.DB
}

func makeFaviconProcessor() *FaviconProcessor {
	db, err := sql.Open("sqlite3", FirefoxProfile+"/favicons.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return &FaviconProcessor{db: db}
}

// retrieves the favicon from the sqlite db if exists, or decodes FavIconUrl
// writes the data to a file if it does not already exist
// returns the filename
func (this *FaviconProcessor) Process(tab *Tab) (string, error) {
	if filename, err := this.getFromSqlite(tab.Url); errors.Is(err, sql.ErrNoRows) {
	// if lookup fails, we decode and potentially convert Favicon string from browser
		return this.decode(tab.FavIconUrl)
	} else if err != nil {
		return "", err
	} else {
		return filename, nil
	}
}

func (this *FaviconProcessor) getFromSqlite(url string) (string, error) {
	row := this.db.QueryRow("select i.fixed_icon_url_hash, i.data from moz_icons i join moz_icons_to_pages ip join moz_pages_w_icons p where p.id = ip.page_id and ip.icon_id = i.id and p.page_url = ?", url)
	var hash int
	var data []byte
	if err := row.Scan(&hash, &data); err != nil {
		return "", err
	}
	ext := "svg"
	if string(data[1:4]) == "PNG" {
		ext = "png"
	}
	filename := fmt.Sprintf("%s/%d.%s", FaviconCacheDir, hash, ext)
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		return "", err
	}
	return filename, nil
}

func (this *FaviconProcessor) decode(favIconUrl string) (string, error) {
	suffixMap := map[string]string{
		"data:image/x-icon":  "*.ico",
		"data:image/png":     "*.png",
		"data:image/svg+xml": "*.svg",
		"data:image/webp":    "*.webp",
	}
	iconType := strings.Split(favIconUrl, ";")[0]
	suffix, exists := suffixMap[iconType]
	if !exists {
		return "", fmt.Errorf("FavIconUrl %s is not base64 decodable png or svg", favIconUrl)
	}
	b64Str := strings.Split(favIconUrl, ",")[1]
	data, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", fmt.Errorf("Failed to decode favicon: %w", err)
	}
	faviconFile, err := os.CreateTemp(FaviconCacheDir, suffix)
	if err != nil {
		return "", fmt.Errorf("Failed to create favicon file: %w", err)
	}
	if _, err := faviconFile.Write(data); err != nil {
		return "", fmt.Errorf("Failed to write favicon: %w", err)
	}
	if suffix == "*.ico" {
		pngFilename := strings.ReplaceAll(faviconFile.Name(), ".ico", ".png")
		cmd := exec.Command("convert", "-verbose", faviconFile.Name(), pngFilename)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("Failed to read conversion output: %w", err)
		} else {
			m := regexp.MustCompile(faviconFile.Name()[:len(faviconFile.Name())-len(suffix)+1] + `(?:-\d+)?\.png`)
			lines := strings.Split(string(out), "\n")
			largest := lines[len(lines)-2]
			pngFilename = m.FindString(largest)
		}
		return pngFilename, nil
	}
	return faviconFile.Name(), nil
}
