package main

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type head struct {
	Timestamp        time.Time              `json:"Timestamp"`
	RequestArguments map[string]interface{} `json:"RequestArguments"`
	Status           struct {
		Code        int    `json:"Code"`
		Reason      string `json:"Reason"`
		UserMessage string `json:"UserMessage"`
		ErrorDetail struct {
			Nodes []interface{} `json:"Nodes"`
		} `json:"ErrorDetail"`
	} `json:"Status"`
}

type archiveData struct {
	Unit    string              `json:"Unit"`
	Comment string              `json:"_comment"`
	Values  map[string]*float64 `json:"Values"`
}

type archiveResponse struct {
	Body struct {
		Data map[string]struct {
			DeviceType int
			NodeType   int
			Start      time.Time `json:"Start"`
			End        time.Time `json:"End"`
			Data       struct {
				TimeSpan                        archiveData `json:"TimeSpanInSec"`
				EnergyRealWACSumProduced        archiveData `json:"EnergyReal_WAC_Sum_Produced"`
				EnergyRealWACSumConsumed        archiveData `json:"EnergyReal_WAC_Sum_Consumed"`
				CurrentDCString1                archiveData `json:"Current_DC_String_1"`
				CurrentDCString2                archiveData `json:"Current_DC_String_2"`
				VoltageDCString1                archiveData `json:"Voltage_DC_String_1"`
				VoltageDCString2                archiveData `json:"Voltage_DC_String_2"`
				TemperaturePowerStage           archiveData `json:"Temperature_Powerstage"`
				VoltageACPhase1                 archiveData `json:"Voltage_AC_Phase_1"`
				VoltageACPhase2                 archiveData `json:"Voltage_AC_Phase_2"`
				VoltageACPhase3                 archiveData `json:"Voltage_AC_Phase_3"`
				CurrentACPhase1                 archiveData `json:"Current_AC_Phase_1"`
				CurrentACPhase2                 archiveData `json:"Current_AC_Phase_2"`
				CurrentACPhase3                 archiveData `json:"Current_AC_Phase_3"`
				PowerRealPACSum                 archiveData `json:"PowerReal_PAC_Sum"`
				EnergyRealWACMinusAbsolute      archiveData `json:"EnergyReal_WAC_Minus_Absolute"`
				EnergyRealWACPlusAbsolute       archiveData `json:"EnergyReal_WAC_Plus_Absolute"`
				MeterLocationCurrent            archiveData `json:"Meter_Location_Current"`
				TemperatureChannel1             archiveData `json:"Temperature_Channel_1"`
				TemperatureChannel2             archiveData `json:"Temperature_Channel_2"`
				DigitalChannel1                 archiveData `json:"Digital_Channel_1"`
				DigitalChannel2                 archiveData `json:"Digital_Channel_2"`
				Radiation                       archiveData `json:"Radiation"`
				DigitalPowerManagementRelayOut1 archiveData `json:"Digital_PowerManagementRelay_Out_1"`
				DigitalPowerManagementRelayOut2 archiveData `json:"Digital_PowerManagementRelay_Out_2"`
				DigitalPowerManagementRelayOut3 archiveData `json:"Digital_PowerManagementRelay_Out_3"`
				DigitalPowerManagementRelayOut4 archiveData `json:"Digital_PowerManagementRelay_Out_4"`
				HybridOperatingState            archiveData `json:"Hybrid_Operating_State"`
			} `json:"Data"`
		} `json:"Data"`
	} `json:"Body"`
	Head head `json:"head"`
}

type timeValues map[time.Time]map[string]interface{}

// ErrStatusNotOk is when the HTTP response code is not 200.
var ErrStatusNotOk = errors.New("status not OK")

func defaultValues() url.Values {
	q := url.Values{}

	q.Add("Channel", "TimeSpanInSec")
	q.Add("Channel", "EnergyReal_WAC_Sum_Produced")
	q.Add("Channel", "EnergyReal_WAC_Sum_Consumed")
	q.Add("Channel", "Current_DC_String_1")
	q.Add("Channel", "Current_DC_String_2")
	q.Add("Channel", "Voltage_DC_String_1")
	q.Add("Channel", "Voltage_DC_String_2")
	q.Add("Channel", "Temperature_Powerstage")
	q.Add("Channel", "Voltage_AC_Phase_1")
	q.Add("Channel", "Voltage_AC_Phase_2")
	q.Add("Channel", "Voltage_AC_Phase_3")
	q.Add("Channel", "Current_AC_Phase_1")
	q.Add("Channel", "Current_AC_Phase_2")
	q.Add("Channel", "Current_AC_Phase_3")
	q.Add("Channel", "PowerReal_PAC_Sum")
	q.Add("Channel", "EnergyReal_WAC_Minus_Absolute")
	q.Add("Channel", "EnergyReal_WAC_Plus_Absolute")
	q.Add("Channel", "Meter_Location_Current")
	q.Add("Channel", "Temperature_Channel_1")
	q.Add("Channel", "Temperature_Channel_2")
	q.Add("Channel", "Digital_Channel_1")
	q.Add("Channel", "Digital_Channel_2")
	q.Add("Channel", "Radiation")
	q.Add("Channel", "Digital_PowerManagementRelay_Out_1")
	q.Add("Channel", "Digital_PowerManagementRelay_Out_2")
	q.Add("Channel", "Digital_PowerManagementRelay_Out_3")
	q.Add("Channel", "Digital_PowerManagementRelay_Out_4")
	q.Add("Channel", "Hybrid_Operating_State")

	return q
}

