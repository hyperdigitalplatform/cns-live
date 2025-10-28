package domain

import "time"

// MilestoneRecordingMetadata represents metadata about Milestone recordings
type MilestoneRecordingMetadata struct {
	Available    bool
	SegmentCount int
	TotalSize    int64
	StartTime    time.Time
	EndTime      time.Time
}
