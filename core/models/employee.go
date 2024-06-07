package models

type Employee struct {
	Name      string            `yaml:"name"`
	Treatwell TreatwellEmployee `yaml:"treatwell"`
	Planity   PlanityEmployee   `yaml:"planity"`
	Classpass ClasspassEmployee `yaml:"classpass"`
}

type TreatwellEmployee struct {
	Id int `yaml:"id"`
}

type PlanityEmployee struct {
	Id string `yaml:"id"`
}

type ClasspassEmployee struct {
	GoogleCalendarId string `yaml:"googleCalendarId"`
}
