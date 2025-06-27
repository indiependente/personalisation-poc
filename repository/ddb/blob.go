package ddb

import (
	"encoding/json"
	"time"
)

const (
	blobItemKeyPrefix = "BLOB"
)

type blob struct {
	PK       string `dynamo:"pk,hash"`  // partition key
	SK       string `dynamo:"sk,range"` // sort key
	ItemType string `dynamo:"typ"`      // item type
	ID       string `dynamo:"id"`
	TTL      int64  `dynamo:"ttl,unixtime"`
	Data     any    `dynamo:"rawdata"`
}

func toDBBlob(profileID string, data []byte) (blob, error) {
	// Parse JSON into interface{} so DynamoDB can store it as a native map
	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return blob{}, err
	}

	return blob{
		PK:       buildPK(profileID),
		SK:       buildSK(blobItemKeyPrefix, profileID, nil),
		ItemType: blobItemKeyPrefix,
		ID:       profileID,
		TTL:      time.Now().AddDate(1, 0, 0).Unix(), // 1 year
		Data:     jsonData,
	}, nil
}
