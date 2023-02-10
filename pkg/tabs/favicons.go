package tabs

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	hash := favIconUrlHash(tab.FavIconUrl)
	files, _ := filepath.Glob(fmt.Sprintf("%s/%s\\.*", FaviconCacheDir, hash))
	if files != nil && len(files) == 1 {
			return files[0], nil
		}
	var (
		data []byte
	    ext string
		err error
	)
	if data, ext, err = this.decode(tab.FavIconUrl); err != nil {
		log.Printf("unable to decode favicon, falling back to sqlite lookup: %s", tab.Url)
		if data, ext, err = this.getFromSqlite(tab.Url); errors.Is(err, sql.ErrNoRows) {
			log.Printf("no favicon found for %s", tab.Url)
			return "", err
		} else if err != nil {
			return "", err
		}
	}
	if ext == "ico" {
		if data, err = convertIcoToPng(data); err != nil {
			return "", fmt.Errorf("failed to convert ico to png: %w", err)
		}
		ext = "png"
	}
	filename := fmt.Sprintf("%s/%s.%s", FaviconCacheDir, hash, ext)
	faviconFile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("Failed to create favicon file: %w", err)
	}
	if _, err := faviconFile.Write(data); err != nil {
		return "", fmt.Errorf("Failed to write favicon: %w", err)
	}
	return filename, nil
}

func (this *FaviconProcessor) getFromSqlite(url string) (data []byte, ext string, err error) {
	row := this.db.QueryRow("select i.data from moz_icons i join moz_icons_to_pages ip join moz_pages_w_icons p where p.id = ip.page_id and ip.icon_id = i.id and p.page_url = ?", url)
	if err := row.Scan(&data); err != nil {
		return nil, "", err
	}
	ext = "svg"
	if string(data[1:4]) == "PNG" {
		ext = "png"
	}
	return
}

var (
	extMap = map[string]string{
		"data:image/x-icon":  "ico",
		"data:image/png":     "png",
		"data:image/svg+xml": "svg",
		"data:image/webp":    "webp",
	}
	pngHeader = []byte{0x89, 0x50, 0x4e, 0x47}
)

func (this *FaviconProcessor) decode(favIconUrl string) (data []byte, ext string, err error) {
	iconType := strings.Split(favIconUrl, ";")[0]
	ext, exists := extMap[iconType]
	if !exists {
		return nil, "", fmt.Errorf("FavIconUrl %s is not base64 decodable png or svg", favIconUrl)
	}
	b64Str := strings.Split(favIconUrl, ",")[1]
	data, err = base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to decode favicon: %w", err)
	}
	return
}

func convertIcoToPng(favIconData []byte) ([]byte, error) {
	var out bytes.Buffer
	cmd := exec.Command("convert", "ico:-", "png:-")
	cmd.Stdin = bytes.NewBuffer(favIconData)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Failed to run `convert`: %w", err)
	}
	pngData := out.Bytes()
	// the resulting output may have several images
	// grab the last one (typically the highest resolution)
	pngPos := 0
	for i := range pngData[:len(pngData) - 4] {
		isHeader := true
		for j := 0; j < 4; j++ {
			if pngData[i+j] != pngHeader[j] {
				isHeader = false
				break
			}
		}
		if isHeader {
			pngPos = i
		}
	}
	return pngData[pngPos:], nil
}

func favIconUrlHash(favIconUrl string) string {
	h := fnv.New64a()
	h.Write([]byte(favIconUrl))
	return fmt.Sprintf("%d", h.Sum64())
}
