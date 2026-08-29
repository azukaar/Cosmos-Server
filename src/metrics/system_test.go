package metrics

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/common"
	"github.com/shirou/gopsutil/v3/host"
)

func iwlwifiENODATAWarning() error {
	warns := &host.Warnings{}
	warns.Add(errors.New("open /sys/class/hwmon/hwmon1/temp1_input: no data available"))
	return warns
}

func TestUsableTemperaturesKeepsValidSensorsWhenIwlwifiReturnsENODATA(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "acpitz", Temperature: 27.8},
		{SensorKey: "coretemp_core_0", Temperature: 45},
		{SensorKey: "coretemp_package_id_0", Temperature: 47},
	}

	got := usableTemperatures(temps, iwlwifiENODATAWarning())
	if len(got) != 3 {
		t.Fatalf("usableTemperatures() returned %d sensors, want 3 valid CPU/ACPI readings despite iwlwifi ENODATA: %v", len(got), got)
	}

	wantKeys := map[string]bool{
		"acpitz":                true,
		"coretemp_core_0":       true,
		"coretemp_package_id_0": true,
	}
	for _, temp := range got {
		if !wantKeys[temp.SensorKey] {
			t.Errorf("unexpected sensor %q", temp.SensorKey)
		}
		delete(wantKeys, temp.SensorKey)
		if temp.Temperature <= 0 {
			t.Errorf("sensor %q has non-positive temperature %v", temp.SensorKey, temp.Temperature)
		}
	}
	for key := range wantKeys {
		t.Errorf("missing sensor %q", key)
	}
}

func TestUsableTemperaturesSkipsZeroReadings(t *testing.T) {
	temps := []host.TemperatureStat{
		{SensorKey: "coretemp_core_0", Temperature: 45},
		{SensorKey: "iwlwifi_1", Temperature: 0},
	}

	got := usableTemperatures(temps, nil)
	if len(got) != 1 || got[0].SensorKey != "coretemp_core_0" {
		t.Fatalf("usableTemperatures() = %v, want only coretemp_core_0", got)
	}
}

func TestUsableTemperaturesHardFailureWithNoReadings(t *testing.T) {
	got := usableTemperatures(nil, errors.New("permission denied"))
	if len(got) != 0 {
		t.Fatalf("usableTemperatures() = %v, want none on hard failure", got)
	}
}

func TestUsableTemperaturesFromIwlwifiSysfsWarning(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("hwmon sysfs is linux-only")
	}

	sys := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(sys, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("class/hwmon/hwmon0/name", "coretemp\n")
	write("class/hwmon/hwmon0/temp1_label", "Core 0\n")
	write("class/hwmon/hwmon0/temp1_input", "45000\n")
	write("class/hwmon/hwmon1/name", "iwlwifi_1\n")
	write("class/hwmon/hwmon1/temp1_input", "")

	ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{
		common.HostSysEnvKey: sys,
	})
	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err == nil {
		t.Fatal("expected gopsutil warnings from iwlwifi")
	}
	if err.Error() != "Number of warnings: 1" {
		t.Fatalf("err = %q, want Number of warnings: 1", err.Error())
	}

	got := usableTemperatures(temps, err)
	if len(got) != 1 || got[0].SensorKey != "coretemp_core_0" || int(got[0].Temperature) != 45 {
		t.Fatalf("usableTemperatures() = %v, want coretemp_core_0 at 45C", got)
	}
}
