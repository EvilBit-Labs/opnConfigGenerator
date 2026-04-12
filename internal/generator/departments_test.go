package generator_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
)

func TestAllDepartments(t *testing.T) {
	expected := 20
	if got := len(generator.AllDepartments); got != expected {
		t.Errorf("AllDepartments length = %d, want %d", got, expected)
	}

	// Verify all expected departments are present
	expectedDepts := []generator.Department{
		generator.DeptSales, generator.DeptIT, generator.DeptHR, generator.DeptFinance,
		generator.DeptMarketing, generator.DeptOperations, generator.DeptEngineering,
		generator.DeptSupport, generator.DeptLegal, generator.DeptProcurement,
		generator.DeptSecurity, generator.DeptDevelopment, generator.DeptQA,
		generator.DeptResearch, generator.DeptTraining, generator.DeptManagement,
		generator.DeptAccounting, generator.DeptCustomerService, generator.DeptLogistics,
		generator.DeptProduction,
	}

	for _, dept := range expectedDepts {
		if !slices.Contains(generator.AllDepartments, dept) {
			t.Errorf("Department %s not found in AllDepartments", dept)
		}
	}
}

func TestDepartmentCount(t *testing.T) {
	expected := 20
	if got := generator.DepartmentCount(); got != expected {
		t.Errorf("DepartmentCount() = %d, want %d", got, expected)
	}
}

func TestDepartmentLeaseTime(t *testing.T) {
	tests := []struct {
		name       string
		dept       generator.Department
		expected   int
		category   string
	}{
		// Corporate departments (24h)
		{"IT corporate", generator.DeptIT, generator.LeaseTimeCorporate, "corporate"},
		{"Finance corporate", generator.DeptFinance, generator.LeaseTimeCorporate, "corporate"},
		{"Legal corporate", generator.DeptLegal, generator.LeaseTimeCorporate, "corporate"},
		{"Accounting corporate", generator.DeptAccounting, generator.LeaseTimeCorporate, "corporate"},
		{"Management corporate", generator.DeptManagement, generator.LeaseTimeCorporate, "corporate"},

		// Production departments (12h)
		{"Engineering production", generator.DeptEngineering, generator.LeaseTimeProduction, "production"},
		{"Development production", generator.DeptDevelopment, generator.LeaseTimeProduction, "production"},
		{"QA production", generator.DeptQA, generator.LeaseTimeProduction, "production"},
		{"Research production", generator.DeptResearch, generator.LeaseTimeProduction, "production"},

		// Dynamic departments (8h)
		{"Sales dynamic", generator.DeptSales, generator.LeaseTimeDynamic, "dynamic"},
		{"Marketing dynamic", generator.DeptMarketing, generator.LeaseTimeDynamic, "dynamic"},
		{"Customer Service dynamic", generator.DeptCustomerService, generator.LeaseTimeDynamic, "dynamic"},

		// Security department (6h)
		{"Security special", generator.DeptSecurity, generator.LeaseTimeSecurity, "security"},

		// High mobility departments (4h)
		{"HR mobility", generator.DeptHR, generator.LeaseTimeHighMobility, "mobility"},
		{"Logistics mobility", generator.DeptLogistics, generator.LeaseTimeHighMobility, "mobility"},
		{"Training mobility", generator.DeptTraining, generator.LeaseTimeHighMobility, "mobility"},
		{"Support mobility", generator.DeptSupport, generator.LeaseTimeHighMobility, "mobility"},
		{"Operations mobility", generator.DeptOperations, generator.LeaseTimeHighMobility, "mobility"},
		{"Procurement mobility", generator.DeptProcurement, generator.LeaseTimeHighMobility, "mobility"},
		{"Production mobility", generator.DeptProduction, generator.LeaseTimeHighMobility, "mobility"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dept.LeaseTime(); got != tt.expected {
				t.Errorf("%s.LeaseTime() = %d, want %d (%s category)", tt.dept, got, tt.expected, tt.category)
			}
		})
	}
}

func TestDepartmentStaticReservationCount(t *testing.T) {
	tests := []struct {
		name     string
		dept     generator.Department
		expected int
	}{
		{"IT has 3", generator.DeptIT, 3},
		{"Engineering has 2", generator.DeptEngineering, 2},
		{"Security has 2", generator.DeptSecurity, 2},
		{"Development has 2", generator.DeptDevelopment, 2},
		{"QA has 2", generator.DeptQA, 2},
		{"Management has 1", generator.DeptManagement, 1},
		{"Finance has 1", generator.DeptFinance, 1},
		{"Sales has 0", generator.DeptSales, 0},
		{"HR has 0", generator.DeptHR, 0},
		{"Marketing has 0", generator.DeptMarketing, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dept.StaticReservationCount(); got != tt.expected {
				t.Errorf("%s.StaticReservationCount() = %d, want %d", tt.dept, got, tt.expected)
			}
		})
	}
}

