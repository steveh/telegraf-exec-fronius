package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"golang.org/x/net/context/ctxhttp"
)

type powerFlowRealtimeResponse struct {
	Body struct {
		Data struct {
			Version   string `json:"Version"`
			Inverters map[string]struct {
				// device type of inverter
				DeviceType int `json:"DT"`

				// AC Energy [Wh] this day
				EnergyDay float64 `json:"E_Day"`

				// AC Energy [Wh] this year
				EnergyYear float64 `json:"E_Year"`

				// AC Energy [Wh] ever
				EnergyTotal float64 `json:"E_Total"`

				// current power in Watt
				// +ve: produce/export
				// -ve: consume/import
				Power float64 `json:"P"`
			} `json:"Inverters"`
			Site struct {
				// AC Energy [Wh] this day
				EnergyDay float64 `json:"E_Day"`

				// AC Energy [Wh] this year
				EnergyYear float64 `json:"E_Year"`

				// AC Energy [Wh] ever
				EnergyTotal float64 `json:"E_Total"`

				// "load", "grid" or "unknown"
				MeterLocation string `json:"Meter_Location"`

				// "produce-only": inverter only
				// "meter", "vague-meter": inverter and meter
				// "bidirectional" or "ac-coupled": inverter, meter, and battery
				Mode string `json:"Mode"`

				// +ve: discharge
				// -ve: charge
				PowerCumulative float64 `json:"P_Akku"`

				// +ve: from grid
				// -ve: to grid
				PowerGrid float64 `json:"P_Grid"`

				// +ve: generator
				// -ve: consumer
				PowerLoad float64 `json:"P_Load"`

				// +ve: production
				PowerConsumption float64 `json:"P_PV"`

				// current relative autonomy in %
				RelativeAutonomy float64 `json:"rel_Autonomy"`

				// current relative self consumption in %
				RelativeSelfConsumption float64 `json:"rel_SelfConsumption"`
			} `json:"Site"`
		} `json:"Data"`
	} `json:"Body"`
	Head head `json:"head"`
}

// PowerFlowRealtime returns realtime power flow data.
func (c Client) PowerFlowRealtime(ctx context.Context) (points []*write.Point, err error) {
	var r powerFlowRealtimeResponse

	u := url.URL{Scheme: "http", Host: c.host, Path: "/solar_api/v1/GetPowerFlowRealtimeData.fcgi"}

	res, err := ctxhttp.Get(ctx, c.client, u.String())
	if err != nil {
		return points, err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = cErr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return points, fmt.Errorf("%w: %d", ErrStatusNotOk, res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return points, err
	}

	for deviceID, deviceData := range r.Body.Data.Inverters {
		tags := map[string]string{
			"device_id":    deviceID,
			"device_class": "inverter",
		}

		values := map[string]interface{}{
			"energy_day":   deviceData.EnergyDay,
			"energy_year":  deviceData.EnergyYear,
			"energy_total": deviceData.EnergyTotal,
			"power":        deviceData.Power,
		}

		point := influxdb2.NewPoint("fronius_powerflow", tags, values, r.Head.Timestamp)

		points = append(points, point)
	}

	tags := map[string]string{
		"device_class": "site",
	}

	values := map[string]interface{}{
		"energy_day":                r.Body.Data.Site.EnergyDay,
		"energy_year":               r.Body.Data.Site.EnergyYear,
		"energy_total":              r.Body.Data.Site.EnergyTotal,
		"power_cumulative":          r.Body.Data.Site.PowerCumulative,
		"power_grid":                r.Body.Data.Site.PowerGrid,
		"power_load":                r.Body.Data.Site.PowerLoad,
		"power_consumption":         r.Body.Data.Site.PowerConsumption,
		"relative_autonomy":         r.Body.Data.Site.RelativeAutonomy,
		"relative_self_consumption": r.Body.Data.Site.RelativeSelfConsumption,
	}

	point := influxdb2.NewPoint("fronius_powerflow", tags, values, r.Head.Timestamp)

	points = append(points, point)

	return points, nil
}
