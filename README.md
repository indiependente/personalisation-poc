# Personalisation Service - Single Table Design PoC

This is a **Proof of Concept (PoC)** service designed to demonstrate **DynamoDB Single Table Design** patterns and best practices. The service provides two distinct API patterns that showcase different approaches to implementing single table design.

## 🎯 Purpose

This PoC demonstrates:

- **Pure Single Table Design** via `/api/v1/profile` endpoints
- **Single Table Design with Blob Storage** via `/api/v1/blob` endpoints using DynamoDB's native map data types

## 🏗️ Architecture Overview

The service implements a user profile system where:

- **Users** have multiple **Segments** (e.g., morning_categories, evening_categories)
- **Segments** contain **Categories** with scores and **Top Categories** lists
- **Blob storage** allows storing complete profile JSON as native DynamoDB maps

### Single Table Schema

The service uses a single DynamoDB table with the following key structure:

| Partition Key (PK) | Sort Key (SK) | Type | Description |
|-------------------|---------------|------|-------------|
| `USER#{userId}` | `USER#{userId}` | `USER` | User profile metadata |
| `USER#{userId}` | `SEG#{segmentType}#{timestamp}` | `SEG` | Individual segments with timestamped versions |
| `USER#{userId}` | `BLOB#{userId}` | `BLOB` | Complete profile stored as DynamoDB map |

### guregu/dynamo Library