func generateArchivePoints(r archiveResponse, measurement string) (points []*write.Point, err error) {
	for deviceID, deviceData := range r.Body.Data {
		tags := map[string]string{
			"device_id":   deviceID,
			"device_type": strconv.Itoa(deviceData.DeviceType),
			"node_type":   strconv.Itoa(deviceData.NodeType),
		}

		timeValues := make(timeValues)

		assignments := map[string]*archiveData{
			"time_span":                            &deviceData.Data.TimeSpan,
			"energy_real_wac_sum_produced":         &deviceData.Data.EnergyRealWACSumProduced,
			"energy_real_wac_sum_consumed":         &deviceData.Data.EnergyRealWACSumConsumed,
			"current_dc_string_1":                  &deviceData.Data.CurrentDCString1,
			"current_dc_string_2":                  &deviceData.Data.CurrentDCString2,
			"voltage_dc_string_1":                  &deviceData.Data.VoltageDCString1,
			"voltage_dc_string_2":                  &deviceData.Data.VoltageDCString2,
			"temperature_power_stage":              &deviceData.Data.TemperaturePowerStage,
			"voltage_ac_phase_1":                   &deviceData.Data.VoltageACPhase1,
			"voltage_ac_phase_2":                   &deviceData.Data.VoltageACPhase2,
			"voltage_ac_phase_3":                   &deviceData.Data.VoltageACPhase3,
			"current_ac_phase_1":                   &deviceData.Data.CurrentACPhase1,
			"current_ac_phase_2":                   &deviceData.Data.CurrentACPhase2,
			"current_ac_phase_3":                   &deviceData.Data.CurrentACPhase3,
			"power_real_pac_sum":                   &deviceData.Data.PowerRealPACSum,
			"energy_real_wac_minus_absolute":       &deviceData.Data.EnergyRealWACMinusAbsolute,
			"energy_real_wac_plus_absolute":        &deviceData.Data.EnergyRealWACPlusAbsolute,
			"meter_location_current":               &deviceData.Data.MeterLocationCurrent,
			"temperature_channel_1":                &deviceData.Data.TemperatureChannel1,
			"temperature_channel_2":                &deviceData.Data.TemperatureChannel2,
			"digital_channel_1":                    &deviceData.Data.DigitalChannel1,
			"digital_channel_2":                    &deviceData.Data.DigitalChannel2,
			"radiation":                            &deviceData.Data.Radiation,
			"digital_power_management_relay_out_1": &deviceData.Data.DigitalPowerManagementRelayOut1,
			"digital_power_management_relay_out_2": &deviceData.Data.DigitalPowerManagementRelayOut2,
			"digital_power_management_relay_out_3": &deviceData.Data.DigitalPowerManagementRelayOut3,
			"digital_power_management_relay_out_4": &deviceData.Data.DigitalPowerManagementRelayOut4,
			"hybrid_operating_state":               &deviceData.Data.HybridOperatingState,
		}

		for key, archiveData := range assignments {
			if timeValues, err = ingest(timeValues, deviceData.Start, archiveData, key); err != nil {
				return points, err
			}
		}

		for timestamp, values := range timeValues {
			point := influxdb2.NewPoint(measurement, tags, values, timestamp)

			points = append(points, point)
		}
	}

	return points, nil
}

func ingest(timeValues timeValues, startTime time.Time, archiveData *archiveData, key string) (timeValues, error) {
	for offsetStr, value := range archiveData.Values {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return timeValues, err
		}

		if value != nil {
			timestamp := startTime.Add(time.Duration(offset) * time.Second)

			_, ok := timeValues[timestamp]
			if !ok {
				timeValues[timestamp] = make(map[string]interface{})
			}

			timeValues[timestamp][key] = *value
		}
	}

	return timeValues, nil
}
