package handler

import (
	"fmt"
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCrudFlags(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	// step 0. it should get 0 flags when db is empty
	res = c.FindFlags(flag.FindFlagsParams{})
	assert.Len(t, res.(*flag.FindFlagsOK).Payload, 0)

	// step 1. it should be able to create one flag
	res = c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
			Key:         "some_random_flag_key",
		},
	})
	assert.NotZero(t, res.(*flag.CreateFlagOK).Payload.ID)
	assert.Equal(t, "some_random_flag_key", res.(*flag.CreateFlagOK).Payload.Key)

	// step 2. it should be able to find some flags after creation
	res = c.FindFlags(flag.FindFlagsParams{})
	assert.NotZero(t, len(res.(*flag.FindFlagsOK).Payload))

	// step 3. it should be able to get the flag after creation
	res = c.GetFlag(flag.GetFlagParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.GetFlagOK).Payload.ID)
	assert.NotZero(t, res.(*flag.GetFlagOK).Payload.Key)

	// step 4. it should be able to put the flag
	res = c.PutFlag(flag.PutFlagParams{
		FlagID: int64(1),
		Body: &models.PutFlagRequest{
			Description:        util.StringPtr("another funny flag"),
			DataRecordsEnabled: util.BoolPtr(true),
			Key:                util.StringPtr("flag_key_1"),
		}},
	)
	assert.NotZero(t, res.(*flag.PutFlagOK).Payload.ID)
	assert.Equal(t, "flag_key_1", res.(*flag.PutFlagOK).Payload.Key)

	// step 5. it should be able to set the flag enabled state
	res = c.SetFlagEnabledState(flag.SetFlagEnabledParams{
		FlagID: int64(1),
		Body: &models.SetFlagEnabledRequest{
			Enabled: util.BoolPtr(true),
		}},
	)
	assert.True(t, *res.(*flag.SetFlagEnabledOK).Payload.Enabled)

	// step 6. it should be able to get the flag snapshot
	res = c.GetFlagSnapshots(flag.GetFlagSnapshotsParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.GetFlagSnapshotsOK).Payload)

	// step 7. it should be able to delete the flag
	res = c.DeleteFlag(flag.DeleteFlagParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.DeleteFlagOK))
}

func TestCrudFlagsWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("GetFlag - can't find non-exist flag", func(t *testing.T) {
		res = c.GetFlag(flag.GetFlagParams{FlagID: int64(1)})
		assert.NotZero(t, res.(*flag.GetFlagDefault).Payload)
	})

	t.Run("GetFlag - got e2r MapFlag error", func(t *testing.T) {
		c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
				Key:         "flag_key_1",
			},
		})
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.GetFlag(flag.GetFlagParams{FlagID: int64(1)})
		assert.NotZero(t, res.(*flag.GetFlagDefault).Payload)
	})

	t.Run("FindFlags - got e2r MapFlags error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlags, nil, fmt.Errorf("e2r MapFlags error")).Reset()
		res = c.FindFlags(flag.FindFlagsParams{})
		assert.NotZero(t, res.(*flag.FindFlagsDefault).Payload)
	})

	t.Run("FindFlags - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.FindFlags(flag.FindFlagsParams{})
		assert.NotZero(t, res.(*flag.FindFlagsDefault).Payload)
		db.Error = nil
	})

	t.Run("CreateFlag - got e2r MapFlag error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("CreateFlag - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
		db.Error = nil
	})

	t.Run("CreateFlag - invalid key error", func(t *testing.T) {
		res = c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr("funny flag"),
				Key:         "1-2-3", // invalid key
			},
		})
		assert.NotZero(t, res.(*flag.CreateFlagDefault).Payload)
	})

	t.Run("PutFlag - try to update a non-existing flag", func(t *testing.T) {
		res = c.PutFlag(flag.PutFlagParams{
			FlagID: int64(99999),
			Body: &models.PutFlagRequest{
				Description:        util.StringPtr("another funny flag"),
				DataRecordsEnabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.PutFlagDefault).Payload)
	})

	t.Run("PutFlag - got e2r MapFlag error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.PutFlag(flag.PutFlagParams{
			FlagID: int64(1),
			Body: &models.PutFlagRequest{
				Description:        util.StringPtr("another funny flag"),
				DataRecordsEnabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.PutFlagDefault).Payload)
	})

	t.Run("PutFlag - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.PutFlag(flag.PutFlagParams{
			FlagID: int64(1),
			Body: &models.PutFlagRequest{
				Description:        util.StringPtr("another funny flag"),
				DataRecordsEnabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.PutFlagDefault).Payload)
		db.Error = nil
	})

	t.Run("PutFlag - cannot set duplicate flag_key", func(t *testing.T) {
		res = c.PutFlag(flag.PutFlagParams{
			FlagID: int64(2),
			Body: &models.PutFlagRequest{
				Description:        util.StringPtr("another funny flag"),
				DataRecordsEnabled: util.BoolPtr(true),
				Key:                util.StringPtr("flag_key_1"),
			}},
		)
		assert.NotZero(t, res.(*flag.PutFlagDefault).Payload)
	})

	t.Run("SetFlagEnabledState - try to set on a non-existing flag", func(t *testing.T) {
		res = c.SetFlagEnabledState(flag.SetFlagEnabledParams{
			FlagID: int64(99999),
			Body: &models.SetFlagEnabledRequest{
				Enabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.SetFlagEnabledDefault).Payload)
	})

	t.Run("SetFlagEnabledState - got e2r error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlag, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.SetFlagEnabledState(flag.SetFlagEnabledParams{
			FlagID: int64(1),
			Body: &models.SetFlagEnabledRequest{
				Enabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.SetFlagEnabledDefault).Payload)
	})

	t.Run("SetFlagEnabledState - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.SetFlagEnabledState(flag.SetFlagEnabledParams{
			FlagID: int64(1),
			Body: &models.SetFlagEnabledRequest{
				Enabled: util.BoolPtr(true),
			}},
		)
		assert.NotZero(t, res.(*flag.SetFlagEnabledDefault).Payload)
		db.Error = nil
	})

	t.Run("DeleteFlag - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.DeleteFlag(flag.DeleteFlagParams{FlagID: int64(99999)})
		assert.NotZero(t, res.(*flag.DeleteFlagDefault).Payload)
		db.Error = nil
	})

	t.Run("GetFlagSnapshots - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.GetFlagSnapshots(flag.GetFlagSnapshotsParams{FlagID: int64(99999)})
		assert.NotZero(t, res.(*flag.GetFlagSnapshotsDefault).Payload)
		db.Error = nil
	})

	t.Run("GetFlagSnapshots - e2r MapFlagSnapshots error", func(t *testing.T) {
		defer gostub.StubFunc(&e2rMapFlagSnapshots, nil, fmt.Errorf("e2r MapFlag error")).Reset()
		res = c.GetFlagSnapshots(flag.GetFlagSnapshotsParams{FlagID: int64(99999)})
		assert.NotZero(t, res.(*flag.GetFlagSnapshotsDefault).Payload)
	})
}

func TestFindFlags(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}
	numOfFlags := 20

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	for i := 0; i < numOfFlags; i++ {
		c.CreateFlag(flag.CreateFlagParams{
			Body: &models.CreateFlagRequest{
				Description: util.StringPtr(fmt.Sprintf("flag_%d", i)),
				Key:         fmt.Sprintf("flag_key_%d", i),
			},
		})
	}

	t.Run("FindFlags - got all the results", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, numOfFlags)
	})

	t.Run("FindFlags (with enabled only) - got all the enabled results", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			Enabled: util.BoolPtr(true),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 0)
	})
	t.Run("FindFlags (with matching description)", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			Description: util.StringPtr("flag_1"),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 1)
	})
	t.Run("FindFlags (with matching key)", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			Key: util.StringPtr("flag_key_1"),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 1)
	})
	t.Run("FindFlags (with matching description_like)", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			DescriptionLike: util.StringPtr("flag_"),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, numOfFlags)
	})
	t.Run("FindFlags (with limit)", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			Limit: util.Int64Ptr(int64(2)),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 2)
	})
	t.Run("FindFlags (with limit and offset)", func(t *testing.T) {
		res = c.FindFlags(flag.FindFlagsParams{
			Limit:  util.Int64Ptr(int64(2)),
			Offset: util.Int64Ptr(int64(2)),
		})
		assert.Len(t, res.(*flag.FindFlagsOK).Payload, 2)
		assert.Equal(t, res.(*flag.FindFlagsOK).Payload[0].ID, int64(3))
		assert.Equal(t, res.(*flag.FindFlagsOK).Payload[1].ID, int64(4))
	})
}

