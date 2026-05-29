package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/dhowden/tag"
)

type ExtractedContent struct {
	Text             string
	PreviewImagePath string
	VisualContent    string
}

type ExtractOptions struct {
	MaxTextBytes    int64
	GeneratePreview bool
}

const defaultExtractTextLimit int64 = 32 * 1024

var previewTempDir struct {
	once sync.Once
	dir  string
	err  error
}

func ReadFile(path string) (string, error) {
	content, err := ExtractFileContent(path)
	if err != nil {
		return "", err
	}
	return content.Text, nil
}

func ExtractFileContent(path string) (ExtractedContent, error) {
	return ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    defaultExtractTextLimit,
		GeneratePreview: true,
	})
}

func ExtractFileContentWithOptions(path string, opts ExtractOptions) (ExtractedContent, error) {
	if opts.MaxTextBytes <= 0 {
		opts.MaxTextBytes = defaultExtractTextLimit
	}

	extension := strings.TrimPrefix(GetFileExtension(path), ".")

	switch extension {
	case "txt", "md", "json":
		content, err := readRawFile(path, opts.MaxTextBytes)
		if err != nil {
			return ExtractedContent{}, fmt.Errorf("there was an error reading the file %s: %w", path, err)
		}
		return ExtractedContent{Text: content}, nil
	case "png", "jpg", "jpeg", "webp":
		text, err := readImageFile(path)
		if err != nil {
			return ExtractedContent{}, fmt.Errorf("there was an error reading the file %s: %w", path, err)
		}
		return ExtractedContent{Text: text, PreviewImagePath: path}, nil
	case "pdf", "docx", "epub", "pptx", "xlsx", "xls":
		return readDocumentContent(path, opts)
	case "mp3", "ogg", "mp4", "flac", "m4a", "dsf", "wav":
		text, err := readMetadata(path)
		if err != nil {
			return ExtractedContent{}, fmt.Errorf("there was an error reading the file %s: %w", path, err)
		}
		return ExtractedContent{Text: text}, nil
	default:
		content, err := readRawFile(path, opts.MaxTextBytes)
		if err != nil {
			return ExtractedContent{}, fmt.Errorf("there was an error reading the file %s: %w", path, err)
		}
		return ExtractedContent{Text: content}, nil
	}
}

func readRawFile(path string, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func readImageFile(_ string) (string, error) {
	return "An image preview is available for this file. Use the visual contents to infer a better filename.", nil
}

func CleanupPreviewTempDir() error {
	if previewTempDir.dir == "" {
		return nil
	}
	return os.RemoveAll(previewTempDir.dir)
}

func fallbackDocumentContent(path string, cause error) (ExtractedContent, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ExtractedContent{}, fmt.Errorf("there was an error reading the file %s: %w", path, cause)
	}

	return ExtractedContent{
		Text: strings.Join([]string{
			"Document extraction fallback.",
			fmt.Sprintf("File: %s", filepath.Base(path)),
			fmt.Sprintf("Extension: %s", strings.ToLower(filepath.Ext(path))),
			fmt.Sprintf("Size: %s bytes", strconv.FormatInt(info.Size(), 10)),
			fmt.Sprintf("Parser error: %v", cause),
			"Use the filename and any available visual preview to infer a better name.",
		}, "\n"),
	}, nil
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

	if lyrics := m.Lyrics(); lyrics != "" {
		metadata = append(metadata, fmt.Sprintf("Lyrics: %s", lyrics))
	}
	if comment := m.Comment(); comment != "" {
		metadata = append(metadata, fmt.Sprintf("Comment: %s", comment))
	}
	if format := m.Format(); format != "" {
		metadata = append(metadata, fmt.Sprintf("Format: %s", format))
	}
	if fileType := m.FileType(); fileType != "" {
		metadata = append(metadata, fmt.Sprintf("File Type: %s", fileType))
	}
	if picture := m.Picture(); picture != nil {
		metadata = append(metadata, "Artwork: Present")
	}

	if len(metadata) == 0 {
		return "", nil
	}

	text := strings.Join(metadata, "\n")
	if text == "" || strings.Count(text, "\n") <= 1 {
		return "Sparse metadata found for file: " + filepath.Base(path) + "\n" + text, nil
	}

	return text, nil
}
