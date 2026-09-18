package cli

import (
	"bytes"
	"encoding/json"
	"io"
)

func printJSON(w io.Writer, raw []byte) error {
	var buffer bytes.Buffer
	if err := json.Indent(&buffer, raw, "", "  "); err != nil {
		return err
	}
	buffer.WriteByte('\n')

	_, err := w.Write(buffer.Bytes())

	return err
}

type attachmentResult struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Size     int64  `json:"size,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Error    string `json:"error,omitempty"`
	Success  bool   `json:"success"`
}

type attachmentReport struct {
	IssueKey         string             `json:"issueKey"`
	OutputDirectory  string             `json:"outputDirectory"`
	TotalAttachments int                `json:"totalAttachments"`
	Results          []attachmentResult `json:"results"`
}

func printReport(w io.Writer, report attachmentReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	return encoder.Encode(report)
}
