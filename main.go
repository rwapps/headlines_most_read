package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/analyticsreporting/v4"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Config struct {
	GistToken string
	GistId    string
	JWT       struct {
		Client_email   string
		Private_key_id string
		Private_key    string
	}
}

type JWTClaimSet struct {
	Iss   string `json:"iss"`
	Scope string `json:"scope"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
	Sub   string `json:"sub"`
}

type Content struct {
	Data []struct {
		Id     string
		Fields struct {
			Title  string
			Source []struct {
				Shortname string
			}
			Date struct {
				Created string
			}
			Country []struct {
				Shortname string
			}
		}
	}
}

type Reports map[string]Report

// TODO: fix what we want here.
// Summary is for headlines only - do we want only most-read that are headlines?
type Report struct {
	URL       string `json:"url"`
	URL_alias string `json:"url_alias"`
	Id        string `json:"id"`
	Language  string `json:"language"`
	Title     string `json:"title"`
	//Summary string `json:"url"`
	//ImgSmall string `json:"url"`
	//Img string `json:"url"`
	Date          string         `json:"date"`
	Countries     []Country      `json:"country"`
	Themes        []Theme        `json:"theme"`
	BodyHtml      string         `json:"body-html"`
	Organizations []Organization `json:"source"`
	File          string         `json:"file"`
}

type Date struct {
	Created string `json:"created"`
}

type Organization struct {
	Name  string
	Image string
}

type Country struct {
	URL     string
	Id      int
	Name    string
	Iso3    string
	Primary bool
}

type Theme struct {
	Id   int
	Name string
}

type Filter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type Fields struct {
	Include []string `json:"include"`
}

type Params struct {
	Fields `json:"fields"`
	Filter `json:"filter"`
}

var config Config

func init() {
	data, err := ioutil.ReadFile("/go/src/github.com/rwapps/headlines_most_read/config/config.json")
	if err != nil {
		log.Fatal("Cannot read configuration file.")
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Invalid configuration file.")
	}
}

// make query to analytics, get the top 10
func getMostRead() *analyticsreporting.Report {
	var scopes []string
	scopes = append(scopes, "https://www.googleapis.com/auth/analytics.readonly")
	jwtConfig := jwt.Config{
		Email:        config.JWT.Client_email,
		PrivateKey:   []byte(config.JWT.Private_key),
		PrivateKeyID: config.JWT.Private_key_id,
		Scopes:       scopes,
		TokenURL:     "https://www.googleapis.com/oauth2/v4/token",
	}
	ctx := context.Background()
	client := jwtConfig.Client(ctx)
	analyticsreportingService, err := analyticsreporting.New(client)
	if err != nil {
		log.Fatal("no new service.")
	}
	var metrics []*analyticsreporting.Metric
	metrics = append(metrics, &analyticsreporting.Metric{Expression: "ga:pageviews"})
	var dimensions []*analyticsreporting.Dimension
	dimensions = append(dimensions, &analyticsreporting.Dimension{Name: "ga:pagePath"})
	dimensions = append(dimensions, &analyticsreporting.Dimension{Name: "ga:pageTitle"})
	var orderBys []*analyticsreporting.OrderBy
	orderBys = append(orderBys, &analyticsreporting.OrderBy{
		FieldName: "ga:pageviews",
		OrderType: "VALUE",
		SortOrder: "DESCENDING",
	})
	var reportRequests []*analyticsreporting.ReportRequest
	request := &analyticsreporting.ReportRequest{
		Dimensions:        dimensions,
		FiltersExpression: "ga:dimension1==Report",
		Metrics:           metrics,
		OrderBys:          orderBys,
		PageSize:          20,
		SamplingLevel:     "LARGE",
		ViewId:            "75062",
	}
	reportRequests = append(reportRequests, request)
	getReportsRequest := &analyticsreporting.GetReportsRequest{ReportRequests: reportRequests}
	reportsService := analyticsreporting.NewReportsService(analyticsreportingService)
	reportsBatchGetCall := reportsService.BatchGet(getReportsRequest)
	response, err := reportsBatchGetCall.Do()
	if err != nil {
		log.Fatal(err)
	}
	return response.Reports[0]
}

func getReports() map[string]Report {
	analyticsReport := getMostRead()
	reports := make(map[string]Report)
	var params Params
	params.Fields.Include = []string{"title", "source.shortname", "country.shortname", "date.created"}
	params.Filter.Field = "url_alias"
	i := 1
	for _, row := range analyticsReport.Data.Rows {
		url := fmt.Sprint("http://reliefweb.int", row.Dimensions[0])
		report := Report{URL: url}
		params.Filter.Value = url
		report = addRWData(report, params)
		reports[strconv.Itoa(i)] = report
		i++
		if i > 10 {
			return reports
		}
	}
	return reports
}

func addRWData(report Report, params Params) Report {
	paramsJson, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post("http://api.reliefweb.int/v1/reports", "application/json", bytes.NewBuffer(paramsJson))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	content := Content{}
	err = json.Unmarshal(body, &content)
	if err != nil {
		log.Fatal(err)
	}
	for _, data := range content.Data {
		report.Id = data.Id
		report.Date = data.Fields.Date.Created
		report.Title = data.Fields.Title
		for _, source := range data.Fields.Source {
			report.Organizations = append(report.Organizations, Organization{Name: source.Shortname})
		}
		for _, country := range data.Fields.Country {
			report.Countries = append(report.Countries, Country{Name: country.Shortname})
		}
	}
	return report
}

func updateGist(name string, gistId string, content []byte) {
	token := config.GistToken
	url := "https://api.github.com/gists/" + gistId
	payload := fmt.Sprintf("{ \"files\": { \"%s.json\": { \"content\": %q } } }", name, content)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "rwapps")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to readall body")
	}
	// TODO: Add error feedback
	if resp.Status == "200 OK" {
		fmt.Println("Success")
	} else {
		fmt.Printf("Failed updating gist, error body\n %s\n", body)
	}
}

func main() {
	reports := getReports()
	reportsJson, err := json.Marshal(reports)
	if err != nil {
		panic(err)
	}
	updateGist("most_read", config.GistId, reportsJson)
}
