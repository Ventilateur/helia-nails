package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/Ventilateur/helia-nails/core"
	"github.com/Ventilateur/helia-nails/core/models"
	google_calendar "github.com/Ventilateur/helia-nails/googlecalendar"
	"github.com/Ventilateur/helia-nails/treatwell"
	"github.com/Ventilateur/helia-nails/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/cobra"
)

const (
	googleKey = "/google/key"
	twATKT    = "/treatwell/ATKT"
	twITKT    = "/treatwell/ITKT"
	twUserID  = "/treatwell/tw_user_id"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use: "sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		params, err := getParams(ctx, googleKey, twATKT, twITKT, twUserID)
		if err != nil {
			return err
		}

		cookieJar, err := cookiejar.New(nil)
		if err != nil {
			return fmt.Errorf("failed to create cookie jar: %w", err)
		}

		u, err := url.Parse("https://treatwell.fr")
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}

		cookieJar.SetCookies(u, []*http.Cookie{
			{
				Name:   "ATKT",
				Value:  params[twATKT],
				Path:   "/",
				Domain: "treatwell.fr",
			},
			{
				Name:   "ITKT",
				Value:  params[twITKT],
				Path:   "/",
				Domain: "treatwell.fr",
			},
			{
				Name:   "tw_user_id",
				Value:  params[twUserID],
				Path:   "/",
				Domain: "connect.treatwell.fr",
			},
		})

		client := &http.Client{
			Timeout: 1 * time.Minute,
			Jar:     cookieJar,
		}

		tw, err := treatwell.New(client, "428563")
		if err != nil {
			panic(err)
		}

		ga, err := google_calendar.New(
			context.Background(),
			"calendar-sync@helia-nails.iam.gserviceaccount.com",
			[]byte(params[googleKey]),
		)

		sync := core.New(tw, ga)

		from := utils.BoD(time.Now().Add(0 * time.Hour))
		to := utils.EoD(from.Add(0 * 24 * time.Hour))

		err = sync.TreatwellToGoogleCalendar(google_calendar.ClassPassCalendarID, from, to, models.SourceClassPass)
		if err != nil {
			return fmt.Errorf("failed to sync TW to Google: %w", err)
		}

		err = sync.GoogleCalendarToTreatwell(google_calendar.ClassPassCalendarID, from, to)
		if err != nil {
			return fmt.Errorf("failed to sync Google to TW: %w", err)
		}

		return nil
	},
}

func getParams(ctx context.Context, paramNames ...string) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-3"))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws context: %w", err)
	}

	svc := ssm.NewFromConfig(cfg)

	params := map[string]string{}

	for _, paramName := range paramNames {
		o, err := svc.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(paramName),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get param %s: %w", paramName, err)
		}
		params[paramName] = *o.Parameter.Value
	}

	return params, nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