func TestCrudSegments(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})

	// step 1. it should be able to create segment
	res = c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	assert.NotZero(t, res.(*segment.CreateSegmentOK).Payload)
	res = c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment2"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	assert.NotZero(t, res.(*segment.CreateSegmentOK).Payload)

	// step 2. it should be able to find the segments
	res = c.FindSegments(segment.FindSegmentsParams{FlagID: int64(1)})
	assert.NotZero(t, len(res.(*segment.FindSegmentsOK).Payload))

	// step 3. it should be able to put the segment
	res = c.PutSegment(segment.PutSegmentParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(0)),
		},
	})
	assert.NotZero(t, res.(*segment.PutSegmentOK).Payload.ID)

	// step 4. it should be able to reorder the segments
	res = c.PutSegmentsReorder(segment.PutSegmentsReorderParams{
		FlagID: int64(1),
		Body: &models.PutSegmentReorderRequest{
			SegmentIds: []int64{int64(2), int64(1)},
		},
	})
	assert.NotZero(t, res.(*segment.PutSegmentsReorderOK))

	// step 5. it should have the correct order of segments
	res = c.FindSegments(segment.FindSegmentsParams{FlagID: int64(1)})
	assert.Equal(t, res.(*segment.FindSegmentsOK).Payload[0].ID, int64(2))

	// step 6. it should be able to delete the segment
	res = c.DeleteSegment(segment.DeleteSegmentParams{
		FlagID:    int64(1),
		SegmentID: int64(2),
	})
	assert.NotZero(t, res.(*segment.DeleteSegmentOK))
}

func TestCrudSegmentsWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})

	t.Run("FindSegments - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.FindSegments(segment.FindSegmentsParams{FlagID: int64(1)})
		assert.NotZero(t, res.(*segment.FindSegmentsDefault).Payload)
		db.Error = nil
	})

	t.Run("CreateSegments - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.CreateSegment(segment.CreateSegmentParams{
			FlagID: int64(1),
			Body: &models.CreateSegmentRequest{
				Description:    util.StringPtr("segment1"),
				RolloutPercent: util.Int64Ptr(int64(100)),
			},
		})
		assert.NotZero(t, res.(*segment.CreateSegmentDefault).Payload)
		db.Error = nil
	})

	t.Run("PutSegments - put on a non-existing segment", func(t *testing.T) {
		res = c.PutSegment(segment.PutSegmentParams{
			FlagID:    int64(1),
			SegmentID: int64(999999),
			Body: &models.PutSegmentRequest{
				Description:    util.StringPtr("segment1"),
				RolloutPercent: util.Int64Ptr(int64(0)),
			},
		})
		assert.NotZero(t, res.(*segment.PutSegmentDefault).Payload)
	})

	t.Run("PutSegmentsReorder - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.PutSegmentsReorder(segment.PutSegmentsReorderParams{
			FlagID: int64(1),
			Body: &models.PutSegmentReorderRequest{
				SegmentIds: []int64{int64(999998), int64(1)},
			},
		})
		assert.NotZero(t, res.(*segment.PutSegmentsReorderDefault).Payload)
		db.Error = nil
	})

	t.Run("DeleteSegment - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.DeleteSegment(segment.DeleteSegmentParams{
			FlagID:    int64(1),
			SegmentID: int64(2),
		})
		assert.NotZero(t, res.(*segment.DeleteSegmentDefault).Payload)
		db.Error = nil
	})
}

