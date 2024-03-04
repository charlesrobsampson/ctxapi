package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"main/utils"
	"time"
)

type (
	Queue struct {
		Name       string          `json:"name,omitempty"`
		Notes      json.RawMessage `json:"notes,omitempty"`
		UserId     string          `json:"userId,omitempty"`
		Id         string          `json:"id,omitempty"`
		NoteString string          `json:"noteString,omitempty"`
		Started    string          `json:"started,omitempty"`
		ContextId  string          `json:"contextId,omitempty"`
	}
)

var (
	SkDateFormat = utils.SkDateFormat
	pkString     = "userId"
	qIdString    = "id"
)

func (q *Queue) Update() (string, error) {
	existingQueue := &Queue{}
	if q.Id != "" {
		eq, err := GetQueue(q.UserId, q.Id)
		existingQueue = eq
		fmt.Printf("existingQueue\n%+v\n----\n", existingQueue)
		fmt.Printf("q\n%+v\n----\n", q)
		if err != nil {
			return "", err
		}
	}
	created := time.Now().UTC().Format(SkDateFormat)
	notes := []string{}
	if !utils.IsNullJSON(q.Notes) {
		// c.Notes = []byte(`[]`)
		err := json.Unmarshal([]byte(q.Notes), &notes)
		if err != nil {
			return "", err
		}
	}
	fmt.Printf("notes\n%+v\n----\n", notes)
	if existingQueue.Id != "" && q.Name == existingQueue.Name {
		q = existingQueue
		// only update notes
		if string(existingQueue.Notes) != "" && len(notes) != 0 {
			fmt.Printf("currentQueue.Notes\n%+v\n----\n", string(existingQueue.Notes))
			fmt.Print("both notes are not empty\n----\n")
			currentNotes := []string{}
			err := json.Unmarshal(existingQueue.Notes, &currentNotes)
			if err != nil {
				fmt.Printf("error unmarshalling current notes\n%s\n----\n", err.Error())
				return "", err
			}

			notes = append(currentNotes, notes...)
			notesJson, err := json.Marshal(notes)
			if err != nil {
				return "", err
			}
			q.NoteString = string(notesJson)
			fmt.Printf("notes\n%+v\n----\n", notes)
		} else if string(existingQueue.Notes) != "" && len(notes) == 0 {
			q.NoteString = existingQueue.NoteString
		}
	} else {
		// create new context
		q.Id = created
	}
	fmt.Printf("update queue\n%+v\n----\n", q)
	noteBytes, err := json.Marshal(notes)
	if err != nil {
		return "", err
	}
	q.NoteString = string(noteBytes)
	saveQueue(q)
	return q.Id, nil
}

func (q *Queue) Start() (*Queue, error) {
	q.Started = time.Now().UTC().Format(SkDateFormat)

	err := saveQueue(q)

	return q, err
}

func (q *Queue) ToJSONString() (string, error) {
	qJSON, err := json.Marshal(q)
	if err != nil {
		fmt.Printf("error marshalling queue\n%s\n----\n", err.Error())
		return "", err
	}

	return string(qJSON), nil
}

func GetQueue(userId, queueId string) (*Queue, error) {
	fmt.Printf("getting queue with id %s for user %s\n", queueId, userId)
	contextResponse, err := utils.GetDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		SK: utils.NameVal{
			Name:  qIdString,
			Value: queueId,
		},
		TableName: utils.QueueTableName,
	})
	if err != nil {
		fmt.Printf("error getting queue\n%s\n----\n", err.Error())
		return &Queue{}, err
	}
	return responseToQueue(contextResponse)
}

func ListQueue(userId string, pendingOnly bool) (*[]Queue, error) {
	queueResponse, err := utils.QueryWithFilter(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: userId,
		},
		TableName: utils.QueueTableName,
	}, "")
	if err != nil {
		fmt.Printf("error getting context\n%s\n----\n", err.Error())
		return &[]Queue{}, err
	}
	var queue []Queue
	for _, qResponse := range queueResponse {
		q, err := responseToQueue(qResponse)
		if err != nil {
			return nil, err
		}
		if !(pendingOnly && q.Started != "") {
			queue = append(queue, *q)
		}
	}
	return &queue, nil
}

func responseToQueue(queueResponse map[string]interface{}) (*Queue, error) {
	fmt.Printf("contextResponse\n%+v\n----\n", queueResponse)
	noteIntf := queueResponse["notesString"]
	noteString, ok := noteIntf.(string)
	qJSON, err := json.Marshal(queueResponse)
	if err != nil {
		fmt.Printf("error marshalling queue\n%s\n----\n", err.Error())
		return &Queue{}, err
	}
	currentQueue := &Queue{}
	err = json.Unmarshal(qJSON, currentQueue)
	if err != nil {
		fmt.Printf("error unmarshalling queue\n%s\n----\n", err.Error())
		return &Queue{}, err
	}
	fmt.Printf("currentQueue\n%+v\n----\n", currentQueue)
	if currentQueue.UserId == "" {
		return &Queue{}, errors.New("not found")
	}

	if ok {
		noteBytes := json.RawMessage{}
		fmt.Printf("noteString\n%+v\n----\n", noteString)
		err = json.Unmarshal([]byte(noteString), &noteBytes)
		if err != nil {
			fmt.Printf("error unmarshalling notes\n%s\n----\n", err.Error())
			return &Queue{}, err
		}
		fmt.Printf("noteBytes\n%+v\n----\n", noteBytes)
		fmt.Printf("string(noteBytes)\n%+v\n----\n", string(noteBytes))
		currentQueue.Notes = noteBytes
		currentQueue.NoteString = noteString
		if noteString == "[]" {
			currentQueue.Notes = json.RawMessage{}
		}
	}

	fmt.Printf("currentQueue at end of responseToQueue\n%+v\n----\n", currentQueue)
	return currentQueue, nil
}

func saveQueue(q *Queue) error {
	createdTime, err := time.Parse(SkDateFormat, q.Id)
	expires := createdTime.AddDate(1, 0, 0).Unix()
	if err != nil {
		return err
	}
	if q.Started != "" {
		startedTime, err := time.Parse(SkDateFormat, q.Started)
		if err != nil {
			return err
		}
		expires = startedTime.AddDate(0, 3, 0).Unix()
	}
	payload := map[string]interface{}{
		pkString:      q.UserId,
		qIdString:     q.Id,
		"name":        q.Name,
		"notesString": q.NoteString,
		"started":     q.Started,
		"contextId":   q.ContextId,
		"expires":     expires,
	}

	err = utils.AddRecordDynamic(utils.Config{
		PK: utils.NameVal{
			Name:  pkString,
			Value: q.UserId,
		},
		SK: utils.NameVal{
			Name:  qIdString,
			Value: q.Id,
		},
		TableName: utils.QueueTableName,
	}, payload)
	return err
}
