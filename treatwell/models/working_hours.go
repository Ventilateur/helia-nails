package models

type EmployeesWorkingHours struct {
	EmployeesWorkingHours []EmployeeWorkingHour `json:"employeesWorkingHours"`
}

type EmployeeWorkingHour struct {
	EmployeeID   int           `json:"employeeId"`
	EmployeeName string        `json:"employeeName"`
	WorkingHours []WorkingHour `json:"workingHours"`
}

type WorkingHour struct {
	Date      string     `json:"date"`
	Type      string     `json:"type"`
	TimeSlots []TimeSlot `json:"timeSlots"`
	Schedule  Schedule   `json:"schedule,omitempty"`
}

type TimeSlot struct {
	TimeFrom string `json:"timeFrom"`
	TimeTo   string `json:"timeTo"`
}

type Schedule struct {
	ScheduleId int    `json:"scheduleId"`
	ValidFrom  string `json:"validFrom"`
}
