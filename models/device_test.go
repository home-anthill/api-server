package models

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestThermostatSupportsOptionalModeFeature(t *testing.T) {
	temperatureFeature := Feature{
		UUID:   "temperature-feature",
		Type:   Sensor,
		Name:   "temperature",
		Enable: true,
		Order:  1,
		Unit:   "°C",
	}
	controllerFeature := Feature{
		UUID:   "controller-feature",
		Type:   Controller,
		Name:   "thermostat",
		Enable: true,
		Order:  2,
		Unit:   "-",
	}
	modeFeature := Feature{
		UUID:   "mode-feature",
		Type:   Sensor,
		Name:   "mode",
		Enable: true,
		Order:  3,
		Unit:   "-",
	}

	tests := []struct {
		name        string
		features    []Feature
		expectsMode bool
	}{
		{
			name:        "thermostat with mode",
			features:    []Feature{temperatureFeature, controllerFeature, modeFeature},
			expectsMode: true,
		},
		{
			name:        "thermostat without mode",
			features:    []Feature{temperatureFeature, controllerFeature},
			expectsMode: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			thermostat := Device{
				UUID:         "thermostat-device",
				Manufacturer: "test",
				Model:        "thermostat",
				Features:     test.features,
			}

			encoded, err := json.Marshal(thermostat)
			if err != nil {
				t.Fatalf("marshal thermostat: %v", err)
			}

			var decoded Device
			if err = json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("unmarshal thermostat: %v", err)
			}
			if len(decoded.Features) != len(test.features) {
				t.Fatalf("expected %d features, got %d", len(test.features), len(decoded.Features))
			}

			hasMode := false
			for _, feature := range decoded.Features {
				if feature.Name == "mode" {
					hasMode = true
					break
				}
			}
			if hasMode != test.expectsMode {
				t.Fatalf("expected mode presence to be %t, got %t", test.expectsMode, hasMode)
			}
		})
	}
}

func TestThermostatModeSupportsNegativeFaultValue(t *testing.T) {
	minValue := float64(-1)
	maxValue := float64(2)
	step := float64(1)
	feature := Feature{
		Type:   Sensor,
		Name:   "mode",
		Enable: true,
		Order:  4,
		Unit:   "-",
		Spec: Spec{
			Format: Int,
			Min:    &minValue,
			Max:    &maxValue,
			Step:   &step,
		},
	}

	encoded, err := json.Marshal(feature)
	if err != nil {
		t.Fatalf("marshal thermostat mode feature: %v", err)
	}

	var decoded Feature
	if err = json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal thermostat mode feature: %v", err)
	}
	if decoded.Spec.Format != Int || decoded.Spec.Min == nil || *decoded.Spec.Min != -1 {
		t.Fatalf("thermostat mode spec was not preserved: %#v", decoded.Spec)
	}

	state := DeviceFeatureState{
		FeatureUUID: "mode-feature",
		Type:        Sensor,
		Name:        "mode",
		Value:       -1,
	}
	if err = validator.New().Struct(state); err != nil {
		t.Fatalf("negative thermostat fault mode should be valid: %v", err)
	}
}
