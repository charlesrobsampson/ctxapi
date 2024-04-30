package cntxt

import (
	"encoding/json"
	"errors"
	"main/utils"
	"time"
)

type (
	Context struct {
		Name        string          `json:"name,omitempty"`
		Notes       json.RawMessage `json:"notes,omitempty"`
		UserId      string          `json:"userId,omitempty"`
		ContextId   string          `json:"contextId,omitempty"`
		ParentId    string          `json:"parentId,omitempty"`
		LastContext string          `json:"lastContext,omitempty"`
		NoteString  string          `json:"noteString,omitempty"`
		Created     string          `json:"created,omitempty"`
		Completed   string          `json:"completed,omitempty"`
	}
)

var (
	SkDateFormat    = utils.SkDateFormat
	pkString        = "userId"
	timestampString = "contextId"
)

func (c *Context) Update() (string, error) {
	currentContext, err := GetCurrentContext(c.UserId)
	lastContext := Context{}
	if err != nil {
		return "", err
	}
	if currentContext.Name == "" {
		last, err := GetLastContext(c.UserId)
		lastContext = *last
		if err != nil {
			return "", err
		}
	}
	created := time.Now().UTC().Format(SkDateFormat)
	notes := []string{}
	if !utils.IsNullJSON(c.Notes) {
		err = json.Unmarshal([]byte(c.Notes), &notes)
		if err != nil {
			return "", err
		}
	}
	if currentContext.Name == c.Name && currentContext.ParentId == c.ParentId {
		c = currentContext
		// only update notes
		if string(currentContext.Notes) != "" && len(notes) != 0 {
			currentNotes := []string{}
			err := json.Unmarshal(currentContext.Notes, &currentNotes)
			if err != nil {
				return "", err
			}

			notes = append(currentNotes, notes...)
			notesJson, err := utils.JsonMarshal(notes, false)
			if err != nil {
				return "", err
			}
			c.NoteString = string(notesJson)
		} else if string(currentContext.Notes) != "" && len(notes) == 0 {
			c.NoteString = currentContext.NoteString
		}
	} else {
		// create new context
		c.Created = created
		c.ContextId = created
		if currentContext.ContextId != "" {
			c.LastContext = currentContext.ContextId
			currentContext.Close()
		}
		if currentContext.UserId != "" {
			SetLastContext(currentContext.UserId, currentContext.ContextId)
		}
		SetCurrentContext(c.UserId, c.ContextId)
	}
	noteBytes, err := utils.JsonMarshal(notes, false)
	if err != nil {
		return "", err
	}
	if lastContext.Name != "" {
		c.LastContext = lastContext.ContextId
	}
	c.NoteString = string(noteBytes)
	saveContext(c)
	return c.ContextId, nil
}

func (c *Context) Close() error {
	beforeClose, err := GetContext(c.UserId, c.ContextId)
	if err != nil {
		return err
	}
	beforeClose.Completed = time.Now().UTC().Format(SkDateFormat)
	return saveContext(beforeClose)
}

func GetCurrentContext(userId string) (*Context, error) {
	currentResponse, err := utils.GetDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: "current",
		},
		TableName: utils.MainTableName,
	})
	if err != nil {
		return &Context{}, err
	}
	currentJSON, err := utils.JsonMarshal(currentResponse, false)
	if err != nil {
		return &Context{}, err
	}
	currentContextId := struct {
		CurrentContext string `json:"contextLookup"`
	}{}
	err = json.Unmarshal(currentJSON, &currentContextId)
	if err != nil {
		return &Context{}, err
	}

	if currentContextId.CurrentContext == "" {
		return &Context{}, nil
	}
	return GetContext(userId, currentContextId.CurrentContext)
}

