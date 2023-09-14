package llo

// TODO: llo datasource
import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-data-streams/llo"

	"github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/streams"
)

var (
	promMissingStreamCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llo_stream_missing_count",
		Help: "Number of times we tried to observe a stream, but it was missing",
	},
		[]string{"streamID"},
	)
	promObservationErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llo_stream_observation_error_count",
		Help: "Number of times we tried to observe a stream, but it failed with an error",
	},
		[]string{"streamID"},
	)
)

type ErrMissingStream struct {
	id string
}

type Registry interface {
	Get(streamID streams.StreamID) (strm streams.Stream, exists bool)
}

func (e ErrMissingStream) Error() string {
	return fmt.Sprintf("missing stream definition for: %q", e.id)
}

var _ llo.DataSource = &dataSource{}

type dataSource struct {
	lggr     logger.Logger
	registry Registry
}

func NewDataSource(lggr logger.Logger, registry Registry) llo.DataSource {
	// TODO: lggr should include job ID
	return &dataSource{lggr, registry}
}

// Observe looks up all streams in the registry and returns a map of stream ID => value
func (d *dataSource) Observe(ctx context.Context, streamIDs map[commontypes.StreamID]struct{}) (llo.StreamValues, error) {
	var wg sync.WaitGroup
	wg.Add(len(streamIDs))
	sv := make(llo.StreamValues)
	var mu sync.Mutex

	for streamID := range streamIDs {
		go func(streamID commontypes.StreamID) {
			defer wg.Done()

			var res llo.ObsResult[*big.Int]

			stream, exists := d.registry.Get(streamID)
			if exists {
				run, trrs, err := stream.Run(ctx)
				if err != nil {
					var runID int64
					if run != nil {
						runID = run.ID
					}
					d.lggr.Debugw("Observation failed for stream", "err", err, "streamID", streamID, "runID", runID)
					promObservationErrorCount.WithLabelValues(fmt.Sprintf("%d", streamID)).Inc()
				} else {
					// TODO: support types other than *big.Int
					val, err := streams.ExtractBigInt(trrs)
					if err == nil {
						res.Val = val
						res.Valid = true
					}
				}
			} else {
				d.lggr.Errorw(fmt.Sprintf("Missing stream: %q", streamID), "streamID", streamID)
				promMissingStreamCount.WithLabelValues(fmt.Sprintf("%d", streamID)).Inc()
			}

			mu.Lock()
			defer mu.Unlock()
			sv[streamID] = res
		}(streamID)
	}

	wg.Wait()

	return sv, nil
}

func ptr[T any](t T) *T { return &t }
