package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/Ventilateur/helia-nails/core"
	google_calendar "github.com/Ventilateur/helia-nails/google-calendar"
	"github.com/Ventilateur/helia-nails/treatwell"
)

func main() {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	u, err := url.Parse("https://treatwell.fr")
	if err != nil {
		panic(err)
	}

	cookieJar.SetCookies(u, []*http.Cookie{})

	client := &http.Client{
		Timeout: 1 * time.Minute,
		Jar:     cookieJar,
	}

	tw, err := treatwell.New(client, "428563")
	if err != nil {
		panic(err)
	}

	//err = tw.Login("helia.universea@gmail.com", "Hoangthanhlinh2@treatwell")
	//if err != nil {
	//	panic(err)
	//}

	ga, err := google_calendar.New(
		context.Background(),
		"calendar-sync@helia-nails.iam.gserviceaccount.com",
		[]byte(),
	)

	_ = core.New(tw, ga)

	//err = sync.TreatwellToGoogleCalendar(
	//	google_calendar.ClassPassCalendarID,
	//	time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
	//	time.Date(2024, 1, 14, 0, 0, 0, 0, time.Local),
	//	models.SourceClassPass,
	//)
	//if err != nil {
	//	panic(err)
	//}

	events, err := ga.List(
		google_calendar.ClassPassCalendarID,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		time.Date(2024, 1, 14, 0, 0, 0, 0, time.Local),
	)
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		fmt.Printf("%+v\n", event.Id)
	}
}
