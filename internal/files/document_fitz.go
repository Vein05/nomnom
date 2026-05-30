//go:build !windows

package files

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"os"
	"strings"

	"github.com/gen2brain/go-fitz"
)

func readDocumentContent(path string, opts ExtractOptions) (ExtractedContent, error) {
	doc, err := fitz.New(path)
	if err != nil {
		return fallbackDocumentContent(path, fmt.Errorf("creating fitz document: %w", err))
	}
	defer doc.Close()

	text, textErr := extractDocumentText(doc, path)

	var previewPath string
	var visualContent string
	var previewErr error
	if opts.GeneratePreview {
		// Encode the preview in-memory first so AI providers can consume it directly.
		previewBytes, err := RenderFirstPageToBytes(doc)
		if err == nil {
			visualContent = base64.StdEncoding.EncodeToString(previewBytes)
			previewPath, previewErr = renderFirstPagePreview(doc, path)
		} else {
			previewErr = err
		}
	}

	if textErr != nil && (!opts.GeneratePreview || previewErr != nil) {
		return fallbackDocumentContent(path, fmt.Errorf("extracting document content failed: %v; preview failed: %v", textErr, previewErr))
	}

	if len(text) > int(opts.MaxTextBytes) {
		text = text[:opts.MaxTextBytes]
	}

	if text == "" {
		text = "Minimal document text was extracted. Prefer the first-page preview if available."
	}

	if opts.GeneratePreview && previewErr == nil {
		text += "\nA first-page preview image is available for this document."
	}

	return ExtractedContent{
		Text:             text,
		PreviewImagePath: previewPath,
		VisualContent:    visualContent,
	}, nil
}

func extractDocumentText(doc *fitz.Document, path string) (string, error) {
	pageCount := doc.NumPage()
	if pageCount == 0 {
		return "", nil
	}

	limit := min(pageCount, 2)
	pages := make([]string, 0, limit)
	for page := 0; page < limit; page++ {
		text, err := doc.Text(page)
		if err != nil {
			return "", fmt.Errorf("extracting text from %s page %d: %w", path, page+1, err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		pages = append(pages, text)
	}

	return strings.Join(pages, "\n\n"), nil
}

func renderFirstPagePreview(doc *fitz.Document, sourcePath string) (string, error) {
	if doc.NumPage() == 0 {
		return "", fmt.Errorf("document has no pages")
	}

	img, err := doc.Image(0)
	if err != nil {
		return "", fmt.Errorf("rendering first page image: %w", err)
	}

	dir, err := previewSessionDir()
	if err != nil {
		return "", fmt.Errorf("creating preview session temp dir for %s: %w", sourcePath, err)
	}

	tmpFile, err := os.CreateTemp(dir, "nomnom-preview-*.jpg")
	if err != nil {
		return "", fmt.Errorf("creating temp preview for %s: %w", sourcePath, err)
	}
	defer tmpFile.Close()

	if err := jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("encoding preview image for %s: %w", sourcePath, err)
	}

	return tmpFile.Name(), nil
}

func RenderFirstPageToBytes(doc *fitz.Document) ([]byte, error) {
	if doc.NumPage() == 0 {
		return nil, fmt.Errorf("document has no pages")
	}

	img, err := doc.Image(0)
	if err != nil {
		return nil, fmt.Errorf("rendering first page image: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
		return nil, fmt.Errorf("encoding preview image to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

func previewSessionDir() (string, error) {
	previewTempDir.once.Do(func() {
		previewTempDir.dir, previewTempDir.err = os.MkdirTemp("", "nomnom-preview-session-*")
	})

	if previewTempDir.err != nil {
		return "", previewTempDir.err
	}

	return previewTempDir.dir, nil
}