func TestCrudConstraints(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})

	// step 1. it should return 0 constraints before creaetion
	res = c.FindConstraints(constraint.FindConstraintsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.Zero(t, len(res.(*constraint.FindConstraintsOK).Payload))

	// step 2. it should be able to create a constraint
	res = c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Operator: util.StringPtr("EQ"),
			Property: util.StringPtr("state"),
			Value:    util.StringPtr(`"NY"`),
		},
	})
	assert.NotZero(t, res.(*constraint.CreateConstraintOK).Payload.ID)

	// step 3. it should return some constraints when we get
	res = c.FindConstraints(constraint.FindConstraintsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.NotZero(t, len(res.(*constraint.FindConstraintsOK).Payload))

	// step 4. it should be able to put the constraint
	res = c.PutConstraint(constraint.PutConstraintParams{
		FlagID:       int64(1),
		SegmentID:    int64(1),
		ConstraintID: int64(1),
		Body: &models.CreateConstraintRequest{
			Operator: util.StringPtr("EQ"),
			Property: util.StringPtr("state"),
			Value:    util.StringPtr(`"CA"`),
		},
	})
	assert.NotZero(t, res.(*constraint.PutConstraintOK).Payload.ID)

	// step 5. it should be able to delete a constraint
	res = c.DeleteConstraint(constraint.DeleteConstraintParams{
		FlagID:       int64(1),
		SegmentID:    int64(1),
		ConstraintID: int64(1),
	})
	assert.NotZero(t, res.(*constraint.DeleteConstraintOK))
}

func TestCrudConstraintsFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Operator: util.StringPtr("EQ"),
			Property: util.StringPtr("state"),
			Value:    util.StringPtr(`"NY"`),
		},
	})

	t.Run("FindConstraints - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.FindConstraints(constraint.FindConstraintsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
		})
		assert.NotZero(t, res.(*constraint.FindConstraintsDefault).Payload)
		db.Error = nil
	})

	t.Run("CreateConstraints - creation validation error", func(t *testing.T) {
		res = c.CreateConstraint(constraint.CreateConstraintParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.CreateConstraintRequest{
				Operator: util.StringPtr("IN"),
				Property: util.StringPtr("state"),
				Value:    util.StringPtr(`"NY"]`), // invalid array []
			},
		})
		assert.NotZero(t, res.(*constraint.CreateConstraintDefault).Payload)
	})

	t.Run("CreateConstraint - generic db error", func(t *testing.T) {
		db.Error = fmt.Errorf("generic db error")
		res = c.CreateConstraint(constraint.CreateConstraintParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.CreateConstraintRequest{
				Operator: util.StringPtr("EQ"),
				Property: util.StringPtr("state"),
				Value:    util.StringPtr(`"NY"`), // invalid array []
			},
		})
		assert.NotZero(t, res.(*constraint.CreateConstraintDefault).Payload)
		db.Error = nil
	})

	t.Run("PutConstraint - put on a non-existing constraint", func(t *testing.T) {
		res = c.PutConstraint(constraint.PutConstraintParams{
			FlagID:       int64(1),
			SegmentID:    int64(1),
			ConstraintID: int64(999999),
			Body: &models.CreateConstraintRequest{
				Operator: util.StringPtr("EQ"),
				Property: util.StringPtr("state"),
				Value:    util.StringPtr(`"CA"`),
			},
		})
		assert.NotZero(t, res.(*constraint.PutConstraintDefault).Payload)
	})

	t.Run("PutConstraint - put validation error", func(t *testing.T) {
		res = c.PutConstraint(constraint.PutConstraintParams{
			FlagID:       int64(1),
			SegmentID:    int64(1),
			ConstraintID: int64(1),
			Body: &models.CreateConstraintRequest{
				Operator: util.StringPtr("IN"),
				Property: util.StringPtr("state"),
				Value:    util.StringPtr(`"CA"]`),
			},
		})
		assert.NotZero(t, res.(*constraint.PutConstraintDefault).Payload)
	})

	t.Run("DeleteConstraint - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("generic db error")
		res = c.DeleteConstraint(constraint.DeleteConstraintParams{
			FlagID:       int64(1),
			SegmentID:    int64(1),
			ConstraintID: int64(1),
		})
		assert.NotZero(t, res.(*constraint.DeleteConstraintDefault))
		db.Error = nil
	})
}