func GetLastContext(userId string) (*Context, error) {
	lastResponse, err := utils.GetDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: "last",
		},
		TableName: utils.MainTableName,
	})
	if err != nil {
		return &Context{}, err
	}
	lastJSON, err := utils.JsonMarshal(lastResponse, false)
	if err != nil {
		return &Context{}, err
	}
	lastContextId := struct {
		LastContext string `json:"contextLookup"`
	}{}
	err = json.Unmarshal(lastJSON, &lastContextId)
	if err != nil {
		return &Context{}, err
	}

	if lastContextId.LastContext == "" {
		return &Context{}, nil
	}
	return GetContext(userId, lastContextId.LastContext)
}

func SetCurrentContext(userId, contextId string) error {
	err := utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: "current",
		},
		TableName: utils.MainTableName,
	}, map[string]interface{}{
		"contextLookup": contextId,
	})
	return err
}

func SetLastContext(userId, contextId string) error {
	err := utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: "last",
		},
		TableName: utils.MainTableName,
	}, map[string]interface{}{
		"contextLookup": contextId,
	})
	return err
}

func GetContext(userId, contextId string) (*Context, error) {
	contextResponse, err := utils.GetDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: contextId,
		},
		TableName: utils.MainTableName,
	})
	if err != nil {
		return &Context{}, err
	}
	return responseToContext(contextResponse)
}

func ListContexts(userId, lower, upper, filter string) (*[]Context, error) {
	contextResponses, err := utils.QueryWithFilter(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name: timestampString,
		},
		SKBetween: utils.Between{
			Lower: utils.NameVal{
				Name:  timestampString,
				Value: lower,
			},
			Upper: utils.NameVal{
				Name:  timestampString,
				Value: upper,
			},
		},
		TableName: utils.MainTableName,
	}, filter)
	if err != nil {
		return &[]Context{}, err
	}
	var contexts []Context
	for _, contextResponse := range contextResponses {
		context, err := responseToContext(contextResponse)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, *context)
	}
	return &contexts, nil
}

func (c *Context) ToJSONString() (string, error) {
	ctxJSON, err := utils.JsonMarshal(c, false)
	if err != nil {
		return "", err
	}

	return string(ctxJSON), nil
}

func responseToContext(contextResponse map[string]interface{}) (*Context, error) {
	noteIntf := contextResponse["notesString"]
	noteString, ok := noteIntf.(string)
	ctxJSON, err := utils.JsonMarshal(contextResponse, false)
	if err != nil {
		return &Context{}, err
	}
	currentContext := &Context{}
	err = json.Unmarshal(ctxJSON, currentContext)
	if err != nil {
		return &Context{}, err
	}
	if currentContext.UserId == "" {
		return &Context{}, errors.New("not found")
	}
	if ok {
		noteBytes := json.RawMessage{}
		err = json.Unmarshal([]byte(noteString), &noteBytes)
		if err != nil {
			return &Context{}, err
		}
		currentContext.Notes = noteBytes
		currentContext.NoteString = noteString
		if noteString == "[]" {
			currentContext.Notes = json.RawMessage{}
		}
	}
	return currentContext, nil
}

func saveContext(c *Context) error {
	createdTime, err := time.Parse(SkDateFormat, c.Created)
	expires := createdTime.AddDate(1, 0, 0).Unix()
	if err != nil {
		return err
	}
	if c.Completed != "" {
		completedTime, err := time.Parse(SkDateFormat, c.Completed)
		if err != nil {
			return err
		}
		expires = completedTime.AddDate(1, 0, 0).Unix()
	}
	payload := map[string]interface{}{
		"userId":        c.UserId,
		timestampString: c.ContextId,
		"parentId":      c.ParentId,
		"lastContext":   c.LastContext,
		"name":          c.Name,
		"notesString":   c.NoteString,
		"created":       c.Created,
		"completed":     c.Completed,
		"expires":       expires,
	}

	err = utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: c.UserId,
		},
		SK: utils.NameVal{
			Name:  timestampString,
			Value: c.ContextId,
		},
		TableName: utils.MainTableName,
	}, payload)
	return err
}
