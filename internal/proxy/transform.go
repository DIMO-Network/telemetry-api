package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
)

// ToSubject converts a vehicle token ID into a DID subject string for dq.
func ToSubject(chainID uint64, vehicleAddr common.Address, tokenID int) string {
	return cloudevent.ERC721DID{
		ChainID:         chainID,
		ContractAddress: vehicleAddr,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()
}

// dqMechanism maps telemetry-api's camelCase DetectionMechanism values to dq's UPPER_SNAKE_CASE.
var dqMechanism = map[model.DetectionMechanism]string{
	model.DetectionMechanismIgnitionDetection:    "IGNITION_DETECTION",
	model.DetectionMechanismFrequencyAnalysis:    "FREQUENCY_ANALYSIS",
	model.DetectionMechanismChangePointDetection: "CHANGE_POINT_DETECTION",
	model.DetectionMechanismIdling:               "IDLING",
	model.DetectionMechanismRefuel:               "REFUEL",
	model.DetectionMechanismRecharge:             "RECHARGE",
}

// ToDQMechanism returns the dq-format DetectionMechanism string for the given telemetry-api value.
func ToDQMechanism(m model.DetectionMechanism) (string, error) {
	v, ok := dqMechanism[m]
	if !ok {
		return "", fmt.Errorf("unknown detection mechanism: %s", m)
	}
	return v, nil
}

// BuildSignalsQuery constructs the dq GraphQL query for the signals resolver.
// Field selections are built from aggArgs to preserve aliases.
// It returns the query string and any per-signal filter variables that must be included
// in the variables map alongside subject/interval/from/to/filter.
func BuildSignalsQuery(aggArgs *model.AggregatedSignalArgs) (query string, filterVars map[string]any) {
	filterVars = make(map[string]any)

	var varDecls strings.Builder // extra variable declarations
	var fields strings.Builder   // field selection set

	fields.WriteString(`timestamp`)

	for i, fa := range aggArgs.FloatArgs {
		if fa.Filter != nil {
			varName := fmt.Sprintf("ff%d", i)
			fmt.Fprintf(&varDecls, `, $%s: SignalFloatFilter`, varName)
			fmt.Fprintf(&fields, ` %s: %s(agg: %s, filter: $%s)`, fa.Alias, fa.Name, fa.Agg, varName)
			filterVars[varName] = fa.Filter
		} else {
			fmt.Fprintf(&fields, ` %s: %s(agg: %s)`, fa.Alias, fa.Name, fa.Agg)
		}
	}
	for _, sa := range aggArgs.StringArgs {
		fmt.Fprintf(&fields, ` %s: %s(agg: %s)`, sa.Alias, sa.Name, sa.Agg)
	}
	for i, la := range aggArgs.LocationArgs {
		if la.Filter != nil {
			varName := fmt.Sprintf("lf%d", i)
			fmt.Fprintf(&varDecls, `, $%s: SignalLocationFilter`, varName)
			fmt.Fprintf(&fields, ` %s: %s(agg: %s, filter: $%s) { latitude longitude hdop }`, la.Alias, la.Name, la.Agg, varName)
			filterVars[varName] = la.Filter
		} else {
			fmt.Fprintf(&fields, ` %s: %s(agg: %s) { latitude longitude hdop }`, la.Alias, la.Name, la.Agg)
		}
	}

	query = fmt.Sprintf(
		`query Signals($subject: String!, $interval: String!, $from: Time!, $to: Time!, $filter: SignalFilter%s) {signals(subject: $subject, interval: $interval, from: $from, to: $to, filter: $filter) {%s}}`,
		varDecls.String(), fields.String(),
	)
	return query, filterVars
}

// UnmarshalSignalsResponse parses the dq response JSON for the signals query into
// []*model.SignalAggregations, preserving the alias→value mapping needed by the
// generated field resolvers.
func UnmarshalSignalsResponse(data json.RawMessage, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	var envelope struct {
		Signals []map[string]json.RawMessage `json:"signals"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling signals: %w", err)
	}

	result := make([]*model.SignalAggregations, 0, len(envelope.Signals))
	for _, row := range envelope.Signals {
		agg := &model.SignalAggregations{
			ValueNumbers:   make(map[string]float64),
			ValueStrings:   make(map[string]string),
			ValueLocations: make(map[string]vss.Location),
		}
		if raw, ok := row["timestamp"]; ok {
			if err := json.Unmarshal(raw, &agg.Timestamp); err != nil {
				return nil, fmt.Errorf("unmarshaling timestamp: %w", err)
			}
		}
		for _, fa := range aggArgs.FloatArgs {
			if raw, ok := row[fa.Alias]; ok && string(raw) != "null" {
				var v float64
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("unmarshaling float %s: %w", fa.Alias, err)
				}
				agg.ValueNumbers[fa.Alias] = v
			}
		}
		for _, sa := range aggArgs.StringArgs {
			if raw, ok := row[sa.Alias]; ok && string(raw) != "null" {
				var v string
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("unmarshaling string %s: %w", sa.Alias, err)
				}
				agg.ValueStrings[sa.Alias] = v
			}
		}
		for _, la := range aggArgs.LocationArgs {
			if raw, ok := row[la.Alias]; ok && string(raw) != "null" {
				var loc struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
					Hdop      float64 `json:"hdop"`
				}
				if err := json.Unmarshal(raw, &loc); err != nil {
					return nil, fmt.Errorf("unmarshaling location %s: %w", la.Alias, err)
				}
				agg.ValueLocations[la.Alias] = vss.Location{
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
					HDOP:      loc.Hdop,
				}
			}
		}
		result = append(result, agg)
	}
	return result, nil
}

// BuildSignalsLatestQuery constructs the dq GraphQL query for the signalsLatest resolver.
func BuildSignalsLatestQuery(latestArgs *model.LatestSignalsArgs) string {
	var sb strings.Builder
	sb.WriteString(`query SignalsLatest($subject: String!, $filter: SignalFilter) {`)
	sb.WriteString(`signalsLatest(subject: $subject, filter: $filter) {`)
	if latestArgs.IncludeLastSeen {
		sb.WriteString(`lastSeen `)
	}
	for name := range latestArgs.SignalNames {
		fmt.Fprintf(&sb, `%s { timestamp value } `, name)
	}
	for name := range latestArgs.LocationSignalNames {
		fmt.Fprintf(&sb, `%s { timestamp value { latitude longitude hdop } } `, name)
	}
	sb.WriteString(`} }`)
	return sb.String()
}

// UnmarshalSignalsLatestResponse parses the dq response into *model.SignalCollection.
func UnmarshalSignalsLatestResponse(data json.RawMessage) (*model.SignalCollection, error) {
	var envelope struct {
		SignalsLatest *model.SignalCollection `json:"signalsLatest"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling signalsLatest: %w", err)
	}
	return envelope.SignalsLatest, nil
}

// BuildAvailableSignalsQuery constructs the dq query for availableSignals.
func BuildAvailableSignalsQuery() string {
	return `query AvailableSignals($subject: String!, $filter: SignalFilter) {
		availableSignals(subject: $subject, filter: $filter)
	}`
}

// UnmarshalAvailableSignalsResponse parses the dq response into []string.
func UnmarshalAvailableSignalsResponse(data json.RawMessage) ([]string, error) {
	var envelope struct {
		AvailableSignals []string `json:"availableSignals"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling availableSignals: %w", err)
	}
	return envelope.AvailableSignals, nil
}

// BuildDataSummaryQuery constructs the dq query for dataSummary.
func BuildDataSummaryQuery() string {
	return `query DataSummary($subject: String!, $filter: SignalFilter) {
		dataSummary(subject: $subject, filter: $filter) {
			numberOfSignals availableSignals firstSeen lastSeen
			signalDataSummary { name numberOfSignals firstSeen lastSeen }
			eventDataSummary { name numberOfEvents firstSeen lastSeen }
		}
	}`
}

// UnmarshalDataSummaryResponse parses the dq response into *model.DataSummary.
func UnmarshalDataSummaryResponse(data json.RawMessage) (*model.DataSummary, error) {
	var envelope struct {
		DataSummary *model.DataSummary `json:"dataSummary"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling dataSummary: %w", err)
	}
	return envelope.DataSummary, nil
}

// BuildEventsQuery constructs the dq query for events.
func BuildEventsQuery() string {
	return `query Events($subject: String!, $from: Time!, $to: Time!, $filter: EventFilter) {
		events(subject: $subject, from: $from, to: $to, filter: $filter) {
			timestamp name source durationNs metadata
		}
	}`
}

// UnmarshalEventsResponse parses the dq response into []*model.Event.
func UnmarshalEventsResponse(data json.RawMessage) ([]*model.Event, error) {
	var envelope struct {
		Events []*model.Event `json:"events"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling events: %w", err)
	}
	return envelope.Events, nil
}

// BuildSegmentsQuery constructs the dq query for segments.
func BuildSegmentsQuery() string {
	return `query Segments($subject: String!, $from: Time!, $to: Time!, $mechanism: DetectionMechanism!,
		$config: SegmentConfig, $signalRequests: [SegmentSignalRequest!], $eventRequests: [SegmentEventRequest!],
		$limit: Int, $after: Time) {
		segments(subject: $subject, from: $from, to: $to, mechanism: $mechanism,
			config: $config, signalRequests: $signalRequests, eventRequests: $eventRequests,
			limit: $limit, after: $after) {
			start { timestamp value { latitude longitude hdop } }
			end { timestamp value { latitude longitude hdop } }
			duration isOngoing startedBeforeRange
			signals { name agg value }
			eventCounts { name count }
		}
	}`
}

// UnmarshalSegmentsResponse parses the dq response into []*model.Segment.
func UnmarshalSegmentsResponse(data json.RawMessage) ([]*model.Segment, error) {
	var envelope struct {
		Segments []*model.Segment `json:"segments"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling segments: %w", err)
	}
	return envelope.Segments, nil
}

// BuildDailyActivityQuery constructs the dq query for dailyActivity.
func BuildDailyActivityQuery() string {
	return `query DailyActivity($subject: String!, $from: Time!, $to: Time!, $mechanism: DetectionMechanism!,
		$config: SegmentConfig, $signalRequests: [SegmentSignalRequest!], $eventRequests: [SegmentEventRequest!],
		$timezone: String) {
		dailyActivity(subject: $subject, from: $from, to: $to, mechanism: $mechanism,
			config: $config, signalRequests: $signalRequests, eventRequests: $eventRequests,
			timezone: $timezone) {
			start { timestamp value { latitude longitude hdop } }
			end { timestamp value { latitude longitude hdop } }
			segmentCount duration
			signals { name agg value }
			eventCounts { name count }
		}
	}`
}

// UnmarshalDailyActivityResponse parses the dq response into []*model.DailyActivity.
func UnmarshalDailyActivityResponse(data json.RawMessage) ([]*model.DailyActivity, error) {
	var envelope struct {
		DailyActivity []*model.DailyActivity `json:"dailyActivity"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshaling dailyActivity: %w", err)
	}
	return envelope.DailyActivity, nil
}


// TimeVar formats a time.Time for use as a GraphQL Time variable.
func TimeVar(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// ProxySignals executes the dq signals query and returns telemetry-api model results.
func (c *Client) ProxySignals(ctx context.Context, subject, interval string, from, to time.Time, filter *model.SignalFilter, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	query, filterVars := BuildSignalsQuery(aggArgs)
	vars := map[string]any{
		"subject":  subject,
		"interval": interval,
		"from":     TimeVar(from),
		"to":       TimeVar(to),
		"filter":   filter,
	}
	for k, v := range filterVars {
		vars[k] = v
	}
	data, err := c.Execute(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	return UnmarshalSignalsResponse(data, aggArgs)
}

// ProxySignalsLatest executes the dq signalsLatest query and returns the result.
func (c *Client) ProxySignalsLatest(ctx context.Context, subject string, filter *model.SignalFilter, latestArgs *model.LatestSignalsArgs) (*model.SignalCollection, error) {
	query := BuildSignalsLatestQuery(latestArgs)
	vars := map[string]any{
		"subject": subject,
		"filter":  filter,
	}
	data, err := c.Execute(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	return UnmarshalSignalsLatestResponse(data)
}

// ProxyAvailableSignals executes the dq availableSignals query and returns the result.
func (c *Client) ProxyAvailableSignals(ctx context.Context, subject string, filter *model.SignalFilter) ([]string, error) {
	data, err := c.Execute(ctx, BuildAvailableSignalsQuery(), map[string]any{
		"subject": subject,
		"filter":  filter,
	})
	if err != nil {
		return nil, err
	}
	return UnmarshalAvailableSignalsResponse(data)
}

// ProxyDataSummary executes the dq dataSummary query and returns the result.
func (c *Client) ProxyDataSummary(ctx context.Context, subject string, filter *model.SignalFilter) (*model.DataSummary, error) {
	data, err := c.Execute(ctx, BuildDataSummaryQuery(), map[string]any{
		"subject": subject,
		"filter":  filter,
	})
	if err != nil {
		return nil, err
	}
	return UnmarshalDataSummaryResponse(data)
}

// ProxyEvents executes the dq events query and returns the result.
func (c *Client) ProxyEvents(ctx context.Context, subject string, from, to time.Time, filter *model.EventFilter) ([]*model.Event, error) {
	data, err := c.Execute(ctx, BuildEventsQuery(), map[string]any{
		"subject": subject,
		"from":    TimeVar(from),
		"to":      TimeVar(to),
		"filter":  filter,
	})
	if err != nil {
		return nil, err
	}
	return UnmarshalEventsResponse(data)
}

// ProxySegments executes the dq segments query and returns the result.
func (c *Client) ProxySegments(ctx context.Context, subject string, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig, signalRequests []*model.SegmentSignalRequest, eventRequests []*model.SegmentEventRequest, limit *int, after *time.Time) ([]*model.Segment, error) {
	mech, err := ToDQMechanism(mechanism)
	if err != nil {
		return nil, err
	}
	vars := map[string]any{
		"subject":        subject,
		"from":           TimeVar(from),
		"to":             TimeVar(to),
		"mechanism":      mech,
		"config":         config,
		"signalRequests": signalRequests,
		"eventRequests":  eventRequests,
		"limit":          limit,
		"after":          nil,
	}
	if after != nil {
		vars["after"] = TimeVar(*after)
	}
	data, err := c.Execute(ctx, BuildSegmentsQuery(), vars)
	if err != nil {
		return nil, err
	}
	return UnmarshalSegmentsResponse(data)
}

// ProxyDailyActivity executes the dq dailyActivity query and returns the result.
func (c *Client) ProxyDailyActivity(ctx context.Context, subject string, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig, signalRequests []*model.SegmentSignalRequest, eventRequests []*model.SegmentEventRequest, timezone *string) ([]*model.DailyActivity, error) {
	mech, err := ToDQMechanism(mechanism)
	if err != nil {
		return nil, err
	}
	data, err := c.Execute(ctx, BuildDailyActivityQuery(), map[string]any{
		"subject":        subject,
		"from":           TimeVar(from),
		"to":             TimeVar(to),
		"mechanism":      mech,
		"config":         config,
		"signalRequests": signalRequests,
		"eventRequests":  eventRequests,
		"timezone":       timezone,
	})
	if err != nil {
		return nil, err
	}
	return UnmarshalDailyActivityResponse(data)
}
