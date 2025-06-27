package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"personalisation-poc/model"
	"personalisation-poc/repository/ddb"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/guregu/dynamo/v2"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	tableName          = "user_profiles"
	awsRegion          = "us-east-1"
	awsAccessKeyID     = "dummy"
	awsSecretAccessKey = "dummy"
)

type Suite struct {
	suite.Suite
	pool         *dockertest.Pool
	dynamoRes    *dockertest.Resource
	dynamoClient *dynamodb.Client
	baseURL      string
	server       *server
	httpServer   *http.Server
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
	var err error

	// Create docker pool
	s.pool, err = dockertest.NewPool("")
	s.Require().NoError(err, "Failed to connect to docker")

	// Test docker connectivity
	err = s.pool.Client.Ping()
	s.Require().NoError(err, "Failed to ping docker")

	// Start DynamoDB container
	s.startDynamoDB()

	// Initialize DynamoDB table
	s.initDynamoDBTable()

	// Start the application server
	s.startApplication()

	// Wait for application to be ready
	s.waitForApplication()
}

func (s *Suite) TearDownSuite() {
	// Stop the HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}

	if s.dynamoRes != nil {
		err := s.pool.Purge(s.dynamoRes)
		s.Require().NoError(err, "Failed to remove DynamoDB container")
	}
}

func (s *Suite) startDynamoDB() {
	var err error

	// Start DynamoDB Local container
	s.dynamoRes, err = s.pool.Run("amazon/dynamodb-local", "latest", []string{
		"AWS_ACCESS_KEY_ID=" + awsAccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + awsSecretAccessKey,
		"AWS_REGION=" + awsRegion,
	})
	s.Require().NoError(err, "Failed to start DynamoDB container")

	// Get the mapped port
	dynamoPort := s.dynamoRes.GetPort("8000/tcp")
	dynamoEndpoint := fmt.Sprintf("http://localhost:%s", dynamoPort)

	client := dynamodb.New(dynamodb.Options{}, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(dynamoEndpoint)
		o.Region = awsRegion
		o.Credentials = credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, "")
	})
	s.dynamoClient = client

	// Wait for DynamoDB to be ready
	s.pool.Retry(func() error {
		_, err := s.dynamoClient.ListTables(context.Background(), &dynamodb.ListTablesInput{})
		return err
	})
}

func (s *Suite) initDynamoDBTable() {
	// Create the user_profiles table as defined in docker-compose.yml
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := s.dynamoClient.CreateTable(context.Background(), createTableInput)
	s.Require().NoError(err, "Failed to create DynamoDB table")

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(s.dynamoClient)
	err = waiter.Wait(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 2*time.Minute)
	s.Require().NoError(err, "Table did not become active in time")
}

func (s *Suite) startApplication() {
	// Get DynamoDB port for environment variables
	dynamoPort := s.dynamoRes.GetPort("8000/tcp")
	dynamoEndpoint := fmt.Sprintf("http://localhost:%s", dynamoPort)

	// Create DynamoDB connection using dynamo/v2 like in main.go
	cfg := aws.Config{
		Region:      awsRegion,
		Credentials: credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, ""),
	}
	db := dynamo.New(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(dynamoEndpoint)
		o.Region = awsRegion
		o.Credentials = credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, "")
	})
	table := db.Table(tableName)

	// Create repository
	repo := ddb.NewDB(table)

	// Create logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create server using newServer function like in main.go
	s.server = newServer(repo, log)

	// Start HTTP server on a random available port
	s.httpServer = &http.Server{
		Addr:    ":0", // Let OS choose available port
		Handler: s.server.router,
	}

	// Start the server and get the actual port
	listener, err := net.Listen("tcp", ":0")
	s.Require().NoError(err, "Failed to create listener")

	port := listener.Addr().(*net.TCPAddr).Port
	s.httpServer.Addr = fmt.Sprintf(":%d", port)

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.Require().NoError(err, "Failed to start HTTP server")
		}
	}()

	// Set base URL for API calls
	s.baseURL = fmt.Sprintf("http://localhost:%d/api/v1", port)
}

