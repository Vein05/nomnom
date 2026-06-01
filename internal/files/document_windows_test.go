//go:build windows

package files

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractPDFText(t *testing.T) {
	// ledongthuc/pdf requires a structurally valid PDF with correct xref offsets.
	// We test that extractPDFText returns an error for invalid input and that
	// readDocumentContent gracefully falls back.
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.0\ninvalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := extractPDFText(path)
	if err == nil {
		t.Fatal("expected error for malformed PDF, got nil")
	}

	// readDocumentContent should fall back gracefully
	content, err := readDocumentContent(path, ExtractOptions{MaxTextBytes: 1024})
	if err != nil {
		t.Fatalf("readDocumentContent() should not error on fallback, got: %v", err)
	}
	if !strings.Contains(content.Text, "Document extraction fallback") {
		t.Errorf("expected fallback text, got: %q", content.Text)
	}
}

func TestExtractDOCXText(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.docx")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)

	docXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>NomNom Document Test</w:t></w:r></w:p>
    <w:p><w:r><w:t>Second paragraph content</w:t></w:r></w:p>
  </w:body>
</w:document>`

	fw, _ := w.Create("word/document.xml")
	fw.Write([]byte(docXML))
	w.Close()
	f.Close()

	text, err := extractDOCXText(path)
	if err != nil {
		t.Fatalf("extractDOCXText() error = %v", err)
	}
	if !strings.Contains(text, "NomNom Document Test") {
		t.Errorf("expected 'NomNom Document Test', got: %q", text)
	}
	if !strings.Contains(text, "Second paragraph content") {
		t.Errorf("expected 'Second paragraph content', got: %q", text)
	}
}

func TestExtractXLSXText(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.xlsx")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)

	// Shared strings
	ssXML := `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2">
  <si><t>Revenue</t></si>
  <si><t>Expenses</t></si>
</sst>`
	fw, _ := w.Create("xl/sharedStrings.xml")
	fw.Write([]byte(ssXML))

	// Sheet
	sheetXML := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row>
    <row r="2"><c r="A2"><v>5000</v></c><c r="B2"><v>3000</v></c></row>
  </sheetData>
</worksheet>`
	fw, _ = w.Create("xl/worksheets/sheet1.xml")
	fw.Write([]byte(sheetXML))

	w.Close()
	f.Close()

	text, err := extractXLSXText(path)
	if err != nil {
		t.Fatalf("extractXLSXText() error = %v", err)
	}
	if !strings.Contains(text, "Revenue") {
		t.Errorf("expected 'Revenue', got: %q", text)
	}
	if !strings.Contains(text, "Expenses") {
		t.Errorf("expected 'Expenses', got: %q", text)
	}
	if !strings.Contains(text, "5000") {
		t.Errorf("expected '5000', got: %q", text)
	}
}

func TestExtractPPTXText(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.pptx")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)

	slideXML := `<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
       xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp><p:txBody><a:p><a:r><a:t>Slide Title Here</a:t></a:r></a:p></p:txBody></p:sp>
      <p:sp><p:txBody><a:p><a:r><a:t>Bullet point content</a:t></a:r></a:p></p:txBody></p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`
	fw, _ := w.Create("ppt/slides/slide1.xml")
	fw.Write([]byte(slideXML))
	w.Close()
	f.Close()

	text, err := extractPPTXText(path)
	if err != nil {
		t.Fatalf("extractPPTXText() error = %v", err)
	}
	if !strings.Contains(text, "Slide Title Here") {
		t.Errorf("expected 'Slide Title Here', got: %q", text)
	}
	if !strings.Contains(text, "Bullet point content") {
		t.Errorf("expected 'Bullet point content', got: %q", text)
	}
}

func TestExtractEPUBText(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.epub")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)

	htmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body>
  <h1>The Beginning</h1>
  <p>Once upon a time in a land far away.</p>
</body>
</html>`
	fw, _ := w.Create("OEBPS/chapter1.xhtml")
	fw.Write([]byte(htmlContent))
	w.Close()
	f.Close()

	text, err := extractEPUBText(path)
	if err != nil {
		t.Fatalf("extractEPUBText() error = %v", err)
	}
	if !strings.Contains(text, "The Beginning") {
		t.Errorf("expected 'The Beginning', got: %q", text)
	}
	if !strings.Contains(text, "Once upon a time") {
		t.Errorf("expected 'Once upon a time', got: %q", text)
	}
}

func TestReadDocumentContentWindows(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test DOCX
	path := filepath.Join(tmpDir, "test.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, _ := w.Create("word/document.xml")
	fw.Write([]byte(xml.Header + `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Integration test</w:t></w:r></w:p></w:body></w:document>`))
	w.Close()
	f.Close()

	content, err := readDocumentContent(path, ExtractOptions{MaxTextBytes: 1024, GeneratePreview: false})
	if err != nil {
		t.Fatalf("readDocumentContent() error = %v", err)
	}
	if !strings.Contains(content.Text, "Integration test") {
		t.Errorf("expected 'Integration test', got: %q", content.Text)
	}
}

func TestReadDocumentContentMaxBytes(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.docx")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, _ := w.Create("word/document.xml")
	longText := strings.Repeat("NomNom ", 100)
	fw.Write([]byte(xml.Header + `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>` + longText + `</w:t></w:r></w:p></w:body></w:document>`))
	w.Close()
	f.Close()

	content, err := readDocumentContent(path, ExtractOptions{MaxTextBytes: 20, GeneratePreview: false})
	if err != nil {
		t.Fatalf("readDocumentContent() error = %v", err)
	}
	if len(content.Text) != 20 {
		t.Errorf("expected text length 20, got %d", len(content.Text))
	}
}
