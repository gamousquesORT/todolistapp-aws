package services

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"todolistapp/internal/models"
)

type DynamoDBService struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBService(region, tableName string) (*DynamoDBService, error) {
	config := &aws.Config{
		Region: aws.String(region),
	}
	
	// Check if we're using local DynamoDB
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		config.Endpoint = aws.String(endpoint)
		config.DisableSSL = aws.Bool(true)
	}
	
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	service := &DynamoDBService{
		db:        dynamodb.New(sess),
		tableName: tableName,
	}
	
	// Create table if using local DynamoDB
	if os.Getenv("DYNAMODB_ENDPOINT") != "" {
		err = service.createTableIfNotExists()
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %v", err)
		}
	}

	return service, nil
}

func (s *DynamoDBService) CreateTodo(todo *models.Todo) error {
	av, err := dynamodbattribute.MarshalMap(todo)
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %v", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(s.tableName),
	}

	_, err = s.db.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to create todo: %v", err)
	}

	return nil
}

func (s *DynamoDBService) GetAllTodos() ([]models.Todo, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	}

	result, err := s.db.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan todos: %v", err)
	}

	var todos []models.Todo
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &todos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal todos: %v", err)
	}

	return todos, nil
}

func (s *DynamoDBService) GetTodoByID(id string) (*models.Todo, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := s.db.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("todo not found")
	}

	var todo models.Todo
	err = dynamodbattribute.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal todo: %v", err)
	}

	return &todo, nil
}

func (s *DynamoDBService) UpdateTodo(id string, updates map[string]interface{}) (*models.Todo, error) {
	updates["updated_at"] = time.Now()

	updateExpr := "SET "
	exprAttrNames := make(map[string]*string)
	exprAttrValues := make(map[string]*dynamodb.AttributeValue)

	i := 0
	for key, value := range updates {
		if i > 0 {
			updateExpr += ", "
		}
		attrName := fmt.Sprintf("#%s", key)
		attrValue := fmt.Sprintf(":%s", key)
		updateExpr += fmt.Sprintf("%s = %s", attrName, attrValue)

		exprAttrNames[attrName] = aws.String(key)
		av, err := dynamodbattribute.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal update value: %v", err)
		}
		exprAttrValues[attrValue] = av
		i++
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
		ReturnValues:              aws.String("ALL_NEW"),
	}

	result, err := s.db.UpdateItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %v", err)
	}

	var todo models.Todo
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &todo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated todo: %v", err)
	}

	return &todo, nil
}

func (s *DynamoDBService) DeleteTodo(id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	_, err := s.db.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %v", err)
	}

	return nil
}

func (s *DynamoDBService) createTableIfNotExists() error {
	// Check if table already exists
	_, err := s.db.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})
	if err == nil {
		// Table already exists
		return nil
	}

	// Create table
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(s.tableName),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
	}

	_, err = s.db.CreateTable(input)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Wait for table to be active
	return s.db.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})
}