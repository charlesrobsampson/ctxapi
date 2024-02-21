package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type (
	NameVal struct {
		Name  string
		Value string
	}
	Config struct {
		PK        NameVal
		SK        NameVal
		SKBetween Between
		TableName string
	}
	Between struct {
		Lower NameVal
		Upper NameVal
	}
)

// func AddRecord(conf Config, payload map[string]string) error {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	lastUpdated := time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	item := map[string]types.AttributeValue{
// 		conf.PK.Name: &types.AttributeValueMemberS{
// 			Value: conf.PK.Value,
// 		},
// 		conf.SK.Name: &types.AttributeValueMemberS{
// 			Value: conf.SK.Value,
// 		},
// 		"lastUpdated": &types.AttributeValueMemberS{
// 			Value: lastUpdated,
// 		},
// 	}

// 	for key, val := range payload {
// 		item[key] = &types.AttributeValueMemberS{
// 			Value: val,
// 		}
// 	}

// 	fmt.Printf("insert item\n%+v\n----\n", item)

// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &conf.TableName,
// 		Item:      item,
// 	})

// 	if err != nil {
// 		return nil
// 	}
// 	return err
// 	// return nil
// }

func AddRecordDynamic(conf Config, payload map[string]interface{}) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		o.Region = "us-west-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	lastUpdated := time.Now().UTC().Format("2006-01-02 15:04:05Z")

	item, err := makeItem(payload)
	if err != nil {
		fmt.Printf("error making item: %s\n", err)
	}

	// item := map[string]types.AttributeValue{
	// conf.PK.Name: &types.AttributeValueMemberS{
	// 	Value: conf.PK.Value,
	// },
	// 	conf.SK.Name: &types.AttributeValueMemberS{
	// 		Value: conf.SK.Value,
	// 	},
	// 	"lastUpdated": &types.AttributeValueMemberS{
	// 		Value: lastUpdated,
	// 	},
	// }
	item[conf.PK.Name] = &types.AttributeValueMemberS{
		Value: conf.PK.Value,
	}
	item[conf.SK.Name] = &types.AttributeValueMemberS{
		Value: conf.SK.Value,
	}
	item["lastUpdated"] = &types.AttributeValueMemberS{
		Value: lastUpdated,
	}

	// for key, val := range payload {
	// 	item[key] = &types.AttributeValueMemberS{
	// 		Value: val,
	// 	}
	// }

	fmt.Printf("insert item\n%+v\n----\n", item)

	output, err := svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:    &conf.TableName,
		Item:         item,
		ReturnValues: types.ReturnValueAllOld,
	})

	fmt.Printf("output\n%+v\n----\n", output)

	if err != nil {
		return nil
	}
	return err
	// return nil
}

func UpdateField(conf Config, field string, value interface{}) error {
	// test, _ := getAttributeAsType(conf, field)
	val, err := getAttributeValue(field, value)
	if err != nil {
		fmt.Printf("error getting attribute value: %s\n", err)
		return err
	}
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		o.Region = "us-west-1"
		return nil
	})
	if err != nil {
		panic(err)
	}
	svc := dynamodb.NewFromConfig(cfg)
	lastUpdated := time.Now().UTC().Format("2006-01-02 15:04:05")
	item := map[string]types.AttributeValue{
		conf.PK.Name: &types.AttributeValueMemberS{
			Value: conf.PK.Value,
		},
		conf.SK.Name: &types.AttributeValueMemberS{
			Value: conf.SK.Value,
		},
		"lastUpdated": &types.AttributeValueMemberS{
			Value: lastUpdated,
		},
	}
	expression := fmt.Sprintf("set #lastUpdated = :lastUpdated, #%s = :%s", field, field)
	_, err = svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        &conf.TableName,
		Key:              item,
		UpdateExpression: &expression,
		ExpressionAttributeNames: map[string]string{
			fmt.Sprintf("#%s", field): field,
			"#lastUpdated":            "lastUpdated",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", field): val[field],
			":lastUpdated":            &types.AttributeValueMemberS{Value: lastUpdated},
		},
	})
	if err != nil {
		return nil
	}
	return err
	// return nil
}

