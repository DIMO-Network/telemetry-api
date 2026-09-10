package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// pt is a (minute offset, value) pair used to build synthetic sample series.
type pt struct {
	m float64
	v float64
}

func newSeries(base time.Time, pts ...pt) []levelSample {
	out := make([]levelSample, 0, len(pts))
	for _, p := range pts {
		out = append(out, levelSample{ts: base.Add(time.Duration(p.m * float64(time.Minute))), value: p.v})
	}
	return out
}

func TestDetectRechargeSessions(t *testing.T) {
	base := time.Date(2026, 9, 1, 15, 47, 0, 0, time.UTC)
	at := func(m float64) time.Time { return base.Add(time.Duration(m * float64(time.Minute))) }
	const (
		minDur  = rechargeDefaultMinDurationSeconds
		minRise = rechargeMinRisePct
	)

	t.Run("device sleeps through the whole charge, first reading after wake-up is the departure", func(t *testing.T) {
		// Sep 8-9 shape: ignition-off reading 44, 1,050 min of silence, wake-up reads 79 then drives off.
		// The integer odometer hides the last few hundred metres (-3..0), so the start must be the last
		// flat 44 reading, not the first one.
		soc := newSeries(base, pt{-10, 46}, pt{-5, 45}, pt{-3, 44}, pt{-2, 44}, pt{0, 44}, pt{1050, 79}, pt{1052, 79}, pt{1060, 78})
		odo := newSeries(base, pt{-10, 995}, pt{-5, 998}, pt{-3, 1000}, pt{-2, 1000}, pt{0, 1000}, pt{1048, 1000}, pt{1050, 1000}, pt{1060, 1004})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(1050), got[0].end)
		require.Equal(t, 1050*60, got[0].maxSampleGapSeconds)
	})

	t.Run("brief wake-ups that carry odometer but no SoC do not split the session", func(t *testing.T) {
		// Aug 23-24 shape: 36 h with three wake-ups reporting odometer only.
		soc := newSeries(base, pt{-5, 42}, pt{0, 40}, pt{2160, 78}, pt{2165, 77})
		odo := newSeries(base, pt{-5, 997}, pt{0, 1000}, pt{600, 1000}, pt{1200, 1000}, pt{1800, 1000}, pt{2158, 1000}, pt{2165, 1002})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(2160), got[0].end)
		require.Equal(t, 2160*60, got[0].maxSampleGapSeconds)
	})

	t.Run("start reading is the last SoC before the car stopped, even if it was read while driving", func(t *testing.T) {
		// Sep 1 shape: 23% read at 15:47 while still driving 2 km, parked 16:14, silent until 20:00,
		// then awake and charging 31->56 until 23:21.
		pts := []pt{{-30, 30}, {-15, 26}, {0, 23}}
		for m := 226; m <= 426; m += 10 {
			pts = append(pts, pt{float64(m), 31 + float64(m-226)*0.125})
		}
		pts = append(pts, pt{430, 56}, pt{440, 55})
		soc := newSeries(base, pts...)
		odo := newSeries(base, pt{-30, 995}, pt{-15, 997}, pt{0, 998}, pt{10, 999}, pt{27, 1000},
			pt{226, 1000}, pt{300, 1000}, pt{426, 1000}, pt{430, 1000}, pt{440, 1002})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start, "start must be the 23%% reading, not the 31%% wake-up reading")
		require.Equal(t, at(426), got[0].end, "end is the first sample reaching the peak")
		require.Equal(t, 226*60, got[0].maxSampleGapSeconds)
	})

	t.Run("parked trough followed by a long silence is detected", func(t *testing.T) {
		// Aug 25-26 shape: the case the old detector already handled.
		soc := newSeries(base, pt{-5, 36}, pt{0, 35}, pt{790, 95}, pt{795, 95}, pt{800, 94})
		odo := newSeries(base, pt{-5, 999}, pt{0, 1000}, pt{790, 1000}, pt{800, 1001})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(790), got[0].end)
	})

	t.Run("sensor glitch while driving is rejected by the rate cap", func(t *testing.T) {
		// Aug 20 shape: 19,18,17,17 then 80,79 within 90 s at 6 km/h.
		soc := newSeries(base, pt{0, 19}, pt{0.5, 18}, pt{1, 17}, pt{1.5, 17}, pt{2.5, 80}, pt{3, 79}, pt{10, 79}, pt{20, 78}, pt{40, 77})
		var odoPts []pt
		for m := 0.0; m <= 40; m += 0.5 {
			odoPts = append(odoPts, pt{m, 1000 + 0.1*m})
		}
		odo := newSeries(base, odoPts...)
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Empty(t, got)
	})

	t.Run("float jitter does not split a dense session", func(t *testing.T) {
		// Tesla shape: one sample per minute, ±0.3 jitter on a steady 20%/h climb.
		var pts []pt
		for m := 0; m <= 60; m++ {
			jitter := 0.3
			if m%2 == 1 {
				jitter = -0.3
			}
			pts = append(pts, pt{float64(m), 40 + float64(m)/3 + jitter})
		}
		pts = append(pts, pt{70, 59}, pt{80, 57})
		soc := newSeries(base, pts...)
		odo := newSeries(base, pt{0, 1000}, pt{30, 1000}, pt{60, 1000}, pt{70, 1005}, pt{80, 1010})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.WithinDuration(t, at(0), got[0].start, time.Minute)
		require.Equal(t, at(60), got[0].end)
		require.Equal(t, 60, got[0].maxSampleGapSeconds)
	})

	t.Run("integer staircase with one-point flicker stays one session", func(t *testing.T) {
		soc := newSeries(base, pt{0, 40}, pt{10, 41}, pt{20, 42}, pt{30, 41}, pt{40, 43}, pt{50, 44}, pt{60, 43}, pt{70, 41})
		odo := newSeries(base, pt{0, 1000}, pt{50, 1000}, pt{60, 1003}, pt{70, 1006})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(50), got[0].end)
	})

	t.Run("one-point flicker on a parked car is not a session", func(t *testing.T) {
		// Integer OBD SoC: a single 47 between 46s is quantization noise, not a 1% charge.
		soc := newSeries(base, pt{0, 46}, pt{10, 46}, pt{130, 47}, pt{131, 46}, pt{140, 45})
		odo := newSeries(base, pt{0, 1000}, pt{130, 1000}, pt{140, 1002})
		require.Empty(t, detectRechargeSessions(soc, odo, minDur, minRise))
	})

	t.Run("regen while driving has no stationary core and is dropped", func(t *testing.T) {
		var socPts, odoPts []pt
		for m := 0; m <= 15; m++ {
			socPts = append(socPts, pt{float64(m), 50 + float64(m)*0.2})
			odoPts = append(odoPts, pt{float64(m), 1000 + float64(m)})
		}
		socPts = append(socPts, pt{20, 52})
		odoPts = append(odoPts, pt{20, 1020})
		got := detectRechargeSessions(newSeries(base, socPts...), newSeries(base, odoPts...), minDur, minRise)
		require.Empty(t, got)
	})

	t.Run("minIncreasePercent override is honored", func(t *testing.T) {
		soc := newSeries(base, pt{0, 50}, pt{30, 55})
		odo := newSeries(base, pt{0, 1000}, pt{30, 1000})
		require.Len(t, detectRechargeSessions(soc, odo, minDur, 1), 1)
		require.Empty(t, detectRechargeSessions(soc, odo, minDur, 10))
	})

	t.Run("min duration is honored", func(t *testing.T) {
		soc := newSeries(base, pt{0, 20}, pt{0.5, 22})
		odo := newSeries(base, pt{0, 1000}, pt{0.5, 1000})
		require.Empty(t, detectRechargeSessions(soc, odo, 60, minRise))
		require.Len(t, detectRechargeSessions(soc, odo, 10, minRise), 1)
	})

	t.Run("no odometer data keeps the trough-to-peak rise", func(t *testing.T) {
		soc := newSeries(base, pt{0, 44}, pt{600, 79})
		got := detectRechargeSessions(soc, nil, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(600), got[0].end)
	})

	t.Run("two sessions within 2h with equal odometer merge and the gap is recomputed", func(t *testing.T) {
		soc := newSeries(base, pt{0, 30}, pt{30, 40}, pt{60, 50}, pt{61, 48}, pt{121, 48}, pt{150, 60}, pt{180, 70}, pt{190, 69})
		odo := newSeries(base, pt{0, 1000}, pt{60, 1000}, pt{61, 1000}, pt{121, 1000}, pt{180, 1000}, pt{190, 1005})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 1)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(180), got[0].end)
		require.Equal(t, 60*60, got[0].maxSampleGapSeconds)
	})

	t.Run("two sessions separated by a drive stay separate", func(t *testing.T) {
		soc := newSeries(base, pt{0, 30}, pt{30, 40}, pt{60, 50}, pt{61, 48}, pt{90, 46}, pt{121, 46}, pt{150, 60}, pt{180, 70}, pt{190, 69})
		odo := newSeries(base, pt{0, 1000}, pt{60, 1000}, pt{61, 1000}, pt{90, 1003}, pt{121, 1005}, pt{180, 1005}, pt{190, 1010})
		got := detectRechargeSessions(soc, odo, minDur, minRise)
		require.Len(t, got, 2)
		require.Equal(t, at(0), got[0].start)
		require.Equal(t, at(60), got[0].end)
		require.Equal(t, at(121), got[1].start)
		require.Equal(t, at(180), got[1].end)
	})

	t.Run("fewer than two SoC samples yields nothing", func(t *testing.T) {
		require.Empty(t, detectRechargeSessions(nil, nil, minDur, minRise))
		require.Empty(t, detectRechargeSessions(newSeries(base, pt{0, 50}), nil, minDur, minRise))
	})
}

