package cntxt

import (
	"encoding/json"
	"errors"
	"fmt"
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
	fmt.Printf("current context\n%+v\n----\n", currentContext)
	fmt.Printf("c\n%+v\n----\n", c)
	if err != nil {
		return "", err
	}
	created := time.Now().UTC().Format(SkDateFormat)
	notes := []string{}
	if !utils.IsNullJSON(c.Notes) {
		// c.Notes = []byte(`[]`)
		err = json.Unmarshal([]byte(c.Notes), &notes)
		if err != nil {
			return "", err
		}
	}
	fmt.Printf("notes\n%+v\n----\n", notes)
	if currentContext.Name == c.Name && currentContext.ParentId == c.ParentId { // for now we will alwaus create a new context with a new note
		c = currentContext
		// only update notes
		if string(currentContext.Notes) != "" && len(notes) != 0 {
			fmt.Printf("currentContext.Notes\n%+v\n----\n", string(currentContext.Notes))
			fmt.Print("both notes are not empty\n----\n")
			currentNotes := []string{}
			err := json.Unmarshal(currentContext.Notes, &currentNotes)
			if err != nil {
				fmt.Printf("error unmarshalling current notes\n%s\n----\n", err.Error())
				return "", err
			}

			notes = append(currentNotes, notes...)
			notesJson, err := utils.JsonMarshal(notes, false)
			// notesJson, err := json.Marshal(notes)
			if err != nil {
				return "", err
			}
			c.NoteString = string(notesJson)
			fmt.Printf("notes\n%+v\n----\n", notes)
		} else if string(currentContext.Notes) != "" && len(notes) == 0 {
			c.NoteString = currentContext.NoteString
		}
	} else {
		// create new context
		c.Created = created
		c.ContextId = created
		if currentContext.ContextId != "" {
			c.LastContext = currentContext.ContextId
			// if currentContext.Sk != c.ParentId {
			// 	currentContext.Close()
			// }
			currentContext.Close()
		}
		if currentContext.UserId != "" {
			SetLastContext(currentContext.UserId, currentContext.ContextId)
		}
		SetCurrentContext(c.UserId, c.ContextId)
	}
	fmt.Printf("update context\n%+v\n----\n", c)
	noteBytes, err := utils.JsonMarshal(notes, false)
	// noteBytes, err := json.Marshal(notes)
	if err != nil {
		return "", err
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
	fmt.Printf("before close\n%+v\n----\n", beforeClose)
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
	fmt.Printf("response\n%+v\n----\n", currentResponse)
	if err != nil {
		fmt.Printf("error getting current context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	currentJSON, err := utils.JsonMarshal(currentResponse, false)
	// currentJSON, err := json.Marshal(currentResponse)
	if err != nil {
		fmt.Printf("error marshalling current context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	currentContextId := struct {
		CurrentContext string `json:"contextLookup"`
	}{}
	err = json.Unmarshal(currentJSON, &currentContextId)
	if err != nil {
		fmt.Printf("error unmarshalling current context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	fmt.Printf("currentContextId\n%+v\n----\n", currentContextId)

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
	fmt.Printf("response\n%+v\n----\n", lastResponse)
	if err != nil {
		fmt.Printf("error getting last context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	lastJSON, err := utils.JsonMarshal(lastResponse, false)
	// lastJSON, err := json.Marshal(lastResponse)
	if err != nil {
		fmt.Printf("error marshalling last context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	lastContextId := struct {
		LastContext string `json:"contextLookup"`
	}{}
	err = json.Unmarshal(lastJSON, &lastContextId)
	if err != nil {
		fmt.Printf("error unmarshalling last context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	fmt.Printf("lastContextId\n%+v\n----\n", lastContextId)

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
		fmt.Printf("error getting context\n%s\n----\n", err.Error())
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
		fmt.Printf("error getting context\n%s\n----\n", err.Error())
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
	// ctxJSON, err := json.Marshal(c)
	if err != nil {
		fmt.Printf("error marshalling context\n%s\n----\n", err.Error())
		return "", err
	}

	return string(ctxJSON), nil
}

func responseToContext(contextResponse map[string]interface{}) (*Context, error) {
	fmt.Printf("contextResponse\n%+v\n----\n", contextResponse)
	noteIntf := contextResponse["notesString"]
	noteString, ok := noteIntf.(string)
	ctxJSON, err := utils.JsonMarshal(contextResponse, false)
	// ctxJSON, err := json.Marshal(contextResponse)
	if err != nil {
		fmt.Printf("error marshalling context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	currentContext := &Context{}
	err = json.Unmarshal(ctxJSON, currentContext)
	if err != nil {
		fmt.Printf("error unmarshalling context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	fmt.Printf("currentContext\n%+v\n----\n", currentContext)
	if currentContext.UserId == "" {
		return &Context{}, errors.New("not found")
	}
	// noteBytes, err := json.Marshal(&currentContext.NoteString)
	// if err != nil {
	// 	fmt.Printf("error marshalling notes\n%s\n----\n", err.Error())
	// 	return &Context{}, err
	// }
	// byteString := []byte(`{"notes": %s}`)
	if ok {
		noteBytes := json.RawMessage{}
		fmt.Printf("noteString\n%+v\n----\n", noteString)
		err = json.Unmarshal([]byte(noteString), &noteBytes)
		if err != nil {
			fmt.Printf("error unmarshalling notes\n%s\n----\n", err.Error())
			return &Context{}, err
		}
		fmt.Printf("noteBytes\n%+v\n----\n", noteBytes)
		fmt.Printf("string(noteBytes)\n%+v\n----\n", string(noteBytes))
		currentContext.Notes = noteBytes
		currentContext.NoteString = noteString
		if noteString == "[]" {
			currentContext.Notes = json.RawMessage{}
		}
	}
	// fmt.Printf("noteString\n%+v\n----\n", currentContext.NoteString)
	// s := fmt.Sprintf(`%s`, currentContext.NoteString)
	// fmt.Printf("s\n%+v\n----\n", s)
	// noteBytes := json.RawMessage{}
	// err = json.Unmarshal([]byte(s), &noteBytes)
	// if err != nil {
	// 	fmt.Printf("error unmarshalling notes\n%s\n----\n", err.Error())
	// 	return &Context{}, err
	// }
	// currentContext.Notes = noteBytes
	// fmt.Printf("[]byte(currentContext.NoteString\n%+v\n----\n", fmt.Sprintf(`%s`, currentContext.NoteString))
	// currentContext.Notes = currentContext.NoteString
	fmt.Printf("currentContext at end of responseToContext\n%+v\n----\n", currentContext)
	return currentContext, nil
}

func saveContext(c *Context) error {
	fmt.Printf("saveContext\n%+v\n----\n", c)
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
	fmt.Printf("saveContext payload\n%+v\n----\n", payload)

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
