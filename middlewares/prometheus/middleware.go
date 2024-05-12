package prometheus

import (
	"context"
	go_orm "github.com/Andras5014/go-orm"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m *MiddlewareBuilder) Build() go_orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})

	prometheus.MustRegister(vector)
	return func(next go_orm.Handler) go_orm.Handler {
		return func(ctx context.Context, qc *go_orm.QueryContext) *go_orm.QueryResult {
			startTime := time.Now()
			defer func() {
				// 统计耗时
				vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(time.Since(startTime).Milliseconds()))
			}()

			return next(ctx, qc)

		}
	}
}
