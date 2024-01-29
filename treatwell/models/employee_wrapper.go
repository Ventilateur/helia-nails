package models

type EmployeeWrapper struct {
	Info         Employee
	WorkingHours []WorkingHour
}