func getAttributeValue(field string, value any) (map[string]types.AttributeValue, error) {
	val := map[string]types.AttributeValue{}
	switch value := value.(type) {
	case string:
		val[field] = &types.AttributeValueMemberS{
			Value: value,
		}
	case int:
		val[field] = &types.AttributeValueMemberN{
			Value: strconv.Itoa(value),
		}
	case map[string]string:
		v := map[string]interface{}{}
		for key, s := range value {
			v[key] = s
		}
		vmap, _ := makeItem(v)
		val[field] = &types.AttributeValueMemberM{
			Value: vmap,
		}
	case bool:
		val[field] = &types.AttributeValueMemberBOOL{
			Value: value,
		}
	default:
		return val, fmt.Errorf("unsupported type: %s", reflect.TypeOf(value))
	}
	return val, nil
}

func makeItem(payload map[string]interface{}) (map[string]types.AttributeValue, error) {
	item := map[string]types.AttributeValue{}
	for key, val := range payload {
		fmt.Printf("key: '%s' val: '%+v'\nvalType: %+v\n", key, val, reflect.TypeOf(val))
		switch val := val.(type) {
		case string:
			item[key] = &types.AttributeValueMemberS{
				Value: val,
			}
		case int:
			item[key] = &types.AttributeValueMemberN{
				Value: strconv.Itoa(val),
			}
		case map[string]string:
			v := map[string]interface{}{}
			for key, s := range val {
				v[key] = s
			}
			vmap, _ := makeItem(v)
			item[key] = &types.AttributeValueMemberM{
				Value: vmap,
			}
		case bool:
			item[key] = &types.AttributeValueMemberBOOL{
				Value: val,
			}
		case map[string]interface{}:
			var v map[string]interface{}
			v = val
			vmap, _ := makeItem(v)
			item[key] = &types.AttributeValueMemberM{
				Value: vmap,
			}
		case []string:
			item[key] = &types.AttributeValueMemberSS{
				Value: val,
			}
		case json.RawMessage:
			item[key] = &types.AttributeValueMemberS{
				Value: string(val),
			}
		case int64:
			item[key] = &types.AttributeValueMemberN{
				Value: strconv.FormatInt(val, 10),
			}
		default:
			fmt.Printf("I dunno what dis is!!!\n%+v\n", val)
			fmt.Printf("well it's actually of type %s\n", reflect.TypeOf(val))
		}
	}
	return item, nil
}

// func Get(conf Config) (map[string]string, error) {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	output := map[string]string{}
// 	svc := dynamodb.NewFromConfig(cfg)

// 	getResponse, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: &conf.TableName,
// 		Key: map[string]types.AttributeValue{
// 			conf.PK.Name: &types.AttributeValueMemberS{
// 				Value: conf.PK.Value,
// 			},
// 			conf.SK.Name: &types.AttributeValueMemberS{
// 				Value: conf.SK.Value,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	item := getResponse.Item

// 	if len(item) == 0 {
// 		return output, nil
// 	}

// 	for key, val := range item {
// 		typed, err := getAttributeAsType(val)
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		output[key] = fmt.Sprintf("%v", typed)
// 	}

// 	return output, nil
// }

func GetDynamic(conf Config) (map[string]interface{}, error) {
	fmt.Println("getDynamic")
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		o.Region = "us-west-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	// output := map[string]interface{}{}
	svc := dynamodb.NewFromConfig(cfg)

	getResponse, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &conf.TableName,
		Key: map[string]types.AttributeValue{
			conf.PK.Name: &types.AttributeValueMemberS{
				Value: conf.PK.Value,
			},
			conf.SK.Name: &types.AttributeValueMemberS{
				Value: conf.SK.Value,
			},
		},
	})

	if err != nil {
		fmt.Printf("err\n%+v\n----\n", err)
	}

	item := getResponse.Item
	output := makeResponse(item)

	if len(item) == 0 {
		return output, nil
	}

	return output, nil
}

func makeResponse(item map[string]types.AttributeValue) map[string]interface{} {
	output := map[string]interface{}{}
	for key, val := range item {
		typed, err := getAttributeAsType(val)
		fmt.Printf("looking at\n%+v\n", typed)
		fmt.Printf("of type: %s\n", reflect.TypeOf(typed))
		if err != nil {
			fmt.Printf("error: %s", err)
		}
		switch val := typed.(type) {
		case string:
			output[key] = val
		case int:
			output[key] = val
		case bool:
			output[key] = val
		case map[string]types.AttributeValue:
			output[key] = makeResponse(val)
		case []string:
			output[key] = val
		default:
			fmt.Printf("missing case for type: %s", reflect.TypeOf(typed))
		}
		// output[key] = fmt.Sprintf("%v", typed)
	}
	return output
}

