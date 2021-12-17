// Package main retrieves data from a Fronius data logger
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

const (
	precision = time.Nanosecond
)

var (
	host     string
	inverter string
	meter    string
	system   bool
	realtime bool
	archive  bool
	days     uint
)

func init() {
	flag.StringVar(&host, "host", "localhost", "Fronius host")
	flag.StringVar(&inverter, "inverter", "1", "Collect inverter data with device ID")
	flag.StringVar(&meter, "meter", "0", "Collect meter data with device ID")
	flag.BoolVar(&system, "system", true, "Collect system data")
	flag.BoolVar(&realtime, "realtime", false, "Collect realtime data")
	flag.BoolVar(&archive, "archive", false, "Collect archive data")
	flag.UintVar(&days, "days", 7, "Days of history to collect")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	client := NewClient(host)

	ctx := context.Background()

	if realtime {
		if inverter != "" {
			inv_points, err := client.InverterRealtime(ctx, inverter)
			check(err)
			for _, p := range inv_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}

		if meter != "" {
			met_points, err := client.MeterRealtime(ctx, meter)
			check(err)
			for _, p := range met_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}

		if system {
			pow_points, err := client.PowerFlowRealtime(ctx)
			check(err)
			for _, p := range pow_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}
	}

	if archive {
		startDate := time.Now().AddDate(0, 0, 0-int(days))
		endDate := time.Now()

		if inverter != "" {
			inv_points, err := client.InverterArchive(ctx, inverter, startDate, endDate)
			check(err)
			for _, p := range inv_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}

			minmax, err := client.InverterMinMax(ctx, inverter)
			check(err)
			for _, p := range minmax {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}

		if meter != "" {
			met_points, err := client.MeterArchive(ctx, meter, startDate, endDate)
			check(err)
			for _, p := range met_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}

		if system {
			sys_points, err := client.SystemArchive(ctx, startDate, endDate)
			check(err)
			for _, p := range sys_points {
				fmt.Print(write.PointToLineProtocol(p, precision))
			}
		}
	}
}
