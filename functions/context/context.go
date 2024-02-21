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
		Pk          string          `json:"pk,omitempty"`
		Sk          string          `json:"sk,omitempty"`
		ParentId    string          `json:"parentId,omitempty"`
		LastContext string          `json:"lastContext,omitempty"`
		Name        string          `json:"name,omitempty"`
		Notes       json.RawMessage `json:"notes,omitempty"`
		NoteString  string          `json:"noteString,omitempty"`
		Created     string          `json:"created,omitempty"`
		Completed   string          `json:"completed,omitempty"`
		Random      string          `json:"random,omitempty"`
	}
)

var (
	SkDateFormat = "2006-01-02T15:04:05Z"
)

func (c *Context) Update() (string, error) {
	currentContext, err := GetCurrentContext(c.Pk)
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
	if false && currentContext.Name == c.Name && currentContext.ParentId == c.ParentId { // for now we will alwaus create a new context with a new note
		// only update notes
		fmt.Printf("only updating notes\n----\n")
		// if len(currentContext.Notes) != 0 && len(c.Notes) != 0 {
		// 	c.Notes = append(currentContext.Notes, c.Notes...)
		// if currentContext.Notes != "" && c.Notes != "" {

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
			notesJson, err := json.Marshal(notes)
			if err != nil {
				return "", err
			}
			c.NoteString = string(notesJson)
			// 	c.Notes = fmt.Sprintf("%s\n%s", currentContext.Notes, c.Notes)
			fmt.Printf("notes\n%+v\n----\n", notes)
		} else if string(currentContext.Notes) != "" && len(notes) == 0 {
			// } else if len(currentContext.Notes) != 0 && len(c.Notes) == 0 {
			c.NoteString = currentContext.NoteString
		}
		return currentContext.Sk, utils.UpdateField(utils.Config{
			PK: utils.NameVal{
				Name:  "pk",
				Value: currentContext.Pk,
			},
			SK: utils.NameVal{
				Name:  "sk",
				Value: currentContext.Sk,
			},
			TableName: utils.MainTableName,
		}, "noteString", c.NoteString)
	} else {
		// create new context
		c.Created = created
		c.Sk = fmt.Sprintf("context#%s", created)
		if currentContext.Sk != "" {
			c.LastContext = currentContext.Sk
			// if currentContext.Sk != c.ParentId {
			// 	currentContext.Close()
			// }
			currentContext.Close()
		}
		if currentContext.Pk != "" {
			SetLastContext(currentContext.Pk, currentContext.Sk)
		}
		SetCurrentContext(c.Pk, c.Sk)
	}
	fmt.Printf("update context\n%+v\n----\n", c)
	noteBytes, err := json.Marshal(notes)
	if err != nil {
		return "", err
	}
	c.NoteString = string(noteBytes)
	saveContext(c)
	return c.Sk, nil
}

func (c *Context) Close() error {
	beforeClose, err := GetContext(c.Pk, c.Sk)
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
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  "sk",
			Value: "context",
		},
		TableName: utils.MainTableName,
	})
	fmt.Printf("response\n%+v\n----\n", currentResponse)
	if err != nil {
		fmt.Printf("error getting current context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	currentJSON, err := json.Marshal(currentResponse)
	if err != nil {
		fmt.Printf("error marshalling current context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	currentContextId := struct {
		CurrentContext string `json:"currentContext"`
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
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  "sk",
			Value: "lastContext",
		},
		TableName: utils.MainTableName,
	})
	fmt.Printf("response\n%+v\n----\n", lastResponse)
	if err != nil {
		fmt.Printf("error getting last context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	lastJSON, err := json.Marshal(lastResponse)
	if err != nil {
		fmt.Printf("error marshalling last context\n%s\n----\n", err.Error())
		return &Context{}, err
	}
	lastContextId := struct {
		LastContext string `json:"lastContext"`
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
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  "sk",
			Value: "context",
		},
		TableName: utils.MainTableName,
	}, map[string]interface{}{
		"currentContext": contextId,
	})
	return err
}

func SetLastContext(userId, contextId string) error {
	err := utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  "sk",
			Value: "lastContext",
		},
		TableName: utils.MainTableName,
	}, map[string]interface{}{
		"lastContext": contextId,
	})
	return err
}

func GetContext(userId, contextId string) (*Context, error) {
	contextResponse, err := utils.GetDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  "sk",
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
	contextResponse, err := utils.QueryWithFilter(utils.Config{
		PK: utils.NameVal{
			Name:  "pk",
			Value: userId,
		},
		SK: utils.NameVal{
			Name: "sk",
		},
		SKBetween: utils.Between{
			Lower: utils.NameVal{
				Name:  "sk",
				Value: lower,
			},
			Upper: utils.NameVal{
				Name:  "sk",
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
	for _, contextResponse := range contextResponse {
		context, err := responseToContext(contextResponse)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, *context)
	}
	return &contexts, nil
}

func (c *Context) ToJSONString() (string, error) {
	ctxJSON, err := json.Marshal(c)
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
	ctxJSON, err := json.Marshal(contextResponse)
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
	if currentContext.Pk == "" {
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
		// add 1 year to c.Completed and update expires to be that date in UNIX
		completedTime, err := time.Parse(SkDateFormat, c.Completed)
		if err != nil {
			return err
		}
		expires = completedTime.AddDate(1, 0, 0).Unix()
	}
	payload := map[string]interface{}{
		"pk":          c.Pk,
		"sk":          c.Sk,
		"parentId":    c.ParentId,
		"lastContext": c.LastContext,
		"name":        c.Name,
		// "notes":       c.Notes,
		"notesString": c.NoteString,
		"created":     c.Created,
		"completed":   c.Completed,
		"random":      c.Random,
		"expires":     expires,
	}
	fmt.Printf("saveContext payload\n%+v\n----\n", payload)

	err = utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  "pk",
			Value: c.Pk,
		},
		SK: utils.NameVal{
			Name:  "sk",
			Value: c.Sk,
		},
		TableName: utils.MainTableName,
	}, payload)
	return err
}