// func Query(conf Config) ([]map[string]string, error) {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	output := []map[string]string{}
// 	svc := dynamodb.NewFromConfig(cfg)

// 	queryString := "#pk = :pk AND begins_with ( #sk , :sk )"
// 	getResponse, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &conf.TableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeNames: map[string]string{
// 			"#pk": conf.PK.Name,
// 			"#sk": conf.SK.Name,
// 		},
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":pk": &types.AttributeValueMemberS{
// 				Value: conf.PK.Value,
// 			},
// 			":sk": &types.AttributeValueMemberS{
// 				Value: conf.SK.Value,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	items := getResponse.Items

// 	if len(items) == 0 {
// 		return output, nil
// 	}

// 	for _, item := range items {
// 		itemMap := map[string]string{}
// 		for key, val := range item {
// 			typed, err := getAttributeAsType(val)
// 			if err != nil {
// 				fmt.Printf("error: %s", err)
// 			}
// 			itemMap[key] = fmt.Sprintf("%v", typed)
// 		}
// 		output = append(output, itemMap)
// 	}

// 	return output, nil
// }

// func QueryDynamic(conf Config) ([]map[string]interface{}, error) {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Printf("query config\n%+v\n----\n", conf)
// 	output := []map[string]interface{}{}
// 	svc := dynamodb.NewFromConfig(cfg)
// 	queryString := ""
// 	params := &dynamodb.QueryInput{}
// 	if len(conf.SK.Value) > 0 {
// 		fmt.Println("getting with pk and sk")
// 		fmt.Printf("sk len: %d\n", len(conf.SK.Value))
// 		queryString = "#pk = :pk AND begins_with ( #sk , :sk )"
// 		params = &dynamodb.QueryInput{
// 			TableName:              &conf.TableName,
// 			KeyConditionExpression: &queryString,
// 			ExpressionAttributeNames: map[string]string{
// 				"#pk": conf.PK.Name,
// 				"#sk": conf.SK.Name,
// 			},
// 			ExpressionAttributeValues: map[string]types.AttributeValue{
// 				":pk": &types.AttributeValueMemberS{
// 					Value: conf.PK.Value,
// 				},
// 				":sk": &types.AttributeValueMemberS{
// 					Value: conf.SK.Value,
// 				},
// 			},
// 		}
// 	} else {
// 		fmt.Println("getting with only pk")
// 		queryString = "#pk = :pk"
// 		params = &dynamodb.QueryInput{
// 			TableName:              &conf.TableName,
// 			KeyConditionExpression: &queryString,
// 			ExpressionAttributeNames: map[string]string{
// 				"#pk": conf.PK.Name,
// 			},
// 			ExpressionAttributeValues: map[string]types.AttributeValue{
// 				":pk": &types.AttributeValueMemberS{
// 					Value: conf.PK.Value,
// 				},
// 			},
// 		}

// 	}
// 	getResponse, err := svc.Query(ctx, params)

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	items := getResponse.Items

// 	if len(items) == 0 {
// 		return output, nil
// 	}

// 	for _, item := range items {
// 		itemMap := makeResponse(item)
// 		// itemMap := map[string]string{}
// 		// for key, val := range item {
// 		// 	typed, err := getAttributeAsType(val)
// 		// 	if err != nil {
// 		// 		fmt.Printf("error: %s", err)
// 		// 	}
// 		// 	itemMap[key] = fmt.Sprintf("%v", typed)
// 		// }
// 		output = append(output, itemMap)
// 	}

// 	return output, nil
// }

func QueryWithFilter(conf Config, filter string) ([]map[string]interface{}, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		o.Region = "us-west-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("query config\n%+v\n----\n", conf)
	output := []map[string]interface{}{}
	svc := dynamodb.NewFromConfig(cfg)
	queryString := ""
	params := &dynamodb.QueryInput{}
	if len(conf.SK.Value) > 0 {
		fmt.Println("getting with pk and sk")
		fmt.Printf("sk len: %d\n", len(conf.SK.Value))
		queryString = "#pk = :pk AND begins_with ( #sk , :sk )"
		params = &dynamodb.QueryInput{
			TableName:              &conf.TableName,
			KeyConditionExpression: &queryString,
			ExpressionAttributeNames: map[string]string{
				"#pk": conf.PK.Name,
				"#sk": conf.SK.Name,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: conf.PK.Value,
				},
				":sk": &types.AttributeValueMemberS{
					Value: conf.SK.Value,
				},
			},
		}
		if filter != "" {
			params.FilterExpression = &filter
		}
		// params.ExpressionAttributeNames[""] = ""
	} else if conf.SKBetween.Lower.Value != "" {
		queryString = "#pk = :pk AND #sk BETWEEN :skLower AND :skUpper"
		params.TableName = &conf.TableName
		params.KeyConditionExpression = &queryString
		params.ExpressionAttributeNames = map[string]string{
			"#pk": conf.PK.Name,
			"#sk": conf.SK.Name,
		}
		params.ExpressionAttributeValues = map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{
				Value: conf.PK.Value,
			},
			":skLower": &types.AttributeValueMemberS{
				Value: conf.SKBetween.Lower.Value,
			},
			":skUpper": &types.AttributeValueMemberS{
				Value: conf.SKBetween.Upper.Value,
			},
		}
		if filter != "" {
			params.FilterExpression = &filter
		}
	} else {
		fmt.Println("getting with only pk")
		queryString = "#pk = :pk"
		params = &dynamodb.QueryInput{
			TableName:              &conf.TableName,
			KeyConditionExpression: &queryString,
			ExpressionAttributeNames: map[string]string{
				"#pk": conf.PK.Name,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: conf.PK.Value,
				},
			},
		}
		if filter != "" {
			params.FilterExpression = &filter
		}
	}
	getResponse, err := svc.Query(ctx, params)

	if err != nil {
		fmt.Printf("err\n%+v\n----\n", err)
	}

	items := getResponse.Items

	if len(items) == 0 {
		return output, nil
	}

	for _, item := range items {
		itemMap := makeResponse(item)
		// itemMap := map[string]string{}
		// for key, val := range item {
		// 	typed, err := getAttributeAsType(val)
		// 	if err != nil {
		// 		fmt.Printf("error: %s", err)
		// 	}
		// 	itemMap[key] = fmt.Sprintf("%v", typed)
		// }
		output = append(output, itemMap)
	}

	return output, nil
}