func TestDepartmentStaticReservationDevices(t *testing.T) {
	tests := []struct {
		name     string
		dept     generator.Department
		expected []string
	}{
		{"IT devices", generator.DeptIT, []string{"printer", "nas", "server"}},
		{"Engineering devices", generator.DeptEngineering, []string{"build-server", "ci-runner"}},
		{"Security devices", generator.DeptSecurity, []string{"camera", "access-controller"}},
		{"Development devices", generator.DeptDevelopment, []string{"dev-server", "staging"}},
		{"QA devices", generator.DeptQA, []string{"test-server", "qa-runner"}},
		{"Management devices", generator.DeptManagement, []string{"exec-printer"}},
		{"Finance devices", generator.DeptFinance, []string{"finance-printer"}},
		{"Sales no devices", generator.DeptSales, nil},
		{"HR no devices", generator.DeptHR, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dept.StaticReservationDevices()
			if !slicesEqual(got, tt.expected) {
				t.Errorf("%s.StaticReservationDevices() = %v, want %v", tt.dept, got, tt.expected)
			}
		})
	}
}

func TestRandomDepartment(t *testing.T) {
	// Test with seeded RNG for deterministic output
	rng := rand.New(rand.NewPCG(42, 1337))

	// Generate multiple departments and verify they're all valid
	seenDepts := make(map[generator.Department]bool)
	for i := 0; i < 100; i++ {
		dept := generator.RandomDepartment(rng)
		if !slices.Contains(generator.AllDepartments, dept) {
			t.Errorf("RandomDepartment() returned invalid department: %s", dept)
		}
		seenDepts[dept] = true
	}

	// With 100 iterations, we should see multiple different departments
	if len(seenDepts) < 5 {
		t.Errorf("RandomDepartment() should produce variety, got only %d unique departments", len(seenDepts))
	}
}

func TestRandomDepartmentDeterministic(t *testing.T) {
	// Test that the same seed produces the same sequence
	seed1 := uint64(12345)
	seed2 := uint64(67890)

	rng1a := rand.New(rand.NewPCG(seed1, seed2))
	rng1b := rand.New(rand.NewPCG(seed1, seed2))

	dept1a := generator.RandomDepartment(rng1a)
	dept1b := generator.RandomDepartment(rng1b)

	if dept1a != dept1b {
		t.Errorf("Same seed should produce same department: got %s and %s", dept1a, dept1b)
	}

	// Test sequence of departments
	rng2a := rand.New(rand.NewPCG(seed1, seed2))
	rng2b := rand.New(rand.NewPCG(seed1, seed2))

	for i := 0; i < 10; i++ {
		deptA := generator.RandomDepartment(rng2a)
		deptB := generator.RandomDepartment(rng2b)
		if deptA != deptB {
			t.Errorf("Same seed should produce same sequence at position %d: got %s and %s", i, deptA, deptB)
		}
	}
}

func TestLeaseTimeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"Corporate 24h", generator.LeaseTimeCorporate, 86400},
		{"Production 12h", generator.LeaseTimeProduction, 43200},
		{"Dynamic 8h", generator.LeaseTimeDynamic, 28800},
		{"Security 6h", generator.LeaseTimeSecurity, 21600},
		{"High Mobility 4h", generator.LeaseTimeHighMobility, 14400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestAllDepartmentsHaveValidLeaseTime(t *testing.T) {
	validTimes := map[int]bool{
		generator.LeaseTimeCorporate:    true,
		generator.LeaseTimeProduction:   true,
		generator.LeaseTimeDynamic:      true,
		generator.LeaseTimeSecurity:     true,
		generator.LeaseTimeHighMobility: true,
	}

	for _, dept := range generator.AllDepartments {
		leaseTime := dept.LeaseTime()
		if !validTimes[leaseTime] {
			t.Errorf("Department %s has invalid lease time %d", dept, leaseTime)
		}
	}
}

func TestAllDepartmentsHaveValidReservationCounts(t *testing.T) {
	for _, dept := range generator.AllDepartments {
		count := dept.StaticReservationCount()
		devices := dept.StaticReservationDevices()

		// If count > 0, should have devices
		if count > 0 && len(devices) != count {
			t.Errorf("Department %s has count %d but %d devices", dept, count, len(devices))
		}

		// If count == 0, should have no devices
		if count == 0 && devices != nil {
			t.Errorf("Department %s has count 0 but devices %v", dept, devices)
		}

		// Count should not exceed 3
		if count > 3 {
			t.Errorf("Department %s has excessive static reservation count: %d", dept, count)
		}
	}
}

// Helper function to compare slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}