The service uses the [guregu/dynamo](https://github.com/guregu/dynamo) library to interact with DynamoDB.

The library is a wrapper around the AWS SDK v2 for Go that provides a more convenient API for working with DynamoDB.

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Running with Docker Compose

```bash
# Start all services (DynamoDB, Admin UI, and UserProfiles service)
docker-compose up -d

# Initialize the DynamoDB table (included in docker-compose)
# The table will be automatically created
```

Services will be available at:

- **UserProfiles API**: http://localhost:8080
- **DynamoDB Local**: http://localhost:8000
- **DynamoDB Admin UI**: http://localhost:8001

### Local Development

```bash
# Install dependencies
go mod download

# Set environment variables
export DYNAMO_ENDPOINT=http://localhost:8000
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

# Run the service
go run .
```

## 📡 API Endpoints

### Pure Single Table Design - `/api/v1/profile`

These endpoints demonstrate **pure single table design** where data is normalized across multiple items:

#### Create/Update Profile

```bash
PUT /api/v1/profile
Content-Type: application/json

{
  "id": "473b82fb-8717-4e69-894c-1844a2f183bf",
  "tags": ["politics_nerd", "binge_watcher"],
  "segments": [
    {
      "type": "morning_categories",
      "categories": [
        {"id": "politics", "score": 0.95},
        {"id": "sports", "score": 0.45}
      ],
      "top_categories": ["politics", "sports"]
    }
  ]
}
```

#### Get Profile

```bash
GET /api/v1/profile/{id}
```

#### Get Specific Segment

```bash
GET /api/v1/profile/{id}/segment/{segmentType}
# Optional: ?createdAt=2025-06-26T12:00:00Z for specific version
```

#### Get Categories from Segment

```bash
GET /api/v1/profile/{id}/segment/{segmentType}/categories
```

#### Get Tags

```bash
GET /api/v1/profile/{id}/segment/{segmentType}/tags
```

#### Get Top Categories

```bash
GET /api/v1/profile/{id}/segment/{segmentType}/topcategories
```

### Blob Storage Design - `/api/v1/blob`

These endpoints demonstrate **single table design with blob storage** where complete JSON is stored as DynamoDB maps:

#### Store Profile as Blob

```bash
PUT /api/v1/blob
Content-Type: application/json

# Same JSON structure as profile, but stored as a single DynamoDB map item
```

#### Get Profile Blob

```bash
GET /api/v1/blob/{id}
```

#### Get Segments from Blob

```bash
GET /api/v1/blob/{id}/segments
```

## 🔍 Single Table Design Patterns Demonstrated

### 1. Pure Single Table Design (`/profile` endpoints)

**Pattern**: Normalize data across multiple items for optimal access patterns

**Key Structure**:

- `PK=USER#{id}, SK=USER#{id}` → User metadata
- `PK=USER#{id}, SK=SEG#{type}#{timestamp}` → Individual segments

**Benefits**:

- Efficient queries for specific segments
- Supports segment versioning via timestamps
- Optimal for analytics and time-series data
- Granular access control

**Use Cases**:

- When you need to query individual segments frequently
- Time-series segment data
- Complex access patterns requiring joins

### 2. Blob Storage Pattern (`/blob` endpoints)

**Pattern**: Store complete JSON as native DynamoDB maps

**Key Structure**:

- `PK=USER#{id}, SK=BLOB#{id}` → Complete profile as map

**Benefits**:

- Simple storage and retrieval
- Native JSON support
- Reduced complexity for simple use cases
- Better for document-style access

**Use Cases**:

- Simple document storage
- When you typically need the complete profile
- Caching scenarios
- Simple CRUD operations

## 🛠️ Implementation Details

### Key Design Decisions

1. **Composite Keys**: Uses `#` as separator for readable, hierarchical keys
2. **TTL Support**: Automatic expiration for data lifecycle management
3. **Timestamp Versioning**: Enables time-based queries and segment history
4. **Native Map Storage**: Leverages DynamoDB's native JSON support for blob endpoints

### Repository Pattern

The service implements a clean repository pattern:

```go
type ProfilesRepo interface {
    // Pure Single Table Design methods
    UpsertProfile(ctx context.Context, profile model.Profile) error
    GetProfileByID(ctx context.Context, id string) (*model.Profile, error)
    GetSegment(ctx context.Context, profileID string, segmentType string, createdAt time.Time) (*model.Segment, error)
    GetCategories(ctx context.Context, profileID string, segmentType string) ([]model.Category, error)
    GetUserTags(ctx context.Context, profileID string) ([]string, error)
    GetTopCategories(ctx context.Context, profileID string, segmentType string) ([]string, error)
   
    // Blob Storage methods
    UpsertBlob(ctx context.Context, profileID string, data []byte) error
    GetBlob(ctx context.Context, profileID string) ([]byte, error)
    GetRawSegmentsFromBlob(ctx context.Context, profileID string) ([]byte, error)
}
```

## 📊 Example Data

See `test_profile.json` for a complete example profile structure.

## 📝 Usage Examples

This section demonstrates both design patterns side by side using the same test data.

### Single Table Design (Pure)

Store and retrieve data using normalized single table design:

```bash
# Store profile (creates multiple items: USER + SEG items)
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Content-Type: application/json" \
  -d @test_profile.json

# Retrieve complete profile (aggregates USER + SEG items)
curl http://localhost:8080/api/v1/profile/473b82fb-8717-4e69-894c-1844a2f183bf
```

### Single Table Design with Blob Storage

Store and retrieve data using blob storage with native DynamoDB maps:

```bash
# Store profile as blob (creates single BLOB item with JSON as map)
curl -X PUT http://localhost:8080/api/v1/blob \
  -H "Content-Type: application/json" \
  -d @test_profile.json

# Access segments field directly from the JSON map
curl http://localhost:8080/api/v1/blob/473b82fb-8717-4e69-894c-1844a2f183bf/segments
```

**Key Difference**: The blob approach stores the entire JSON as a native DynamoDB map, allowing direct field access (like `/segments`) without complex queries.

## 🧪 Additional Testing Examples

### Test Pure Single Table Design Features

```bash
# Get a specific segment with timestamp
curl http://localhost:8080/api/v1/profile/473b82fb-8717-4e69-894c-1844a2f183bf/segment/morning_categories

# Get categories from a specific segment
curl http://localhost:8080/api/v1/profile/473b82fb-8717-4e69-894c-1844a2f183bf/segment/morning_categories/categories

# Get top categories
curl http://localhost:8080/api/v1/profile/473b82fb-8717-4e69-894c-1844a2f183bf/segment/morning_categories/topcategories
```

### Test Blob Storage Features

```bash
# Get complete blob profile
curl http://localhost:8080/api/v1/blob/473b82fb-8717-4e69-894c-1844a2f183bf

# Access segments from blob (demonstrates map field access)
curl http://localhost:8080/api/v1/blob/473b82fb-8717-4e69-894c-1844a2f183bf/segments
```

## 📈 Performance Considerations

### Pure Single Table Design

- **Pros**: Optimized for specific access patterns, supports complex queries
- **Cons**: Multiple items per profile, more complex aggregation

### Blob Storage Design

- **Pros**: Single item per profile, simple retrieval, better for caching
- **Cons**: Less flexible for partial updates, entire document must be retrieved

## 🔧 Configuration

The service is configured via environment variables:

- `DYNAMO_ENDPOINT`: DynamoDB endpoint (default: AWS DynamoDB)
- `AWS_REGION`: AWS region
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `TABLE_NAME`: DynamoDB table name (default: user_profiles)
- `PORT`: Server port (default: :8080)


## 🧪 Testing

The service includes a comprehensive test suite built with [`dockertest`](https://github.com/ory/dockertest) that provides full integration testing with real DynamoDB containers.

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run only the main test suite
go test -v ./... -run TestSuite

# Run tests with coverage
go test -v -cover ./...
```

### Test Architecture

The test suite uses **dockertest** to provide:

- **Real DynamoDB Container**: Spins up `amazon/dynamodb-local` automatically
- **Table Initialization**: Creates the `user_profiles` table matching production schema
- **Server Integration**: Starts the actual server using the `newServer` function
- **Automatic Cleanup**: Gracefully shuts down containers and servers after tests

#### Test Infrastructure

```go
type Suite struct {
    suite.Suite
    pool         *dockertest.Pool      // Docker container pool
    dynamoRes    *dockertest.Resource  // DynamoDB container
    dynamoClient *dynamodb.Client      // DynamoDB client for setup
    baseURL      string               // API base URL for tests
    server       *server              // Server instance
    httpServer   *http.Server         // HTTP server
}
```

#### Setup Process

1. **Docker Container**: Starts DynamoDB Local container with random port
2. **Table Creation**: Creates `user_profiles` table with proper key schema
3. **Repository Setup**: Initializes DynamoDB repository with `ddb.NewDB(table)`
4. **Server Creation**: Uses `newServer(repo, log)` function like production
5. **HTTP Server**: Starts server on random available port to avoid conflicts

### Test Coverage

#### TestProfile - Pure Single Table Design

Tests the `/api/v1/profile` endpoints that demonstrate normalized single table design:

- ✅ **CreateProfile**: Profile creation and validation
- ✅ **GetProfile**: Profile retrieval and data integrity
- ✅ **VerifyProfileData**: Comprehensive segment data validation
- ✅ **GetTags**: Tag extraction and verification
- ✅ **GetNonExistentProfile**: 404 error handling

#### TestBlob - Blob Storage Design

Tests the `/api/v1/blob` endpoints that demonstrate blob storage with DynamoDB maps:

- ✅ **CreateBlob**: Blob storage as DynamoDB map
- ✅ **GetBlob**: Blob retrieval and JSON decoding
- ✅ **GetSegmentsFromBlob**: Direct field access from stored map

### Test Data

Tests use dynamically generated profiles with realistic data:

```go
testProfile := model.Profile{
    ID:   uuid.New(),
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
        // Additional segments...
    },
}
```

### Key Testing Features

#### Isolated Environment

- Each test run gets fresh DynamoDB container
- Random port allocation prevents conflicts
- Complete environment teardown after tests

#### Real Integration

- Tests actual HTTP endpoints, not mocked interfaces
- Real DynamoDB operations with actual AWS SDK
- Same server initialization as production

#### Comprehensive Coverage

- Tests both single table design patterns
- Validates data integrity across storage approaches
- Error condition testing (404s, invalid data)

#### Performance

- Fast test execution (~3-4 seconds total)
- Parallel test execution where possible
- Efficient container reuse within test suite

### Example Test Run

```bash
$ go test -v ./... -run TestSuite

=== RUN   TestSuite
=== RUN   TestSuite/TestBlob
=== RUN   TestSuite/TestBlob/CreateBlob
=== RUN   TestSuite/TestBlob/GetBlob
=== RUN   TestSuite/TestBlob/GetSegmentsFromBlob
=== RUN   TestSuite/TestProfile
=== RUN   TestSuite/TestProfile/CreateProfile
=== RUN   TestSuite/TestProfile/GetProfile
=== RUN   TestSuite/TestProfile/VerifyProfileData
=== RUN   TestSuite/TestProfile/GetTags
=== RUN   TestSuite/TestProfile/GetNonExistentProfile
--- PASS: TestSuite (3.69s)
    --- PASS: TestSuite/TestBlob (0.07s)
    --- PASS: TestSuite/TestProfile (0.02s)
PASS
```

### Testing Prerequisites

- **Docker**: Required for DynamoDB container
- **Go 1.24+**: For running tests
- **Network Access**: To pull `amazon/dynamodb-local` image

### Testing Best Practices

The test suite demonstrates several testing best practices:

- **Infrastructure as Code**: Complete test environment in code
- **Isolation**: Each test run is completely isolated
- **Realistic Testing**: Uses real AWS services, not mocks
- **Comprehensive Coverage**: Tests all major API endpoints
- **Error Testing**: Validates error conditions and edge cases
- **Clean Assertions**: Clear, descriptive test assertions
- **Parallel Execution**: Tests run efficiently in parallel where possible

## 🏭 Production Considerations

This is a **PoC** - for production use, consider:

- Proper AWS IAM roles instead of static credentials
- VPC and security group configurations  
- Monitoring and observability
- Error handling and retry logic
- API rate limiting and authentication
- Data validation and sanitization
- Backup and disaster recovery strategies

## 📚 Learning Resources

- [DynamoDB Single Table Design](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-general-nosql-design.html)
- [Single Table Design Patterns](https://www.alexdebrie.com/posts/dynamodb-single-table/)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [AWS re:Invent 2018: Amazon DynamoDB Deep Dive: Advanced Design Patterns for DynamoDB (DAT401)](https://www.youtube.com/watch?v=HaEPXoXVf2k)
- [Reserved Words in DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ReservedWords.html)