func TestCrudVariants(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})

	// step 0. it should return 0 variants before creaetion
	res = c.FindVariants(variant.FindVariantsParams{
		FlagID: int64(1),
	})
	assert.Zero(t, len(res.(*variant.FindVariantsOK).Payload))

	// step 1. it should be able to create variant
	res = c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})
	assert.NotZero(t, res.(*variant.CreateVariantOK).Payload.ID)

	// step 2. it should return some variants after creaetion
	res = c.FindVariants(variant.FindVariantsParams{
		FlagID: int64(1),
	})
	assert.NotZero(t, len(res.(*variant.FindVariantsOK).Payload))

	// step 3. it should be able to put variant
	res = c.PutVariant(variant.PutVariantParams{
		FlagID:    int64(1),
		VariantID: int64(1),
		Body: &models.PutVariantRequest{
			Key: util.StringPtr("another_control"),
			Attachment: map[string]interface{}{
				"valid_string_value": "1",
			},
		},
	})
	assert.Equal(t, *res.(*variant.PutVariantOK).Payload.Key, "another_control")

	// step 4. it should be able to delete the variant
	res = c.DeleteVariant(variant.DeleteVariantParams{
		FlagID:    int64(1),
		VariantID: int64(1),
	})
	assert.NotZero(t, res.(*variant.DeleteVariantOK))
}

func TestCrudVariantsWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})

	t.Run("CreateVariant - r2e MapAttachment error", func(t *testing.T) {
		defer gostub.StubFunc(&r2eMapAttachment, nil, fmt.Errorf("r2e MapAttachment error")).Reset()
		res = c.CreateVariant(variant.CreateVariantParams{
			FlagID: int64(1),
			Body: &models.CreateVariantRequest{
				Key: util.StringPtr("control"),
			},
		})
		assert.NotZero(t, res.(*variant.CreateVariantDefault).Payload)
	})

	t.Run("CreateVariant - creation validation error", func(t *testing.T) {
		res = c.CreateVariant(variant.CreateVariantParams{
			FlagID: int64(1),
			Body: &models.CreateVariantRequest{
				Key: util.StringPtr("123_invalid_key"),
			},
		})
		assert.NotZero(t, res.(*variant.CreateVariantDefault).Payload)
	})

	t.Run("CreateVariant - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.CreateVariant(variant.CreateVariantParams{
			FlagID: int64(1),
			Body: &models.CreateVariantRequest{
				Key: util.StringPtr("key"),
			},
		})
		assert.NotZero(t, res.(*variant.CreateVariantDefault).Payload)
		db.Error = nil
	})

	t.Run("FindVariants - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.FindVariants(variant.FindVariantsParams{
			FlagID: int64(1),
		})
		assert.NotZero(t, res.(*variant.FindVariantsDefault).Payload)
		db.Error = nil
	})

	t.Run("PutVariant - put on a non-existing variant", func(t *testing.T) {
		res = c.PutVariant(variant.PutVariantParams{
			FlagID:    int64(1),
			VariantID: int64(999999),
			Body: &models.PutVariantRequest{
				Key: util.StringPtr("another_control"),
			},
		})
		assert.NotZero(t, *res.(*variant.PutVariantDefault).Payload)
	})

	t.Run("PutVariant - put invalid attachment", func(t *testing.T) {
		res = c.PutVariant(variant.PutVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
			Body: &models.PutVariantRequest{
				Key: util.StringPtr("another_control"),
				Attachment: map[string]interface{}{
					"invalid_int_value": 1,
				},
			},
		})
		assert.NotZero(t, *res.(*variant.PutVariantDefault).Payload)
	})

	t.Run("PutVariant - put validation error", func(t *testing.T) {
		res = c.PutVariant(variant.PutVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
			Body: &models.PutVariantRequest{
				Key: util.StringPtr("123_invalid_key"),
			},
		})
		assert.NotZero(t, *res.(*variant.PutVariantDefault).Payload)
	})

	t.Run("PutVariant - validatePutVariantForDistributions error", func(t *testing.T) {
		defer gostub.StubFunc(&validatePutVariantForDistributions, NewError(500, "validatePutVariantForDistributions error")).Reset()
		res = c.PutVariant(variant.PutVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
			Body: &models.PutVariantRequest{
				Key: util.StringPtr("key"),
			},
		})
		assert.NotZero(t, *res.(*variant.PutVariantDefault).Payload)
	})

	t.Run("DeleteVariant - validateDeleteVariant error", func(t *testing.T) {
		defer gostub.StubFunc(&validateDeleteVariant, NewError(500, "validateDeleteVariant error")).Reset()
		res = c.DeleteVariant(variant.DeleteVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
		})
		assert.NotZero(t, res.(*variant.DeleteVariantDefault).Payload)
	})

	t.Run("DeleteVariant - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.DeleteVariant(variant.DeleteVariantParams{
			FlagID:    int64(1),
			VariantID: int64(1),
		})
		assert.NotZero(t, res.(*variant.DeleteVariantDefault).Payload)
		db.Error = nil
	})
}