// func ListAllTasks(home string) string {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	// taskOutput, err := svc.Scan(ctx, &dynamodb.ScanInput{
// 	// 	TableName: &taskTableName,
// 	// })
// 	queryString := "home = :home"
// 	taskOutput, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &taskTableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	fmt.Printf("taskOutput\n%+v\n-----\n", taskOutput)
// 	items := taskOutput.Items
// 	tasks := []Task{}
// 	for _, item := range items {
// 		task := Task{}
// 		// home, err := getAttributeAsType(item["home"])
// 		// if err != nil {
// 		// 	fmt.Printf("error: %s", err)
// 		// }
// 		name, err := getAttributeAsType(item["name"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		frequency, err := getAttributeAsType(item["frequency"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		unit, err := getAttributeAsType(item["unit"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		lastUpdated, err := getAttributeAsType(item["lastUpdated"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		// task.Home = fmt.Sprintf("%v", home)
// 		task.Name = fmt.Sprintf("%v", name)
// 		task.Frequency = fmt.Sprintf("%v", frequency)
// 		task.Unit = fmt.Sprintf("%v", unit)
// 		task.LastUpdated = fmt.Sprintf("%v", lastUpdated)
// 		if name != "home" {
// 			tasks = append(tasks, task)
// 		}
// 	}
// 	output, err := json.Marshal(tasks)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func ListAllLists(home, person string) string {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	// taskOutput, err := svc.Scan(ctx, &dynamodb.ScanInput{
// 	// 	TableName: &taskTableName,
// 	// })
// 	queryString := "homePerson = :homePerson"
// 	listOutput, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &listTableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":homePerson": &types.AttributeValueMemberS{
// 				Value: fmt.Sprintf("%s#%s", home, person),
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	fmt.Printf("listOutput\n%+v\n-----\n", listOutput)
// 	items := listOutput.Items
// 	todoList := []List{}
// 	for _, item := range items {
// 		listitem := List{}
// 		// home, err := getAttributeAsType(item["home"])
// 		// if err != nil {
// 		// 	fmt.Printf("error: %s", err)
// 		// }
// 		name, err := getAttributeAsType(item["name"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		list, err := getAttributeAsType(item["list"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		added, err := getAttributeAsType(item["added"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		lastUpdated, err := getAttributeAsType(item["lastUpdated"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		listitem.Name = fmt.Sprintf("%v", name)
// 		listitem.List = fmt.Sprintf("%v", list)
// 		listitem.Added = fmt.Sprintf("%v", added)
// 		listitem.LastUpdated = fmt.Sprintf("%v", lastUpdated)
// 		if name != "home" {
// 			todoList = append(todoList, listitem)
// 		}
// 	}
// 	output, err := json.Marshal(todoList)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func ListAllPersonalTasks(home, person string) string {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	// taskOutput, err := svc.Scan(ctx, &dynamodb.ScanInput{
// 	// 	TableName: &taskTableName,
// 	// })
// 	queryString := "home = :home AND begins_with ( nameTask , :person )"
// 	taskOutput, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &personalTaskTableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			":person": &types.AttributeValueMemberS{
// 				Value: person,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	// fmt.Printf("taskOutput\n%+v\n-----\n", taskOutput)
// 	items := taskOutput.Items
// 	tasks := []Task{}
// 	for _, item := range items {
// 		task := Task{}
// 		// home, err := getAttributeAsType(item["home"])
// 		// if err != nil {
// 		// 	fmt.Printf("error: %s", err)
// 		// }
// 		name, err := getAttributeAsType(item["nameTask"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		frequency, err := getAttributeAsType(item["frequency"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		unit, err := getAttributeAsType(item["unit"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		lastUpdated, err := getAttributeAsType(item["lastUpdated"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		// task.Home = fmt.Sprintf("%v", home)
// 		task.Name = strings.Split(fmt.Sprintf("%v", name), "#")[1]
// 		task.Frequency = fmt.Sprintf("%v", frequency)
// 		task.Unit = fmt.Sprintf("%v", unit)
// 		task.LastUpdated = fmt.Sprintf("%v", lastUpdated)
// 		if name != "home" {
// 			tasks = append(tasks, task)
// 		}
// 	}
// 	output, err := json.Marshal(tasks)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetAllHistory(home string) string {
// 	tasksString := ListAllTasks(home)
// 	var tasks []Task
// 	history := map[string][]map[string]string{}
// 	err := json.Unmarshal([]byte(tasksString), &tasks)
// 	if err != nil {
// 		panic(fmt.Sprintf("error unmarshaling json: %s", err))
// 	}
// 	for _, task := range tasks {
// 		historyString := GetHistoryByTask(home, task.Name)
// 		var hist []map[string]string
// 		err = json.Unmarshal([]byte(historyString), &hist)
// 		if err != nil {
// 			panic(fmt.Sprintf("error unmarshaling json: %s", err))
// 		}
// 		history[task.Name] = hist
// 	}
// 	output, err := json.Marshal(history)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetAllPersonalHistory(home, person string) string {
// 	tasksString := ListAllPersonalTasks(home, person)
// 	var tasks []Task
// 	history := map[string][]map[string]string{}
// 	err := json.Unmarshal([]byte(tasksString), &tasks)
// 	if err != nil {
// 		panic(fmt.Sprintf("error unmarshaling json: %s", err))
// 	}
// 	for _, task := range tasks {
// 		historyString := GetPersonalHistoryByTask(home, task.Name, person)
// 		var hist []map[string]string
// 		err = json.Unmarshal([]byte(historyString), &hist)
// 		if err != nil {
// 			panic(fmt.Sprintf("error unmarshaling json: %s", err))
// 		}
// 		history[task.Name] = hist
// 	}
// 	output, err := json.Marshal(history)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func ListAllHomes() string {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	queryString := "#n = :home"
// 	homeOutput, err := svc.Scan(ctx, &dynamodb.ScanInput{
// 		TableName:        &taskTableName,
// 		FilterExpression: &queryString,
// 		ExpressionAttributeNames: map[string]string{
// 			"#n": "name",
// 		},
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":home": &types.AttributeValueMemberS{
// 				Value: "home",
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	fmt.Printf("homeOutput\n%+v\n----\n", homeOutput)
// 	items := homeOutput.Items
// 	homes := []Task{}
// 	for _, item := range items {
// 		h := Task{}
// 		home, err := getAttributeAsType(item["home"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		// name, err := getAttributeAsType(item["name"])
// 		// if err != nil {
// 		// 	fmt.Printf("error: %s", err)
// 		// }
// 		lastUpdated, err := getAttributeAsType(item["lastUpdated"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		h.Home = fmt.Sprintf("%v", home)
// 		// home.Name = fmt.Sprintf("%v", name)
// 		h.LastUpdated = fmt.Sprintf("%v", lastUpdated)
// 		homes = append(homes, h)
// 	}
// 	output, err := json.Marshal(homes)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetTask(home, taskName string) string {
// 	var task = Task{
// 		Name: taskName,
// 	}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	taskOutput, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: &taskTableName,
// 		Key: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: taskName,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	resultTask := taskOutput.Item

// 	if len(resultTask) == 0 {
// 		return "{}"
// 	}

// 	frequency, err := getAttributeAsType(resultTask["frequency"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	unit, err := getAttributeAsType(resultTask["unit"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	lastUpdated, err := getAttributeAsType(resultTask["lastUpdated"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	task.Frequency = fmt.Sprintf("%v", frequency)
// 	task.Unit = fmt.Sprintf("%v", unit)
// 	task.LastUpdated = fmt.Sprintf("%v", lastUpdated)

// 	output, err := json.Marshal(task)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetPersonalList(home, person, list string) string {
// 	var todoList = List{
// 		Name: list,
// 	}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	taskOutput, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: &listTableName,
// 		Key: map[string]types.AttributeValue{
// 			"homePerson": &types.AttributeValueMemberS{
// 				Value: fmt.Sprintf("%s#%s", home, person),
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: list,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	resultTask := taskOutput.Item

// 	if len(resultTask) == 0 {
// 		return "{}"
// 	}

// 	listItems, err := getAttributeAsType(resultTask["list"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	added, err := getAttributeAsType(resultTask["added"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	lastUpdated, err := getAttributeAsType(resultTask["lastUpdated"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	// listItemsBytes, err := json.Marshal(listItems)
// 	// if err != nil {
// 	// 	panic(fmt.Sprintf("error marshaling list: %s", err))
// 	// }
// 	// todoList.List = string(listItemsBytes)
// 	todoList.List = fmt.Sprintf("%v", listItems)
// 	todoList.Added = fmt.Sprintf("%v", added)
// 	todoList.LastUpdated = fmt.Sprintf("%v", lastUpdated)

// 	output, err := json.Marshal(todoList)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetPersonalTask(home, taskName, person string) string {
// 	var task = Task{
// 		Name: taskName,
// 	}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	taskOutput, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: &personalTaskTableName,
// 		Key: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			"nameTask": &types.AttributeValueMemberS{
// 				Value: fmt.Sprintf("%s#%s", person, taskName),
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	resultTask := taskOutput.Item

// 	if len(resultTask) == 0 {
// 		return "{}"
// 	}

// 	frequency, err := getAttributeAsType(resultTask["frequency"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	unit, err := getAttributeAsType(resultTask["unit"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	lastUpdated, err := getAttributeAsType(resultTask["lastUpdated"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}
// 	task.Frequency = fmt.Sprintf("%v", frequency)
// 	task.Unit = fmt.Sprintf("%v", unit)
// 	task.LastUpdated = fmt.Sprintf("%v", lastUpdated)

// 	output, err := json.Marshal(task)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetHome(home string) string {
// 	var h = Task{
// 		Home: home,
// 	}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	taskOutput, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
// 		TableName: &taskTableName,
// 		Key: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: "home",
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	resultTask := taskOutput.Item

// 	lastUpdated, err := getAttributeAsType(resultTask["lastUpdated"])
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	h.LastUpdated = fmt.Sprintf("%v", lastUpdated)

// 	output, err := json.Marshal(h)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetHistoryByTask(home, task string) string {
// 	var history = []map[string]string{}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	queryString := "home = :home AND begins_with ( taskTime , :taskTime )"
// 	historyOutput, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &histTableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			":taskTime": &types.AttributeValueMemberS{
// 				Value: task,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	for _, item := range historyOutput.Items {
// 		h := map[string]string{}
// 		taskTime, err := getAttributeAsType(item["taskTime"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		tt := fmt.Sprintf("%v", taskTime)
// 		h["timestamp"] = strings.Split(tt, "#")[1]
// 		who, err := getAttributeAsType(item["who"])
// 		if err == nil {
// 			h["doneBy"] = fmt.Sprintf("%v", who)
// 		}
// 		history = append(history, h)
// 	}

// 	sort.SliceStable(history, func(i, j int) bool {
// 		return history[i]["timestamp"] > history[j]["timestamp"]
// 	})

// 	output, err := json.Marshal(history)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	return string(output)
// }

// func GetPersonalHistoryByTask(home, task, person string) string {
// 	var history = []map[string]string{}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	queryString := "home = :home AND begins_with ( taskTime , :taskTime )"
// 	historyOutput, err := svc.Query(ctx, &dynamodb.QueryInput{
// 		TableName:              &personalHistTableName,
// 		KeyConditionExpression: &queryString,
// 		ExpressionAttributeValues: map[string]types.AttributeValue{
// 			":home": &types.AttributeValueMemberS{
// 				Value: home,
// 			},
// 			":taskTime": &types.AttributeValueMemberS{
// 				Value: fmt.Sprintf("%s#%s", person, task),
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Printf("err\n%+v\n----\n", err)
// 	}

// 	for _, item := range historyOutput.Items {
// 		h := map[string]string{}
// 		taskTime, err := getAttributeAsType(item["taskTime"])
// 		if err != nil {
// 			fmt.Printf("error: %s", err)
// 		}
// 		tt := fmt.Sprintf("%v", taskTime)
// 		h["timestamp"] = strings.Split(tt, "#")[2]
// 		who, err := getAttributeAsType(item["who"])
// 		if err == nil {
// 			h["doneBy"] = fmt.Sprintf("%v", who)
// 		}
// 		history = append(history, h)
// 	}

// 	sort.SliceStable(history, func(i, j int) bool {
// 		return history[i]["timestamp"] > history[j]["timestamp"]
// 	})

// 	output, err := json.Marshal(history)
// 	if err != nil {
// 		fmt.Printf("error: %s", err)
// 	}

// 	fmt.Printf("output\n%s\n----\n", output)
// 	return string(output)
// }

// func AddTask(task Task) error {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	task.LastUpdated = time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &taskTableName,
// 		Item: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: task.Home,
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: task.Name,
// 			},
// 			"frequency": &types.AttributeValueMemberS{
// 				Value: task.Frequency,
// 			},
// 			"unit": &types.AttributeValueMemberS{
// 				Value: task.Unit,
// 			},
// 			"lastUpdated": &types.AttributeValueMemberS{
// 				Value: task.LastUpdated,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return nil
// 	}
// 	return err
// }

// func UpdateList(home, person, name, list string) error {
// 	existingString := GetPersonalList(home, person, name)
// 	var existing List
// 	err := json.Unmarshal([]byte(existingString), &existing)
// 	if err != nil {
// 		panic(fmt.Sprintf("error unmarshaling json body: %s", err))
// 	}
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	now := time.Now().UTC().Format("2006-01-02 15:04:05Z")
// 	fmt.Printf("added\n%+v\n----\n", existing.Added)
// 	if existing.Added == "" {
// 		existing.Added = now
// 	}
// 	fmt.Printf("added\n%+v\n----\n", existing.Added)
// 	existing.LastUpdated = now

// 	homePerson := home
// 	if person != "" {
// 		homePerson += fmt.Sprintf("#%s", person)
// 	}

// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &listTableName,
// 		Item: map[string]types.AttributeValue{
// 			"homePerson": &types.AttributeValueMemberS{
// 				Value: homePerson,
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: name,
// 			},
// 			"list": &types.AttributeValueMemberS{
// 				Value: list,
// 			},
// 			"added": &types.AttributeValueMemberS{
// 				Value: existing.Added,
// 			},
// 			"lastUpdated": &types.AttributeValueMemberS{
// 				Value: existing.LastUpdated,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return nil
// 	}
// 	return err
// }

// func AddPersonalTask(task Task) error {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	task.LastUpdated = time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &personalTaskTableName,
// 		Item: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: task.Home,
// 			},
// 			"nameTask": &types.AttributeValueMemberS{
// 				Value: task.Name,
// 			},
// 			"frequency": &types.AttributeValueMemberS{
// 				Value: task.Frequency,
// 			},
// 			"unit": &types.AttributeValueMemberS{
// 				Value: task.Unit,
// 			},
// 			"lastUpdated": &types.AttributeValueMemberS{
// 				Value: task.LastUpdated,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return nil
// 	}
// 	return err
// }

// func AddHome(home Task) error {
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	home.LastUpdated = time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &taskTableName,
// 		Item: map[string]types.AttributeValue{
// 			"home": &types.AttributeValueMemberS{
// 				Value: home.Home,
// 			},
// 			"name": &types.AttributeValueMemberS{
// 				Value: home.Name,
// 			},
// 			"lastUpdated": &types.AttributeValueMemberS{
// 				Value: home.LastUpdated,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return nil
// 	}
// 	return err
// }

// func AddDone(done HistoryBody) (string, error) {
// 	fmt.Printf("adding\n%+v\n----\n", done)
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	done.Time = time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	taskTime := fmt.Sprintf("%s#%s", done.Task, done.Time)
// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &histTableName,
// 		Item: map[string]types.AttributeValue{
// 			"group": &types.AttributeValueMemberS{
// 				Value: done.Group,
// 			},
// 			"taskTime": &types.AttributeValueMemberS{
// 				Value: taskTime,
// 			},
// 			"who": &types.AttributeValueMemberS{
// 				Value: done.Who,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return taskTime, nil
// 	}
// 	return taskTime, err
// }

// func AddDonePersonal(done HistoryBody, person string) (string, error) {
// 	fmt.Printf("adding\n%+v\n----\n", done)
// 	ctx := context.TODO()
// 	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
// 		o.Region = "us-west-1"
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	svc := dynamodb.NewFromConfig(cfg)

// 	done.Time = time.Now().UTC().Format("2006-01-02 15:04:05Z")

// 	taskTime := fmt.Sprintf("%s#%s", done.Task, done.Time)
// 	_, err = svc.PutItem(ctx, &dynamodb.PutItemInput{
// 		TableName: &personalHistTableName,
// 		Item: map[string]types.AttributeValue{
// 			"group": &types.AttributeValueMemberS{
// 				Value: done.Group,
// 			},
// 			"taskTime": &types.AttributeValueMemberS{
// 				Value: fmt.Sprintf("%s#%s", person, taskTime),
// 			},
// 			"who": &types.AttributeValueMemberS{
// 				Value: done.Who,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return taskTime, nil
// 	}
// 	return taskTime, err
// }

func getAttributeAsType(union types.AttributeValue) (any, error) {
	// var union types.AttributeValue
	// type switches can be used to check the union value
	switch v := union.(type) {
	case *types.AttributeValueMemberB:
		val := v.Value // Value is []byte
		return val, nil

	case *types.AttributeValueMemberBOOL:
		val := v.Value // Value is bool
		return val, nil

	case *types.AttributeValueMemberBS:
		val := v.Value // Value is [][]byte
		return val, nil

	case *types.AttributeValueMemberL:
		val := v.Value // Value is []types.AttributeValue
		return val, nil

	case *types.AttributeValueMemberM:
		val := v.Value // Value is map[string]types.AttributeValue
		return val, nil

	case *types.AttributeValueMemberN:
		val := v.Value // Value is string
		return val, nil

	case *types.AttributeValueMemberNS:
		val := v.Value // Value is []string
		return val, nil

	case *types.AttributeValueMemberNULL:
		val := v.Value // Value is bool
		return val, nil

	case *types.AttributeValueMemberS:
		val := v.Value // Value is string
		return val, nil

	case *types.AttributeValueMemberSS:
		val := v.Value // Value is []string
		return val, nil

	case *types.UnknownUnionMember:
		// fmt.Println("unknown tag:", v.Tag)
		return nil, fmt.Errorf("unknown tag: %v", v.Tag)

	default:
		// fmt.Println("union is nil or unknown type")
		return nil, errors.New("union is nil or unknown type")
	}
}
