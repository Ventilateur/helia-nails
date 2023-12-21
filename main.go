package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/Ventilateur/helia-nails/treatwell"
)

func main() {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Timeout: 1 * time.Minute,
		Jar:     cookieJar,
	}

	tw, err := treatwell.New(client, "428563")
	if err != nil {
		panic(err)
	}

	err = tw.Login("", "")
	if err != nil {
		panic(err)
	}

	calendar, err := tw.GetCalendar(
		time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(calendar, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	//err = tw.BookAnonymously(twmodels.Appointment{
	//	AppointmentDate:          "2023-12-21",
	//	Platform:                 "DESKTOP",
	//	EmployeeId:               461944,
	//	StartTime:                "14:00",
	//	PaymentProtected:         false,
	//	PaymentProtectionApplied: false,
	//	Skus: []twmodels.Sku{
	//		{
	//			SkuId: 8333866,
	//		},
	//	},
	//	ServiceId: 4435253,
	//}, "Test client")
	//
	//if err != nil {
	//	panic(err)
	//}
}
