package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ctxtype "main/types"
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

	item[conf.PK.Name] = &types.AttributeValueMemberS{
		Value: conf.PK.Value,
	}
	item[conf.SK.Name] = &types.AttributeValueMemberS{
		Value: conf.SK.Value,
	}
	item["lastUpdated"] = &types.AttributeValueMemberS{
		Value: lastUpdated,
	}
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
		case ctxtype.Document:
			v := map[string]types.AttributeValue{}
			v["realtivePath"] = &types.AttributeValueMemberS{
				Value: val.RealtivePath,
			}
			v["github"] = &types.AttributeValueMemberS{
				Value: val.Github,
			}
			item["document"] = &types.AttributeValueMemberM{
				Value: v,
			}
		default:
			fmt.Printf("I dunno what dis is!!!\n%+v\n", val)
			fmt.Printf("well actually, it's  of type %s\n", reflect.TypeOf(val))
		}
	}
	return item, nil
}

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
	}
	return output
}

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
		output = append(output, itemMap)
	}

	return output, nil
}

func getAttributeAsType(union types.AttributeValue) (any, error) {
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
