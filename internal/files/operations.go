package nomnom

import (
	"fmt"
	"image/jpeg"

	"os"
	"path/filepath"
	"strings"

	log "log"

	"github.com/dhowden/tag"
	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
)

func ReadFile(path string) (string, error) {

	extension := GetFileExtension(path)
	extension = strings.TrimPrefix(extension, ".")

	if extension == "txt" || extension == "md" || extension == "json" {
		content, err := readRawFile(path)
		if err != nil {
			return "There was an error reading the file" + path, err
		}
		return content, nil
	}

	if extension == "png" || extension == "jpg" || extension == "jpeg" || extension == ".webp" {
		text, err := readImageFile(path)
		if err != nil {
			return "There was an error reading the file" + path, err
		}
		return text, nil
	}

	if extension == "pdf" || extension == "docx" || extension == "epub" || extension == "pptx" || extension == "xlsx" || extension == "xls" {
		text, err := readFromFitz(path)
		if err != nil {
			return "There was an error reading the file" + path, err
		}
		return text, nil
	}

	if extension == "mp3" || extension == "ogg" || extension == "mp4" || extension == "flac" || extension == "m4a" || extension == "dsf" || extension == "wav" {
		text, err := readMetadata(path)
		if err != nil {
			return "There was an error reading the file" + path, err
		}
		return text, nil
	}

	// We can't check every file type, so try to read the file as a string using the os package
	content, err := os.ReadFile(path)
	if err != nil {
		return "There was an error reading the file" + path, err
	}
	return string(content), nil
}

func readRawFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func readImageFile(path string) (string, error) {
	// Verify file exists and is readable
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("error accessing image file: %v", err)
	}

	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetImage(path); err != nil {
		return "", fmt.Errorf("error setting image in Tesseract: %v", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("error performing OCR: %v", err)
	}

	if text == "" {
		return "No text extracted from image file: " + path, nil
	}

	return text, nil
}

// convert pdf to image and then read the image using the readImageFile function
func readFromFitz(path string) (string, error) {
	doc, err := fitz.New(path)
	if err != nil {
		return "There was an error reading the file: creating fitz document" + path, err
	}
	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "nomnom")
	if err != nil {
		panic(err)
	}

	var text string

	// Extract pages as images
	// we only need the first 2 pages as they should give us the most context for the name
	for n := range 2 {
		img, err := doc.Image(n)
		if err != nil {
			return "There was an error reading the file: extracting image" + path, err
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("nomnom-%03d.jpg", n)))
		if err != nil {
			return "There was an error reading the file: creating image" + path, err
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			return "There was an error reading the file: encoding image" + path, err
		}

		f.Close()

		// give the full path to the image file
		text, err = readImageFile(filepath.Join(tmpDir, fmt.Sprintf("nomnom-%03d.jpg", n)))

		if err != nil {
			return "There was an error reading the file" + path, err
		}
		text += text

	}
	// delete the tmpDir
	os.RemoveAll(tmpDir)

	return text, nil
}

func readMetadata(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return "", err
	}

	var metadata []string

	// Basic metadata
	if title := m.Title(); title != "" {
		metadata = append(metadata, fmt.Sprintf("Title: %s", title))
	}
	if album := m.Album(); album != "" {
		metadata = append(metadata, fmt.Sprintf("Album: %s", album))
	}
	if artist := m.Artist(); artist != "" {
		metadata = append(metadata, fmt.Sprintf("Artist: %s", artist))
	}
	if albumArtist := m.AlbumArtist(); albumArtist != "" {
		metadata = append(metadata, fmt.Sprintf("Album Artist: %s", albumArtist))
	}
	if composer := m.Composer(); composer != "" {
		metadata = append(metadata, fmt.Sprintf("Composer: %s", composer))
	}
	if genre := m.Genre(); genre != "" {
		metadata = append(metadata, fmt.Sprintf("Genre: %s", genre))
	}
	if year := m.Year(); year != 0 {
		metadata = append(metadata, fmt.Sprintf("Year: %d", year))
	}

	// Track and disc information
	trackNum, trackTotal := m.Track()
	if trackNum != 0 {
		if trackTotal != 0 {
			metadata = append(metadata, fmt.Sprintf("Track: %d/%d", trackNum, trackTotal))
		} else {
			metadata = append(metadata, fmt.Sprintf("Track: %d", trackNum))
		}
	}

	discNum, discTotal := m.Disc()
	if discNum != 0 {
		if discTotal != 0 {
			metadata = append(metadata, fmt.Sprintf("Disc: %d/%d", discNum, discTotal))
		} else {
			metadata = append(metadata, fmt.Sprintf("Disc: %d", discNum))
		}
	}

	// Additional metadata
	if lyrics := m.Lyrics(); lyrics != "" {
		metadata = append(metadata, fmt.Sprintf("Lyrics: %s", lyrics))
	}
	if comment := m.Comment(); comment != "" {
		metadata = append(metadata, fmt.Sprintf("Comment: %s", comment))
	}

	// Format information
	if format := m.Format(); format != "" {
		metadata = append(metadata, fmt.Sprintf("Format: %s", format))
	}
	if fileType := m.FileType(); fileType != "" {
		metadata = append(metadata, fmt.Sprintf("File Type: %s", fileType))
	}

	// Picture/artwork presence indicator
	if picture := m.Picture(); picture != nil {
		metadata = append(metadata, "Artwork: Present")
	}

	if len(metadata) == 0 {
		return "", nil
	}

	text := strings.Join(metadata, "\n")

	//if text is empty or only contains one line, just return no context and print a log
	if text == "" || strings.Count(text, "\n") <= 1 {
		log.Printf("[2/6] No metadata found for file: %s", path)
		return "[2/6] No metadata found for file: " + text + path, nil
	}

	return text, nil
}
