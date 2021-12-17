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

type meterRealtimeResponse struct {
	Body struct {
		Data struct {
			Details struct {
				Manufacturer string `json:"Manufacturer"`
				Model        string `json:"Model"`
				Serial       string `json:"Serial"`
			} `json:"Details"`

			// 1...enabled, 0...disabled
			Enable int `json:"Enable"`

			Timestamp int `json:"TimeStamp"`

			// 1...use values, 0...incomplete or outdated values
			Visible int `json:"Visible"`

			// 0...grid interconnection point (primary meter
			// 1...load (primary meter)
			// 3...external generator (secondary meters)(multiple)
			// 256-511 subloads (secondary meters)(unique)
			MeterLocationCurrent int `json:"Meter_Location_Current"`

			// absolute values
			CurrentACPhase1 float64 `json:"Current_AC_Phase_1"`
			CurrentACSum    float64 `json:"Current_AC_Sum"`

			// system specific view
			EnergyRealWattsACMinusAbsolute float64 `json:"EnergyReal_WAC_Minus_Absolute"`
			EnergyRealWattsACPlusAbsolute  float64 `json:"EnergyReal_WAC_Plus_Absolute"`

			// meter specific view
			EnergyRealWattsACPhase1Consumed float64 `json:"EnergyReal_WAC_Phase_1_Consumed"`
			EnergyRealWattsACPhase1Produced float64 `json:"EnergyReal_WAC_Phase_1_Produced"`
			EnergyRealWattsACSumConsumed    float64 `json:"EnergyReal_WAC_Sum_Consumed"`
			EnergyRealWattsACSumProduced    float64 `json:"EnergyReal_WAC_Sum_Produced"`

			// meter specific view
			EnergyReactiveVArACPhase1Consumed float64 `json:"EnergyReactive_VArAC_Phase_1_Consumed"`
			EnergyReactiveVArACPhase1Produced float64 `json:"EnergyReactive_VArAC_Phase_1_Produced"`
			EnergyReactiveVArACSumConsumed    float64 `json:"EnergyReactive_VArAC_Sum_Consumed"`
			EnergyReactiveVArACSumProduced    float64 `json:"EnergyReactive_VArAC_Sum_Produced"`

			FrequencyPhaseAverage float64 `json:"Frequency_Phase_Average"`

			PowerApparentSPhase1 float64 `json:"PowerApparent_S_Phase_1"`
			PowerApparentSSum    float64 `json:"PowerApparent_S_Sum"`

			PowerFactorPhase1 float64 `json:"PowerFactor_Phase_1"`
			PowerFactorSum    float64 `json:"PowerFactor_Sum"`

			PowerReactiveQPhase1 float64 `json:"PowerReactive_Q_Phase_1"`
			PowerReactiveQSum    float64 `json:"PowerReactive_Q_Sum"`

			PowerRealPPhase1 float64 `json:"PowerReal_P_Phase_1"`
			PowerRealPSum    float64 `json:"PowerReal_P_Sum"`

			VoltageACPhase1 float64 `json:"Voltage_AC_Phase_1"`
		} `json:"Data"`
	} `json:"Body"`
	Head head `json:"head"`
}

// MeterRealtime returns realtime meter data.
func (c Client) MeterRealtime(ctx context.Context, deviceID string) (points []*write.Point, err error) {
	var r meterRealtimeResponse

	q := url.Values{}

	q.Set("Scope", "Device")
	q.Set("DeviceId", deviceID)

	u := url.URL{Scheme: "http", Host: c.host, Path: "/solar_api/v1/GetMeterRealtimeData.cgi", RawQuery: q.Encode()}

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
		"current_ac_phase_1":                      r.Body.Data.CurrentACPhase1,
		"current_ac_sum":                          r.Body.Data.CurrentACSum,
		"energy_real_watts_ac_minus_absolute":     r.Body.Data.EnergyRealWattsACMinusAbsolute,
		"energy_real_watts_ac_plus_absolute":      r.Body.Data.EnergyRealWattsACPlusAbsolute,
		"energy_real_watts_ac_phase_1_consumed":   r.Body.Data.EnergyRealWattsACPhase1Consumed,
		"energy_real_watts_ac_phase_1_produced":   r.Body.Data.EnergyRealWattsACPhase1Produced,
		"energy_real_watts_ac_sum_consumed":       r.Body.Data.EnergyRealWattsACSumConsumed,
		"energy_real_watts_ac_sum_produced":       r.Body.Data.EnergyRealWattsACSumProduced,
		"energy_reactive_var_ac_phase_1_consumed": r.Body.Data.EnergyReactiveVArACPhase1Consumed,
		"energy_reactive_var_ac_phase_1_produced": r.Body.Data.EnergyReactiveVArACPhase1Produced,
		"energy_reactive_var_ac_sum_consumed":     r.Body.Data.EnergyReactiveVArACSumConsumed,
		"energy_reactive_var_ac_sum_produced":     r.Body.Data.EnergyReactiveVArACSumProduced,
		"frequency_phase_average":                 r.Body.Data.FrequencyPhaseAverage,
		"power_apparent_s_phase_1":                r.Body.Data.PowerApparentSPhase1,
		"power_apparent_s_sum":                    r.Body.Data.PowerApparentSSum,
		"power_factor_phase_1":                    r.Body.Data.PowerFactorPhase1,
		"power_factor_sum":                        r.Body.Data.PowerFactorSum,
		"power_reactive_q_phase_1":                r.Body.Data.PowerReactiveQPhase1,
		"power_reactive_q_sum":                    r.Body.Data.PowerReactiveQSum,
		"power_real_p_phase_1":                    r.Body.Data.PowerRealPPhase1,
		"power_real_p_sum":                        r.Body.Data.PowerRealPSum,
		"voltage_ac_phase_1":                      r.Body.Data.VoltageACPhase1,
	}

	return []*write.Point{
		influxdb2.NewPoint("fronius_meter", tags, values, r.Head.Timestamp),
	}, nil
}

// MeterArchive returns historical meter data.
func (c Client) MeterArchive(ctx context.Context, deviceID string, startDate time.Time, endDate time.Time) (points []*write.Point, err error) {
	q := defaultValues()

	q.Set("Scope", "Device")
	q.Set("SeriesType", "Detail")
	q.Set("HumanReadable", "False")
	q.Set("DeviceClass", "Meter")
	q.Set("DeviceId", deviceID)
	q.Set("StartDate", startDate.Format("2006-01-02"))
	q.Set("EndDate", endDate.Format("2006-01-02"))

	r, err := c.readArchive(ctx, q)
	if err != nil {
		return points, err
	}

	return generateArchivePoints(r, "meter_archive")
}
