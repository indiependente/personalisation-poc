package ddb

import (
	"context"
	"fmt"
	"personalisation-poc/model"
)

func (d *DB) UpsertProfile(ctx context.Context, profile model.Profile) error {
	user, segments := toDBItems(profile)

	bw := d.table.Batch().Write().Put(user)
	for _, segment := range segments {
		bw.Put(segment)
	}
	_, err := bw.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to write batch: %w", err)
	}

	return nil
}

func (d *DB) UpsertBlob(ctx context.Context, profileID string, data []byte) error {
	blob, err := toDBBlob(profileID, data)
	if err != nil {
		return fmt.Errorf("failed to parse blob data: %w", err)
	}

	err = d.table.Put(blob).Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}
	return nil
}
