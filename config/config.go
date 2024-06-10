package config

import (
	"strconv"
	"strings"

	"github.com/Ventilateur/helia-nails/core/models"
)

type Config struct {
	OpenTime  string            `yaml:"openTime"`
	CloseTime string            `yaml:"closeTime"`
	Services  []models.Service  `yaml:"services"`
	Employees []models.Employee `yaml:"employees"`
	Treatwell TreatwellConfig   `yaml:"treatwell"`
	Planity   PlanityConfig     `yaml:"planity"`
	Classpass ClasspassConfig   `yaml:"classpass"`
}

type TreatwellConfig struct {
	VenueId string `yaml:"venueId"`
	UserId  string `yaml:"userId"`
	ATKT    string `yaml:"atkt"`
	ITKT    string `yaml:"itkt"`
}

type PlanityConfig struct {
	ApiKey       string `yaml:"apiKey"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	WebsocketUrl string `yaml:"websocketUrl"`
	AccessToken  string `yaml:"-"`
}

type ClasspassConfig struct {
	GoogleKey   string `yaml:"googleKey"`
	GoogleEmail string `yaml:"googleEmail"`
}

func (c Config) GetEmployee(source models.Source, id string) models.Employee {
	for _, employee := range c.Employees {
		switch source {
		case models.SourceTreatwell:
			if id == strconv.Itoa(employee.Treatwell.Id) {
				return employee
			}
		case models.SourcePlanity:
			if id == employee.Planity.Id {
				return employee
			}
		case models.SourceClassPass:
			if id == employee.Classpass.GoogleCalendarId {
				return employee
			}
		default:
			return models.Employee{}
		}
	}

	return models.Employee{}
}

func (c Config) GetService(source models.Source, id string, subId string) models.Service {
	for _, service := range c.Services {
		switch source {
		case models.SourceTreatwell:
			if id == strconv.Itoa(service.Treatwell.OfferId) && subId == strconv.Itoa(service.Treatwell.SkuId) {
				return service
			}
		case models.SourcePlanity:
			if id == service.Planity.Id {
				return service
			}
		case models.SourceClassPass:
			for _, name := range service.Classpass.PossibleNames {
				if strings.Contains(id, name) {
					return service
				}
			}
		default:
			return models.Service{}
		}
	}

	return models.Service{}
}

func (c Config) GetServiceFromPlanity(id string) models.Service {
	for _, service := range c.Services {
		if id == service.Planity.Id {
			return service
		}
	}

	return models.Service{}
}
