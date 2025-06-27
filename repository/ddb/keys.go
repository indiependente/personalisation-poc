package ddb

import (
	"fmt"
	"time"
)

const (
	// DynamoDB field names are counted towards the RCU and WCU.
	// This is why it's a good practice to use short field names https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/CheatSheet.html
	// Also, be aware of reserved words https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ReservedWords.html
	partitionKey = "pk"
	sortKey      = "sk"

	keySeparator           = "#"
	sortKeyTimestampLayout = time.RFC3339
	itemType               = "typ"
)

func buildPK(id string) string {
	return fmt.Sprintf("%s%s%s", userItemKeyPrefix, keySeparator, id)
}

func buildSK(itemPrefix, itemType string, timestamp *time.Time) string {
	if timestamp != nil {
		return fmt.Sprintf("%s%s%s%s%s", itemPrefix, keySeparator, itemType, keySeparator, timestamp.Format(sortKeyTimestampLayout))
	}

	return fmt.Sprintf("%s%s%s", itemPrefix, keySeparator, itemType)
}

// These functions can be used to parse the PK and SK from the DynamoDB item
// func parsePK(pk string) string {
// 	return strings.Split(pk, keySeparator)[1]
// }

// func parseSortKey(key string) (string, time.Time, error) {
// 	split := strings.Split(key, keySeparator)
// 	if len(split) != 2 {
// 		return "", time.Time{}, errors.New("missing sort key fields")
// 	}
// 	item := split[0]
// 	tstamp, err := time.Parse(sortKeyTimestampLayout, split[1])
// 	if err != nil {
// 		return "", time.Time{}, err
// 	}

// 	return item, tstamp, nil
// }

// daysAgoSK creates a SK for an item by attaching a timestamp in RFC3339 format.
// The timestamp is the current time minus the number of days specified.
// func daysAgoSK(prefix string, days int) string {
// 	return prefix + keySeparator + time.Now().Add(-time.Duration(days)*time.Hour*24).Format("2006-01-02")
// }
