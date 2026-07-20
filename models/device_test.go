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

func TestThermostatModeSupportsAdmittedFloatValues(t *testing.T) {
	minValue := float64(-1.0)
	maxValue := float64(2.0)
	step := float64(1.0)
	feature := Feature{
		Type:   Sensor,
		Name:   "mode",
		Enable: true,
		Order:  4,
		Unit:   "-",
		Spec: Spec{
			Format: Float,
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
	if decoded.Spec.Format != Float || decoded.Spec.Min == nil || *decoded.Spec.Min != -1.0 ||
		decoded.Spec.Max == nil || *decoded.Spec.Max != 2.0 ||
		decoded.Spec.Step == nil || *decoded.Spec.Step != 1.0 {
		t.Fatalf("thermostat mode spec was not preserved: %#v", decoded.Spec)
	}

	admittedValues := []float32{-1.0, 0.0, 1.0, 2.0}
	for _, value := range admittedValues {
		state := DeviceFeatureState{
			FeatureUUID: "mode-feature",
			Type:        Sensor,
			Name:        "mode",
			Value:       value,
		}
		if err = validator.New().Struct(state); err != nil {
			t.Fatalf("thermostat mode value %.1f should be valid: %v", value, err)
		}
	}
}
