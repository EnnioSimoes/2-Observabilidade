package weather

import (
	"testing"
)

func TestFormatTemperature(t *testing.T) {
	celsius := 28.0
	temp := formatTemparature(celsius)

	if temp.Temp_C != celsius {
		t.Errorf("Expected Temp_C to be %f, but got %f", celsius, temp.Temp_C)
	}

	expectedKelvin := celsius + 273
	if temp.Temp_K != expectedKelvin {
		t.Errorf("Expected Temp_K to be %f, but got %f", expectedKelvin, temp.Temp_K)
	}

	expectedFahrenheit := celsius*1.8 + 32
	if temp.Temp_F != expectedFahrenheit {
		t.Errorf("Expected Temp_F to be %f, but got %f", expectedFahrenheit, temp.Temp_F)
	}
}
