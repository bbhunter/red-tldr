package render

import (
	"testing"
)

func TestNew_DefaultFormat(t *testing.T) {
	r := New("", true)
	if r.format != FormatText {
		t.Errorf("expected FormatText, got %s", r.format)
	}
}

func TestNew_JSONFormat(t *testing.T) {
	r := New("json", true)
	if r.format != FormatJSON {
		t.Errorf("expected FormatJSON, got %s", r.format)
	}
}

func TestNew_MarkdownFormat(t *testing.T) {
	r := New("markdown", false)
	if r.format != FormatMarkdown {
		t.Errorf("expected FormatMarkdown, got %s", r.format)
	}
}

func TestNew_MdShorthand(t *testing.T) {
	r := New("md", false)
	if r.format != FormatMarkdown {
		t.Errorf("expected FormatMarkdown for 'md', got %s", r.format)
	}
}

func TestNew_UnknownFormat(t *testing.T) {
	r := New("xml", true)
	if r.format != FormatText {
		t.Errorf("expected FormatText for unknown format, got %s", r.format)
	}
}