func TestRechargeSessionsToSegments(t *testing.T) {
	base := time.Date(2026, 9, 8, 19, 19, 47, 0, time.UTC)
	from := base.Add(-24 * time.Hour)
	sessions := []rechargeSession{{start: base, end: base.Add(62880 * time.Second), maxSampleGapSeconds: 62880}}

	segs := rechargeSessionsToSegments(sessions, from)
	require.Len(t, segs, 1)
	require.Equal(t, base, segs[0].Start.Timestamp)
	require.NotNil(t, segs[0].End)
	require.Equal(t, base.Add(62880*time.Second), segs[0].End.Timestamp)
	require.Equal(t, 62880, segs[0].Duration)
	require.False(t, segs[0].IsOngoing)
	require.False(t, segs[0].StartedBeforeRange)
	require.NotNil(t, segs[0].MaxSampleGapSeconds)
	require.Equal(t, 62880, *segs[0].MaxSampleGapSeconds)

	require.Empty(t, rechargeSessionsToSegments(nil, from))
	require.NotNil(t, rechargeSessionsToSegments(nil, from))
}

func TestLevelFirstLastInRange(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("returns first and last in range", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 10},
			{ts: min(5), value: 50},
			{ts: min(10), value: 90},
		}
		first, last, ok := levelFirstLastInRange(samples, min(0), min(10))
		require.True(t, ok)
		require.Equal(t, 10.0, first)
		require.Equal(t, 90.0, last)
	})

	t.Run("no samples in range", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 10},
		}
		_, _, ok := levelFirstLastInRange(samples, min(5), min(10))
		require.False(t, ok)
	})

	t.Run("empty samples", func(t *testing.T) {
		_, _, ok := levelFirstLastInRange(nil, min(0), min(10))
		require.False(t, ok)
	})
}
