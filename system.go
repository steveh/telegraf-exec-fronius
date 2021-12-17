package main

import (
	"context"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// SystemArchive returns historical system data.
func (c Client) SystemArchive(ctx context.Context, startDate time.Time, endDate time.Time) (points []*write.Point, err error) {
	q := defaultValues()

	q.Set("Scope", "System")
	q.Set("SeriesType", "Detail")
	q.Set("HumanReadable", "False")
	q.Set("StartDate", startDate.Format("2006-01-02"))
	q.Set("EndDate", endDate.Format("2006-01-02"))

	r, err := c.readArchive(ctx, q)
	if err != nil {
		return points, err
	}

	return generateArchivePoints(r, "system_archive")
}
