package ch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// ActiveWindow represents a time window with sufficient signal activity.
// Used by frequency and changepoint detectors.
type ActiveWindow struct {
	WindowStart         time.Time
	WindowEnd           time.Time
	SignalCount         uint64
	DistinctSignalCount uint64
}

// getWindowedSignalCounts queries per-window signal counts from ClickHouse.
// Shared by FrequencyDetector and ChangePointDetector.
//
// Performance notes:
//   - PREWHERE filters on primary key (token_id) before FINAL merge
//   - Pre-allocates result slice based on expected window count
func getWindowedSignalCounts(
	ctx context.Context,
	conn clickhouse.Conn,
	tokenID uint32,
	from, to time.Time,
	windowSizeSeconds int,
	signalThreshold int,
	distinctSignalThreshold int,
) (_ []ActiveWindow, retErr error) {
	query := `
SELECT
    toStartOfInterval(timestamp, INTERVAL ? second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL ? second) + INTERVAL ? second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
PREWHERE token_id = ?
WHERE timestamp >= ?
  AND timestamp < ?
GROUP BY window_start
HAVING signal_count >= ? AND distinct_signal_count >= ?
ORDER BY window_start`

	rows, err := conn.Query(ctx, query, windowSizeSeconds, windowSizeSeconds, windowSizeSeconds, tokenID, from, to, signalThreshold, distinctSignalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query windowed signal counts: %w", err)
	}
	defer func() { retErr = errors.Join(retErr, rows.Close()) }()

	expectedWindows := int(to.Sub(from).Seconds()) / windowSizeSeconds
	windows := make([]ActiveWindow, 0, expectedWindows)
	for rows.Next() {
		var w ActiveWindow
		if err := rows.Scan(&w.WindowStart, &w.WindowEnd, &w.SignalCount, &w.DistinctSignalCount); err != nil {
			return nil, fmt.Errorf("failed to scan windowed signal count row: %w", err)
		}
		windows = append(windows, w)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate windowed signal count rows: %w", rows.Err())
	}
	return windows, nil
}

// levelSample is a timestamped numeric sample (fuel, SoC, RPM, etc.).
type levelSample struct {
	ts    time.Time
	value float64
}

// timeRange is a lightweight start/end pair used internally by detection and merge pipelines.
// Converted to *model.Segment only at the final return to avoid intermediate heap allocations.
type timeRange struct {
	start time.Time
	end   time.Time
}

// sampleAtOrBefore returns the value of the latest sample at or before t.
// Returns 0 if no sample exists at or before t (i.e. all samples are in the future).
// samples must be sorted by ts.
func sampleAtOrBefore(samples []levelSample, t time.Time) float64 {
	if len(samples) == 0 {
		return 0
	}
	idx := sort.Search(len(samples), func(i int) bool { return samples[i].ts.After(t) })
	if idx == 0 {
		// All samples are after t; no valid sample exists at or before t.
		return 0
	}
	return samples[idx-1].value
}

// levelFirstLastInRange returns the first and last level value within [segStart, segEnd].
// samples must be sorted by ts. ok is false if no samples fall in range.
func levelFirstLastInRange(samples []levelSample, segStart, segEnd time.Time) (first, last float64, ok bool) {
	if len(samples) == 0 {
		return 0, 0, false
	}
	startIdx := sort.Search(len(samples), func(i int) bool { return !samples[i].ts.Before(segStart) })
	if startIdx >= len(samples) || samples[startIdx].ts.After(segEnd) {
		return 0, 0, false
	}
	endIdx := sort.Search(len(samples), func(i int) bool { return samples[i].ts.After(segEnd) })
	if endIdx == 0 {
		return 0, 0, false
	}
	endIdx--
	if samples[endIdx].ts.Before(segStart) {
		return 0, 0, false
	}
	return samples[startIdx].value, samples[endIdx].value, true
}

