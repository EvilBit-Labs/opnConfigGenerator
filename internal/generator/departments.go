// Package generator produces realistic fake OPNsense configuration data.
package generator

import "math/rand/v2"

// Department represents a network department for VLAN naming.
type Department string

// Department constants for all supported network departments.
const (
	DeptSales           Department = "Sales"
	DeptIT              Department = "IT"
	DeptHR              Department = "HR"
	DeptFinance         Department = "Finance"
	DeptMarketing       Department = "Marketing"
	DeptOperations      Department = "Operations"
	DeptEngineering     Department = "Engineering"
	DeptSupport         Department = "Support"
	DeptLegal           Department = "Legal"
	DeptProcurement     Department = "Procurement"
	DeptSecurity        Department = "Security"
	DeptDevelopment     Department = "Development"
	DeptQA              Department = "QA"
	DeptResearch        Department = "Research"
	DeptTraining        Department = "Training"
	DeptManagement      Department = "Management"
	DeptAccounting      Department = "Accounting"
	DeptCustomerService Department = "Customer Service"
	DeptLogistics       Department = "Logistics"
	DeptProduction      Department = "Production"
)

// AllDepartments is the complete list of departments.
var AllDepartments = []Department{
	DeptSales, DeptIT, DeptHR, DeptFinance, DeptMarketing,
	DeptOperations, DeptEngineering, DeptSupport, DeptLegal, DeptProcurement,
	DeptSecurity, DeptDevelopment, DeptQA, DeptResearch, DeptTraining,
	DeptManagement, DeptAccounting, DeptCustomerService, DeptLogistics, DeptProduction,
}

// DHCP lease times in seconds by department category.
const (
	LeaseTimeCorporate    = 86400 // 24h - IT, Finance, Legal, Accounting, Management
	LeaseTimeProduction   = 43200 // 12h - Engineering, Development, QA, Research
	LeaseTimeDynamic      = 28800 // 8h  - Sales, Marketing, Customer Service
	LeaseTimeHighMobility = 14400 // 4h  - HR, Logistics, Training, Support, Operations, Procurement, Production
	LeaseTimeSecurity     = 21600 // 6h  - Security
)

// LeaseTime returns the DHCP lease time for a department.
func (d Department) LeaseTime() int {
	switch d {
	case DeptIT, DeptFinance, DeptLegal, DeptAccounting, DeptManagement:
		return LeaseTimeCorporate
	case DeptEngineering, DeptDevelopment, DeptQA, DeptResearch:
		return LeaseTimeProduction
	case DeptSales, DeptMarketing, DeptCustomerService:
		return LeaseTimeDynamic
	case DeptSecurity:
		return LeaseTimeSecurity
	default:
		return LeaseTimeHighMobility
	}
}

// Static DHCP reservation counts per department type.
const (
	reservationsHigh   = 3 // IT: printer, NAS, server
	reservationsMedium = 2 // Engineering/Security: build-server or camera
	reservationsLow    = 1 // Management/Finance: single device
)

// StaticReservationCount returns how many static DHCP reservations a department gets.
func (d Department) StaticReservationCount() int {
	switch d {
	case DeptIT:
		return reservationsHigh
	case DeptEngineering, DeptSecurity:
		return reservationsMedium
	case DeptDevelopment, DeptQA:
		return reservationsMedium
	case DeptManagement, DeptFinance:
		return reservationsLow
	default:
		return 0
	}
}

// StaticReservationDevices returns device name templates for a department.
func (d Department) StaticReservationDevices() []string {
	switch d {
	case DeptIT:
		return []string{"printer", "nas", "server"}
	case DeptEngineering:
		return []string{"build-server", "ci-runner"}
	case DeptSecurity:
		return []string{"camera", "access-controller"}
	case DeptDevelopment:
		return []string{"dev-server", "staging"}
	case DeptQA:
		return []string{"test-server", "qa-runner"}
	case DeptManagement:
		return []string{"exec-printer"}
	case DeptFinance:
		return []string{"finance-printer"}
	default:
		return nil
	}
}

// RandomDepartment returns a random department using the provided RNG.
func RandomDepartment(rng *rand.Rand) Department {
	return AllDepartments[rng.IntN(len(AllDepartments))]
}

// DepartmentCount returns the total number of departments.
func DepartmentCount() int {
	return len(AllDepartments)
}
