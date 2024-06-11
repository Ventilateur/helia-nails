package treatwell

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/core/models"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

const (
	baseURL = "https://connect.treatwell.fr"
	apiURL  = baseURL + "/api"
)

type Treatwell struct {
	httpClient *http.Client
	config     *config.Config

	venueID              string
	employees            *twmodels.Employees
	employeeWorkingHours *twmodels.EmployeesWorkingHours

	employeeInfo map[int]twmodels.EmployeeWrapper
}

func New(httpClient *http.Client, config *config.Config) (*Treatwell, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	u, err := url.Parse("https://treatwell.fr")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	cookieJar.SetCookies(u, []*http.Cookie{
		{
			Name:   "ATKT",
			Value:  config.Treatwell.ATKT,
			Path:   "/",
			Domain: "treatwell.fr",
		},
		{
			Name:   "ITKT",
			Value:  config.Treatwell.ITKT,
			Path:   "/",
			Domain: "treatwell.fr",
		},
		{
			Name:   "tw_user_id",
			Value:  config.Treatwell.UserId,
			Path:   "/",
			Domain: "connect.treatwell.fr",
		},
	})
	httpClient.Jar = cookieJar

	return &Treatwell{
		httpClient: httpClient,
		config:     config,
		venueID:    config.Treatwell.VenueId,
	}, nil
}

func (tw *Treatwell) Name() models.Source {
	return models.SourceTreatwell
}

func (tw *Treatwell) bootstrapCookie() error {
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return utils.RequestCreationErr(baseURL, err)
	}

	req.Header.Add("authority", "connect.treatwell.fr")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")

	resp, err := tw.httpClient.Do(req)
	if err != nil {
		return utils.DoRequestErr(baseURL, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

// Login hasn't been used so far because Treatwell is dumb, login once and use the tokens forever.
func (tw *Treatwell) Login(user, password string) error {
	err := tw.bootstrapCookie()
	if err != nil {
		return err
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{"user":"%s","password":"%s","persistentLogin":false}`,
		user, password,
	))

	req, err := http.NewRequest(http.MethodPost, apiURL+"/authentication.json", payload)
	if err != nil {
		return utils.RequestCreationErr(apiURL+"/authentication.json", err)
	}

	writeCommonHeaders(req, map[string]string{
		"referer": "https://connect.treatwell.fr/login",
	})

	resp, err := tw.httpClient.Do(req)
	if err != nil {
		return utils.DoRequestErr(apiURL+"/authentication.json", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

func (tw *Treatwell) Preload(from, to time.Time) error {
	var err error
	tw.employeeInfo = map[int]twmodels.EmployeeWrapper{}

	tw.employees, err = tw.getEmployeesInfo()
	if err != nil {
		return fmt.Errorf("failed to get employees info: %w", err)
	}

	for _, employee := range tw.employees.Employees {
		tw.employeeInfo[employee.Id] = twmodels.EmployeeWrapper{
			Info: employee,
		}
	}

	tw.employeeWorkingHours, err = tw.getAllEmployeesWorkingHours(from, to)
	if err != nil {
		return fmt.Errorf("failed to get employees working hours: %w", err)
	}

	for _, employeeWorkingHours := range tw.employeeWorkingHours.EmployeesWorkingHours {
		if info, ok := tw.employeeInfo[employeeWorkingHours.EmployeeID]; ok {
			info.WorkingHours = employeeWorkingHours.WorkingHours
			tw.employeeInfo[employeeWorkingHours.EmployeeID] = info
		} else {
			return fmt.Errorf("unknown employee %s", employeeWorkingHours.EmployeeName)
		}
	}

	return nil
}

func (tw *Treatwell) getAllEmployeesWorkingHours(from, to time.Time) (*twmodels.EmployeesWorkingHours, error) {
	var employeeIDs []string
	for _, employee := range tw.employees.Employees {
		employeeIDs = append(employeeIDs, strconv.Itoa(employee.Id))
	}

	return doRequestWithResponse[twmodels.EmployeesWorkingHours](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/employees/"+strings.Join(employeeIDs, ",")+"/working-hours.json",
		nil,
		map[string]string{
			"from": from.Format(time.DateOnly),
			"to":   to.Format(time.DateOnly),
		},
	)
}

func (tw *Treatwell) getEmployeesInfo() (*twmodels.Employees, error) {
	return doRequestWithResponse[twmodels.Employees](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/employees.json",
		nil,
		map[string]string{
			"include": "employee-offers",
		},
	)
}

func writeCommonHeaders(req *http.Request, extra map[string]string) {
	req.Header.Add("authority", "connect.treatwell.fr")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://connect.treatwell.fr")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

	for k, v := range extra {
		req.Header.Add(k, v)
	}
}

func doRequestWithResponse[T any](tw *Treatwell, method, u string, body io.Reader, params map[string]string) (*T, error) {
	resp, err := tw.doRequest(method, u, body, params)
	if err != nil {
		return nil, utils.DoRequestErr(u, err)
	}
	defer resp.Body.Close()

	ok := func() bool {
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated, http.StatusNoContent:
			return true
		default:
			return false
		}
	}()

	if !ok {
		return nil, utils.UnexpectedErrorCode(resp.StatusCode)
	}

	var ret T
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
	}

	return &ret, nil
}

func doRequestWithoutResponse(tw *Treatwell, method, u string, body io.Reader, params map[string]string) error {
	resp, err := tw.doRequest(method, u, body, params)
	if err != nil {
		return utils.DoRequestErr(u, err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	defer resp.Body.Close()

	ok := func() bool {
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated, http.StatusNoContent:
			return true
		default:
			return false
		}
	}()

	if !ok {
		slog.Error(string(respBody))
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

func (tw *Treatwell) doRequest(method, u string, body io.Reader, params map[string]string) (*http.Response, error) {
	urlObj, err := url.Parse(u)
	if err != nil {
		return nil, utils.URLParseErr(u, err)
	}

	q := urlObj.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	urlObj.RawQuery = q.Encode()

	req, err := http.NewRequest(method, urlObj.String(), body)
	if err != nil {
		return nil, utils.RequestCreationErr(urlObj.String(), err)
	}

	writeCommonHeaders(req, map[string]string{})

	return tw.httpClient.Do(req)
}
