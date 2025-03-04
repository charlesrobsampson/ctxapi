package types

import "encoding/json"

type (
	Context struct {
		Name        string          `json:"name,omitempty"`
		Notes       json.RawMessage `json:"notes,omitempty"`
		UserId      string          `json:"userId,omitempty"`
		ContextId   string          `json:"contextId,omitempty"`
		ParentId    string          `json:"parentId,omitempty"`
		LastContext string          `json:"lastContext,omitempty"`
		NoteString  string          `json:"noteString,omitempty"`
		Document    Document        `json:"document,omitempty"`
		Created     string          `json:"created,omitempty"`
		Completed   string          `json:"completed,omitempty"`
	}

	Document struct {
		RealtivePath string `json:"realtivePath,omitempty"`
		Github       string `json:"github,omitempty"`
	}
)