// getLevelSamples fetches timestamped level samples for a signal.
// Results are returned in timestamp order (ORDER BY in the query).
// Uses PREWHERE on token_id for efficient primary-key filtering before FINAL merge.
func getLevelSamples(ctx context.Context, conn clickhouse.Conn, tokenID uint32, name string, from, to time.Time) (_ []levelSample, retErr error) {
	query := "SELECT " + vss.TimestampCol + ", " + vss.ValueNumberCol +
		" FROM " + vss.TableName + " FINAL" +
		" PREWHERE " + vss.TokenIDCol + " = ?" +
		" WHERE " + vss.NameCol + " = ? AND " + vss.TimestampCol + " >= ? AND " + vss.TimestampCol + " < ?" +
		" ORDER BY " + vss.TimestampCol
	rows, err := conn.Query(ctx, query, tokenID, name, from, to)
	if err != nil {
		return nil, err
	}
	defer func() { retErr = errors.Join(retErr, rows.Close()) }()
	out := make([]levelSample, 0, 1024)
	for rows.Next() {
		var s levelSample
		if err := rows.Scan(&s.ts, &s.value); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// mergeTimeRanges merges sorted time ranges within maxGap. If shouldMerge is non-nil it is called
// to decide whether two ranges within maxGap should actually merge (e.g. odometer check).
// Only ranges with duration >= minDuration are kept. Ranges are clipped to [from, to].
func mergeTimeRanges(ranges []timeRange, maxGap time.Duration, minDuration int, from, to time.Time, shouldMerge func(a, b timeRange) bool) []timeRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start.Before(ranges[j].start) })
	var out []timeRange
	cur := ranges[0]
	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		gap := next.start.Sub(cur.end)
		doMerge := gap <= maxGap
		if doMerge && shouldMerge != nil {
			doMerge = shouldMerge(cur, next)
		}
		if doMerge {
			if next.end.After(cur.end) {
				cur.end = next.end
			}
		} else {
			if tr, ok := clipTimeRange(cur, from, to, minDuration); ok {
				out = append(out, tr)
			}
			cur = next
		}
	}
	if tr, ok := clipTimeRange(cur, from, to, minDuration); ok {
		out = append(out, tr)
	}
	return out
}

// clipTimeRange clips a timeRange to [from, to] and checks minDuration. Zero from/to disables clipping.
func clipTimeRange(tr timeRange, from, to time.Time, minDuration int) (timeRange, bool) {
	if !from.IsZero() && tr.start.Before(from) {
		tr.start = from
	}
	if !to.IsZero() && tr.end.After(to) {
		tr.end = to
	}
	if !tr.end.After(tr.start) {
		return timeRange{}, false
	}
	if int(tr.end.Sub(tr.start).Seconds()) < minDuration {
		return timeRange{}, false
	}
	return tr, true
}

// timeRangesToSegments converts a slice of timeRange to []*model.Segment.
// This is the single point where heap-allocated model objects are created.
// Returns an empty (non-nil) slice when ranges is empty for consistent downstream handling.
func timeRangesToSegments(ranges []timeRange, from time.Time) []*model.Segment {
	if len(ranges) == 0 {
		return []*model.Segment{}
	}
	out := make([]*model.Segment, 0, len(ranges))
	for _, tr := range ranges {
		startedBefore := tr.start.Equal(from) || tr.start.Before(from)
		end := tr.end
		durSec := int32(end.Sub(tr.start).Seconds())
		out = append(out, newSegment(tr.start, &end, durSec, false, startedBefore))
	}
	return out
}

// timeNow is the clock function used by mergeWindowsIntoSegments/windowRunToSegment.
// Tests override this to make ongoing-detection deterministic.
var timeNow = time.Now

// mergeWindowsIntoSegments merges consecutive ActiveWindows within maxGap seconds, clips to [from, to],
// and marks segments as ongoing when the last window end is within maxGap of to.
// Used by frequency and changepoint detectors.
func mergeWindowsIntoSegments(windows []ActiveWindow, from, to time.Time, maxGap, minDuration int) []*model.Segment {
	if len(windows) == 0 {
		return []*model.Segment{}
	}

	// Convert ActiveWindows to timeRanges and delegate to the shared merge pipeline.
	ranges := make([]timeRange, len(windows))
	for i, w := range windows {
		ranges[i] = timeRange{start: w.WindowStart, end: w.WindowEnd}
	}
	merged := mergeTimeRanges(ranges, time.Duration(maxGap)*time.Second, minDuration, from, to, nil)

	// Check whether the last merged range qualifies as ongoing:
	// the run end is within maxGap of the query boundary AND the query boundary is near real-time.
	maxGapDur := time.Duration(maxGap) * time.Second
	out := make([]*model.Segment, 0, len(merged))
	for i, tr := range merged {
		isLast := i == len(merged)-1
		if isLast && to.Sub(tr.end) <= maxGapDur && timeNow().Sub(to) <= maxGapDur {
			// Ongoing segment: duration extends to 'to', end is nil.
			durSec := int32(to.Sub(tr.start).Seconds())
			if int(durSec) >= minDuration {
				out = append(out, newSegment(tr.start, nil, durSec, true, tr.start.Equal(from) || tr.start.Before(from)))
			}
		} else {
			end := tr.end
			durSec := int32(end.Sub(tr.start).Seconds())
			out = append(out, newSegment(tr.start, &end, durSec, false, tr.start.Equal(from) || tr.start.Before(from)))
		}
	}
	if len(out) == 0 {
		return []*model.Segment{}
	}
	return out
}
