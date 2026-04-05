package postgres

import (
	"context"
	"database/sql"
	"errors"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

var ErrMetricNotFound = errors.New("metric is not found")

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO metrics (id, mtype, value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (id)
		DO UPDATE SET
			mtype = 'gauge',
			value = EXCLUDED.value,
			delta = NULL`, name, value,
	)
	return err
}

func (s *Storage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO metrics (id, mtype, delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (id)
		DO UPDATE SET
			mtype = 'counter',
			delta = metrics.delta + EXCLUDED.delta,
			value = NULL`, name, delta,
	)
	return err
}
func (s *Storage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64

	err := s.db.QueryRowContext(
		ctx,
		`SELECT value
		FROM metrics
		WHERE id = $1 AND mtype = 'gauge'`,
		name,
	).Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrMetricNotFound
		}
		return 0, err
	}
	return value, nil
}

func (s *Storage) GetCounter(ctx context.Context, name string) (int64, error) {
	var delta int64

	err := s.db.QueryRowContext(
		ctx,
		`SELECT delta
		FROM metrics
		WHERE id = $1 AND mtype = 'counter'`,
		name,
	).Scan(&delta)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrMetricNotFound
		}
		return 0, err
	}
	return delta, nil
}

func (s *Storage) GetAll(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, mtype, value, delta FROM metrics`,
	)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			id    string
			mType string
			value sql.NullFloat64
			delta sql.NullInt64
		)
		if err := rows.Scan(&id, &mType, &value, &delta); err != nil {
			return nil, nil, err
		}

		switch mType {
		case "gauge":
			if value.Valid {
				gauges[id] = value.Float64
			}
		case "counter":
			if delta.Valid {
				counters[id] = delta.Int64
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return gauges, counters, nil

}
func (s *Storage) SetAll(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `TRUNCATE metrics`); err != nil {
		return err
	}

	for id, value := range gauges {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metrics (id, mtype, value) VALUES ($1, 'gauge', $2)`,
			id, value,
		)
		if err != nil {
			return err
		}
	}

	for id, delta := range counters {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metrics (id, mtype, delta) VALUES ($1, 'counter', $2)`,
			id, delta,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	return withRetry(ctx, func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer tx.Rollback()

		for _, m := range metrics {
			switch m.MType {
			case models.Gauge:
				if m.Value == nil {
					return errors.New("value is nil")
				}

				_, err = tx.ExecContext(ctx,
					`INSERT INTO metrics (id, mtype, value)
				VALUES ($1, 'gauge', $2)
				ON CONFLICT (id)
				DO UPDATE SET
					mtype = 'gauge',
					value = EXCLUDED.value,
					delta = NULL`,
					m.ID, *m.Value,
				)
				if err != nil {
					return err
				}
			case models.Counter:
				if m.Delta == nil {
					return errors.New("delta is nil")
				}

				_, err = tx.ExecContext(ctx,
					`INSERT INTO metrics (id, mtype, delta)
				 VALUES ($1, 'counter', $2)
				 ON CONFLICT (id)
				 DO UPDATE SET
				     mtype = 'counter',
				     delta = metrics.delta + EXCLUDED.delta,
				     value = NULL`,
					m.ID, *m.Delta,
				)
				if err != nil {
					return err
				}

			default:
				return errors.New("invalid metric type")
			}
		}
		return tx.Commit()
	})
}
