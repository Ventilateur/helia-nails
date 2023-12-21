package models

type Employees struct {
	Employees []Employee `json:"employees"`
}

type Employee struct {
	Active                  bool        `json:"active"`
	Id                      int         `json:"id"`
	Name                    string      `json:"name"`
	CanLogin                bool        `json:"canLogin"`
	DoesAllOffers           bool        `json:"doesAllOffers"`
	EmployeeCategoryId      interface{} `json:"employeeCategoryId"`
	EmployeeOffers          []int       `json:"employeeOffers"`
	ExternalCalendarUri     interface{} `json:"externalCalendarUri"`
	ImageId                 interface{} `json:"imageId"`
	Image                   interface{} `json:"image"`
	JobTitle                *string     `json:"jobTitle"`
	Notes                   *string     `json:"notes"`
	Phone                   interface{} `json:"phone"`
	TakesAppointments       bool        `json:"takesAppointments"`
	SupplierBound           bool        `json:"supplierBound"`
	EmailAddress            *string     `json:"emailAddress"`
	Permissions             []string    `json:"permissions"`
	Role                    *string     `json:"role"`
	LinkedExternalSalonName interface{} `json:"linkedExternalSalonName"`
	EmploymentStatus        interface{} `json:"employmentStatus"`
}
