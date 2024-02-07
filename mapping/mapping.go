package mapping

const (
	employeeTee     = 461942
	employeeJade    = 461945
	employeeChloe   = 461946
	employeeMinette = 461944
)

const (
	EmployeeNameJade    = "Jade"
	EmployeeNameChloe   = "Chloé"
	EmployeeNameMinette = "Minette"
)

var (
	AllEmployees                = []string{EmployeeNameJade, EmployeeNameChloe, EmployeeNameMinette}
	EmployeeGoogleCalendarIDMap = map[string]string{
		EmployeeNameJade:    "c1473969314d968f7068537f5ddb96e9a316a1788c8512255139f44b8000ad0c@group.calendar.google.com",
		EmployeeNameMinette: "bf675e61950db2cc249c3d57b687654988273604cc4f400da558889b477825d0@group.calendar.google.com",
		EmployeeNameChloe:   "b72bdf75141c2d7489e6da0ceeb7eed069c36ace2ff9374235b9440b817e7e8f@group.calendar.google.com",
	}
	EmployeeTreatwellIDMap = map[string]int{
		EmployeeNameJade:    461945,
		EmployeeNameChloe:   461946,
		EmployeeNameMinette: 461944,
	}
	CalendarIDToEmployeeMap = func() map[string]string {
		m := map[string]string{}
		for employee, calendarID := range EmployeeGoogleCalendarIDMap {
			m[calendarID] = employee
		}
		return m
	}()
)

type TreatwellOffer struct {
	OfferID       int
	SkuID         int
	PossibleNames []string
}

var TreatwellOffers = []TreatwellOffer{
	{
		PossibleNames: []string{
			"Beauté des mains avec pose de vernis semi-permanent",
		},
		OfferID: 4435453,
		SkuID:   8334380,
	},
	{
		PossibleNames: []string{
			"Beauté des pieds avec pose de vernis semi-permanent",
		},
		OfferID: 4435255,
		SkuID:   8403680,
	},
	{
		PossibleNames: []string{
			"Beauté des mains et pieds avec pose de vernis semi-permanent",
		},
		OfferID: 4470656,
		SkuID:   8403758,
	},
	{
		PossibleNames: []string{
			"Extension de cils pose mixte / volume russe / cil à cil",
		},
		OfferID: 4435348,
		SkuID:   8403725,
	},
	{
		PossibleNames: []string{
			"Rehaussement de cils",
		},
		OfferID: 4435345,
		SkuID:   8334167,
	},
}