func TestCrudDistributions(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})

	// step 0. it should return 0 distributions before the creation
	res = c.FindDistributions(distribution.FindDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.Zero(t, len(res.(*distribution.FindDistributionsOK).Payload))

	// step 1. it should be able to create distribution
	res = c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{
				{
					Percent:    util.Int64Ptr(int64(100)),
					VariantID:  util.Int64Ptr(int64(1)),
					VariantKey: util.StringPtr("control"),
				},
			},
		},
	})
	assert.NotZero(t, res.(*distribution.PutDistributionsOK).Payload)

	// step 2. it should return some distributions before the creation
	res = c.FindDistributions(distribution.FindDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.NotZero(t, len(res.(*distribution.FindDistributionsOK).Payload))
}

func TestCrudDistributionsWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})
	c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{
				{
					Percent:    util.Int64Ptr(int64(100)),
					VariantID:  util.Int64Ptr(int64(1)),
					VariantKey: util.StringPtr("control"),
				},
			},
		},
	})

	t.Run("FindDistributions - db generic error", func(t *testing.T) {
		db.Error = fmt.Errorf("db generic error")
		res = c.FindDistributions(distribution.FindDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
		})
		assert.NotZero(t, res.(*distribution.FindDistributionsDefault).Payload)
		db.Error = nil
	})

	t.Run("PutDistributions - validatePutDistributions error", func(t *testing.T) {
		res = c.PutDistributions(distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(50)), // not adds up to 100
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		})
		assert.NotZero(t, res.(*distribution.PutDistributionsDefault).Payload)
	})

	t.Run("PutDistributions - cannot delete previous distribution error", func(t *testing.T) {
		defer gostub.StubFunc(&validatePutDistributions, nil).Reset()
		db.Error = fmt.Errorf("cannot delete previous distribution")
		res = c.PutDistributions(distribution.PutDistributionsParams{
			FlagID:    int64(1),
			SegmentID: int64(1),
			Body: &models.PutDistributionsRequest{
				Distributions: []*models.Distribution{
					{
						Percent:    util.Int64Ptr(int64(100)),
						VariantID:  util.Int64Ptr(int64(1)),
						VariantKey: util.StringPtr("control"),
					},
				},
			},
		})
		assert.NotZero(t, res.(*distribution.PutDistributionsDefault).Payload)
		db.Error = nil
	})
}
