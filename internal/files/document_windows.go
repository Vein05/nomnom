//go:build windows

package files

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

func readDocumentContent(path string, opts ExtractOptions) (ExtractedContent, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

	var text string
	var err error

	switch ext {
	case "pdf":
		text, err = extractPDFText(path)
	case "docx":
		text, err = extractDOCXText(path)
	case "xlsx", "xls":
		text, err = extractXLSXText(path)
	case "pptx":
		text, err = extractPPTXText(path)
	case "epub":
		text, err = extractEPUBText(path)
	default:
		return fallbackDocumentContent(path, fmt.Errorf("unsupported document type: %s", ext))
	}

	if err != nil {
		return fallbackDocumentContent(path, err)
	}

	if int64(len(text)) > opts.MaxTextBytes && opts.MaxTextBytes > 0 {
		text = text[:opts.MaxTextBytes]
	}

	if strings.TrimSpace(text) == "" {
		return fallbackDocumentContent(path, fmt.Errorf("no text extracted from %s", ext))
	}

	return ExtractedContent{Text: text}, nil
}

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	var buf strings.Builder
	limit := 3
	if r.NumPage() < limit {
		limit = r.NumPage()
	}
	for i := 1; i <= limit; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		if i < limit {
			buf.WriteString("\n\n")
		}
	}
	return buf.String(), nil
}

func extractDOCXText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("open document.xml: %w", err)
			}
			defer rc.Close()
			return extractXMLText(rc), nil
		}
	}
	return "", fmt.Errorf("word/document.xml not found in docx")
}

func extractXLSXText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open xlsx: %w", err)
	}
	defer r.Close()

	sharedStrings := parseSharedStrings(r)

	var buf strings.Builder
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, "xl/worksheets/sheet") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text := extractSheetText(rc, sharedStrings)
		rc.Close()
		if text != "" {
			buf.WriteString(text)
			buf.WriteString("\n")
		}
	}
	return buf.String(), nil
}

func parseSharedStrings(r *zip.ReadCloser) []string {
	for _, f := range r.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil
			}
			defer rc.Close()

			var result []string
			decoder := xml.NewDecoder(rc)
			var inSI, inT bool
			var current strings.Builder
			for {
				tok, err := decoder.Token()
				if err != nil {
					break
				}
				switch t := tok.(type) {
				case xml.StartElement:
					if t.Name.Local == "si" {
						inSI = true
						current.Reset()
					} else if inSI && t.Name.Local == "t" {
						inT = true
					}
				case xml.CharData:
					if inT {
						current.Write(t)
					}
				case xml.EndElement:
					if t.Name.Local == "t" {
						inT = false
					} else if t.Name.Local == "si" {
						inSI = false
						result = append(result, current.String())
					}
				}
			}
			return result
		}
	}
	return nil
}

func extractSheetText(rc io.ReadCloser, sharedStrings []string) string {
	var buf strings.Builder
	decoder := xml.NewDecoder(rc)
	var inV, inC bool
	var cellType string
	var cellVal strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "c" {
				inC = true
				cellType = ""
				for _, attr := range t.Attr {
					if attr.Name.Local == "t" {
						cellType = attr.Value
					}
				}
			} else if inC && t.Name.Local == "v" {
				inV = true
				cellVal.Reset()
			}
		case xml.CharData:
			if inV {
				cellVal.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "v" {
				inV = false
				val := cellVal.String()
				if cellType == "s" && sharedStrings != nil {
					idx := 0
					fmt.Sscanf(val, "%d", &idx)
					if idx < len(sharedStrings) {
						val = sharedStrings[idx]
					}
				}
				if val != "" {
					buf.WriteString(val)
					buf.WriteString("\t")
				}
			} else if t.Name.Local == "row" {
				buf.WriteString("\n")
			} else if t.Name.Local == "c" {
				inC = false
			}
		}
	}
	return buf.String()
}

func extractPPTXText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open pptx: %w", err)
	}
	defer r.Close()

	var buf strings.Builder
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, "ppt/slides/slide") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text := extractXMLText(rc)
		rc.Close()
		if text != "" {
			buf.WriteString(text)
			buf.WriteString("\n")
		}
	}
	return buf.String(), nil
}

func extractEPUBText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open epub: %w", err)
	}
	defer r.Close()

	var buf strings.Builder
	for _, f := range r.File {
		lower := strings.ToLower(f.Name)
		if !strings.HasSuffix(lower, ".xhtml") && !strings.HasSuffix(lower, ".html") && !strings.HasSuffix(lower, ".htm") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text := extractXMLText(rc)
		rc.Close()
		if text != "" {
			buf.WriteString(text)
			buf.WriteString("\n")
		}
	}
	return buf.String(), nil
}

// extractXMLText walks XML tokens and collects all character data.
func extractXMLText(r io.Reader) string {
	var buf strings.Builder
	decoder := xml.NewDecoder(r)
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if cd, ok := tok.(xml.CharData); ok {
			text := strings.TrimSpace(string(cd))
			if text != "" {
				buf.WriteString(text)
				buf.WriteString(" ")
			}
		}
	}
	return strings.TrimSpace(buf.String())
}