func (s *Suite) waitForApplication() {
	// Wait for application to be ready
	maxRetries := 30
	for range maxRetries {
		resp, err := http.Get(s.baseURL + "/profile/test")
		if err == nil {
			resp.Body.Close()
			// We expect 404 for non-existent profile, which means the API is working
			if resp.StatusCode == http.StatusNotFound {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	s.Require().FailNow("Application did not become ready in time")
}

func (s *Suite) TestProfile() {
	// Create test profile
	profileID := uuid.New()
	testProfile := model.Profile{
		ID:   profileID,
		Tags: []string{"test_tag", "profile_test"},
		Segments: []model.Segment{
			{
				Type: "morning_categories",
				Categories: []model.Category{
					{ID: "news", Score: 0.85},
					{ID: "sports", Score: 0.65},
				},
				TopCategories: []string{"news", "sports"},
			},
			{
				Type: "evening_categories",
				Categories: []model.Category{
					{ID: "entertainment", Score: 0.92},
					{ID: "tech", Score: 0.78},
				},
				TopCategories: []string{"entertainment", "tech"},
			},
		},
	}

	// Test 1: Create profile
	s.T().Run("CreateProfile", func(t *testing.T) {
		profileJSON, err := json.Marshal(testProfile)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", s.baseURL+"/profile", bytes.NewReader(profileJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		httpResp, err := client.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	})

	// Test 2: Get full profile
	s.T().Run("GetProfile", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/profile/" + profileID.String())
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedProfile model.Profile
		err = json.NewDecoder(resp.Body).Decode(&retrievedProfile)
		require.NoError(t, err)

		require.Equal(t, profileID, retrievedProfile.ID)
		require.ElementsMatch(t, testProfile.Tags, retrievedProfile.Tags)
		require.Len(t, retrievedProfile.Segments, 2)
	})

	// Test 3: Verify profile data integrity
	s.T().Run("VerifyProfileData", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/profile/" + profileID.String())
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedProfile model.Profile
		err = json.NewDecoder(resp.Body).Decode(&retrievedProfile)
		require.NoError(t, err)

		// Verify segments are stored correctly within the profile
		require.Len(t, retrievedProfile.Segments, 2)

		// Check morning categories segment
		morningSegment := retrievedProfile.Segments[0]
		if morningSegment.Type != "morning_categories" {
			morningSegment = retrievedProfile.Segments[1]
		}
		require.Equal(t, "morning_categories", morningSegment.Type)
		require.Len(t, morningSegment.Categories, 2)
		require.Equal(t, "news", morningSegment.Categories[0].ID)
		require.Equal(t, 0.85, morningSegment.Categories[0].Score)
		require.ElementsMatch(t, []string{"news", "sports"}, morningSegment.TopCategories)

		// Check evening categories segment
		eveningSegment := retrievedProfile.Segments[1]
		if eveningSegment.Type != "evening_categories" {
			eveningSegment = retrievedProfile.Segments[0]
		}
		require.Equal(t, "evening_categories", eveningSegment.Type)
		require.Len(t, eveningSegment.Categories, 2)
		require.ElementsMatch(t, []string{"entertainment", "tech"}, eveningSegment.TopCategories)
	})

	// Test 4: Get tags
	s.T().Run("GetTags", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/profile/" + profileID.String() + "/tags")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var tags []string
		err = json.NewDecoder(resp.Body).Decode(&tags)
		require.NoError(t, err)

		require.ElementsMatch(t, []string{"test_tag", "profile_test"}, tags)
	})

	// Test 5: Test non-existent profile
	s.T().Run("GetNonExistentProfile", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/profile/" + uuid.New().String())
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func (s *Suite) TestBlob() {
	// Create test profile for blob
	blobProfileID := uuid.New()
	testBlobProfile := model.Profile{
		ID:   blobProfileID,
		Tags: []string{"blob_tag", "test_blob"},
		Segments: []model.Segment{
			{
				Type: "blob_categories",
				Categories: []model.Category{
					{ID: "finance", Score: 0.88},
					{ID: "health", Score: 0.72},
				},
				TopCategories: []string{"finance", "health"},
			},
		},
	}

	// Test 1: Create blob
	s.T().Run("CreateBlob", func(t *testing.T) {
		blobJSON, err := json.Marshal(testBlobProfile)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", s.baseURL+"/blob", bytes.NewReader(blobJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Test 2: Get blob
	s.T().Run("GetBlob", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/blob/" + blobProfileID.String())
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var blobData []byte
		err = json.NewDecoder(resp.Body).Decode(&blobData)
		require.NoError(t, err)

		// Verify we can decode the blob back to the original profile
		var retrievedProfile model.Profile
		err = json.Unmarshal(blobData, &retrievedProfile)
		require.NoError(t, err)

		require.Equal(t, blobProfileID, retrievedProfile.ID)
		require.Equal(t, testBlobProfile.Tags, retrievedProfile.Tags)
	})

	// Test 3: Get segments from blob
	s.T().Run("GetSegmentsFromBlob", func(t *testing.T) {
		resp, err := http.Get(s.baseURL + "/blob/" + blobProfileID.String() + "/segments")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var segments []model.Segment
		err = json.NewDecoder(resp.Body).Decode(&segments)
		require.NoError(t, err)

		require.Len(t, segments, 1)
		require.Equal(t, "blob_categories", segments[0].Type)
		require.Len(t, segments[0].Categories, 2)
		require.Equal(t, "finance", segments[0].Categories[0].ID)
	})
}
