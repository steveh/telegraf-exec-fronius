package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"golang.org/x/net/context/ctxhttp"
)

type inverterRealtimeResponse struct {
	Body struct {
		Data struct {
			// Status information about inverter
			DeviceStatus struct {
				ErrorCode              int  `json:"ErrorCode"`
				LEDColor               int  `json:"LEDColor"`
				LEDState               int  `json:"LEDState"`
				MgmtTimerRemainingTime int  `json:"MgmtTimerRemainingTime"`
				StateToReset           bool `json:"StateToReset"`
				StatusCode             int  `json:"StatusCode"`
			} `json:"DeviceStatus"`

			// AC current (absolute, accumulated over all lines)
			CurrentAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"IAC"`

			// DC current
			CurrentDC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"IDC"`

			// AC voltage
			VoltageAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"UAC"`

			// DC voltage
			VoltageDC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"UDC"`

			// AC power (negative value for consuming power)
			PowerAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"PAC"`

			// AC frequency
			FrequencyAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"FAC"`

			// AC Energy generated on current day
			EnergyDayAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"DAY_ENERGY"`

			// AC Energy generated in current year
			EnergyYearAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"YEAR_ENERGY"`

			// AC Energy generated overall
			EnergyTotalAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"TOTAL_ENERGY"`
		} `json:"Data"`
	} `json:"Body"`
	Head head `json:"head"`
}

type inverterMinMaxResponse struct {
	Body struct {
		Data struct {
			// Maximum AC power of current day
			PowerDayMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"DAY_PMAX"`

			// Maximum AC voltage of current day
			VoltageDayMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"DAY_UACMAX"`

			// Minimum AC voltage of current day
			VoltageDayMinAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"DAY_UACMIN"`

			// Maximum DC voltage of current day
			VoltageDayMaxDC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"DAY_UDCMAX"`

			// Maximum AC power of current year
			PowerYearMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"YEAR_PMAX"`

			// Maximum AC voltage of current year
			VoltageYearMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"YEAR_UACMAX"`

			// Minimum AC voltage of current year
			VoltageYearMinAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"YEAR_UACMIN"`

			// Maximum DC voltage of current year
			VoltageYearMaxDC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"YEAR_UDCMAX"`

			// Maximum AC power overall
			PowerTotalMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"TOTAL_PMAX"`

			// Maximum AC voltage overall
			VoltageTotalMaxAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"TOTAL_UACMAX"`

			// Minimum AC voltage overall
			VoltageTotalMinAC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"TOTAL_UACMIN"`

			// Maximum DC voltage overall
			VoltageTotalMaxDC struct {
				Unit  string  `json:"Unit"`
				Value float64 `json:"Value"`
			} `json:"TOTAL_UDCMAX"`
		} `json:"Data"`
	} `json:"Body"`
	Head head `json:"head"`
}

// InverterRealtime returns realtime inverter data.
func (c Client) InverterRealtime(ctx context.Context, deviceID string) (points []*write.Point, err error) {
	var r inverterRealtimeResponse

	q := url.Values{}

	q.Set("Scope", "Device")
	q.Set("DeviceId", deviceID)
	q.Set("DataCollection", "CommonInverterData")

	u := url.URL{Scheme: "http", Host: c.host, Path: "/solar_api/v1/GetInverterRealtimeData.cgi", RawQuery: q.Encode()}

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

	tags := map[string]string{
		"device_id": deviceID,
	}

	values := map[string]interface{}{
		"current_ac":      r.Body.Data.CurrentAC.Value,
		"current_dc":      r.Body.Data.CurrentDC.Value,
		"voltage_ac":      r.Body.Data.VoltageAC.Value,
		"voltage_dc":      r.Body.Data.VoltageDC.Value,
		"power_ac":        r.Body.Data.PowerAC.Value,
		"frequency_ac":    r.Body.Data.FrequencyAC.Value,
		"energy_day_ac":   r.Body.Data.EnergyDayAC.Value,
		"energy_year_ac":  r.Body.Data.EnergyYearAC.Value,
		"energy_total_ac": r.Body.Data.EnergyTotalAC.Value,
	}

	return []*write.Point{
		influxdb2.NewPoint("fronius_inverter", tags, values, r.Head.Timestamp),
	}, nil
}

// InverterMinMax returns minimum and maximum inverter data.
func (c Client) InverterMinMax(ctx context.Context, deviceID string) (points []*write.Point, err error) {
	var r inverterMinMaxResponse

	q := url.Values{}

	q.Set("Scope", "Device")
	q.Set("DeviceId", deviceID)
	q.Set("DataCollection", "MinMaxInverterData")

	u := url.URL{Scheme: "http", Host: c.host, Path: "/solar_api/v1/GetInverterRealtimeData.cgi", RawQuery: q.Encode()}

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

	tags := map[string]string{
		"device_id": deviceID,
	}

	values := map[string]interface{}{
		"power_day_max_ac":     r.Body.Data.PowerDayMaxAC.Value,
		"voltage_day_max_ac":   r.Body.Data.VoltageDayMaxAC.Value,
		"voltage_day_min_ac":   r.Body.Data.VoltageDayMinAC.Value,
		"voltage_day_max_dc":   r.Body.Data.VoltageDayMaxDC.Value,
		"power_year_max_ac":    r.Body.Data.PowerYearMaxAC.Value,
		"voltage_year_max_ac":  r.Body.Data.VoltageYearMaxAC.Value,
		"voltage_year_min_ac":  r.Body.Data.VoltageYearMinAC.Value,
		"voltage_year_max_dc":  r.Body.Data.VoltageYearMaxDC.Value,
		"power_total_max_ac":   r.Body.Data.PowerTotalMaxAC.Value,
		"voltage_total_max_ac": r.Body.Data.VoltageTotalMaxAC.Value,
		"voltage_total_min_ac": r.Body.Data.VoltageTotalMinAC.Value,
		"voltage_total_max_dc": r.Body.Data.VoltageTotalMaxDC.Value,
	}

	return []*write.Point{
		influxdb2.NewPoint("fronius_inverter_minmax", tags, values, r.Head.Timestamp),
	}, nil
}

// InverterArchive returns historical inverter data.
func (c Client) InverterArchive(ctx context.Context, deviceID string, startDate time.Time, endDate time.Time) (points []*write.Point, err error) {
	q := defaultValues()

	q.Set("Scope", "Device")
	q.Set("SeriesType", "Detail")
	q.Set("HumanReadable", "False")
	q.Set("DeviceClass", "Inverter")
	q.Set("DeviceId", deviceID)
	q.Set("StartDate", startDate.Format("2006-01-02"))
	q.Set("EndDate", endDate.Format("2006-01-02"))

	r, err := c.readArchive(ctx, q)
	if err != nil {
		return points, err
	}

	return generateArchivePoints(r, "inverter_archive")
}
