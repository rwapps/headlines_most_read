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
			URL_alias string
			Body_html string `json:"body-html"`
			Image     struct {
				URL_small string `json:"url-small"`
			}
			Title string
			File  []struct {
				URL string
			}
			Source []struct {
				Id        int
				Shortname string
			}
			Date struct {
				Created string
			}
		}
	}
}

type Source struct {
	Data []struct {
		Fields struct {
			Logo struct {
				URL string
			}
		}
	}
}

type Reports map[string]Report

type Report struct {
	URL           string         `json:"url"`
	URL_alias     string         `json:"url_alias"`
	Id            string         `json:"id"`
	Title         string         `json:"title"`
	Image         string         `json:"image"`
	Type          string         `json:"type"`
	Date          string         `json:"date"`
	BodyHtml      string         `json:"body_html"`
	Organizations []Organization `json:"source"`
	Files         []string       `json:"files"`
}

type Date struct {
	Created string `json:"created"`
}

type Organization struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Filter struct {
	Field string   `json:"field"`
	Value []string `json:"value"`
}

type Fields struct {
	Include []string `json:"include"`
}

type Params struct {
	Fields `json:"fields"`
	Filter `json:"filter"`
}

var config Config

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
		PageSize:          10,
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
	params.Fields.Include = []string{"title", "body-html", "url_alias", "source.id", "file.url", "image.url-small", "source.shortname", "image.url-small", "country.shortname", "date.created"}
	params.Filter.Field = "url_alias"

	i := 1
	for _, row := range analyticsReport.Data.Rows {
		url_alias := fmt.Sprint("http://reliefweb.int", row.Dimensions[0])
		report := Report{URL_alias: url_alias}
		params.Filter.Value = append(params.Filter.Value, url_alias)
		// TODO: add err if we have incomplete information. If so, continue.
		reports[strconv.Itoa(i)] = report
		i++
		// 8 is enough.
		if i > 8 {
			addRWDataMultiple(reports, params)
			return reports
		}
	}
	return reports
}

func queryRWApi(contentType string, params Params) []byte {
	paramsJson, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post("http://api.reliefweb.int/v1/"+contentType, "application/json", bytes.NewBuffer(paramsJson))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func addRWDataMultiple(reports map[string]Report, params Params) map[string]Report {
	contentType := "reports"
	body := queryRWApi(contentType, params)
	content := Content{}
	err := json.Unmarshal(body, &content)
	if err != nil {
		log.Fatal(err)
	}
	sourceIds := []string{}
	for _, data := range content.Data {
		for k, _ := range reports {
			if reports[k].URL_alias == data.Fields.URL_alias {
				report := reports[k]
				report.Id = data.Id
				report.URL = fmt.Sprint("http://reliefweb.int/node/", data.Id)
				report.Date = data.Fields.Date.Created
				report.Title = data.Fields.Title
				report.Type = "report"
				report.BodyHtml = data.Fields.Body_html
				report.Image = data.Fields.Image.URL_small
				for _, source := range data.Fields.Source {
					sourceId := strconv.Itoa(source.Id)
					sourceIds = append(sourceIds, sourceId)
					report.Organizations = append(report.Organizations, Organization{Name: source.Shortname, Id: sourceId})
				}
				for _, file := range data.Fields.File {
					report.Files = append(report.Files, file.URL)
				}
				reports[k] = report
				break
			}
		}
	}
	reports = addSourceImages(sourceIds, reports)
	return reports
}

func addSourceImages(sourceIds []string, reports map[string]Report) map[string]Report {
	sourceImages := getSourceImages(sourceIds)
	for k, _ := range reports {
		report := reports[k]
		for index, _ := range report.Organizations {
			organization := report.Organizations[index]
			organization.Image = sourceImages[organization.Id]
			report.Organizations[index] = organization
		}
		reports[k] = report
	}
	return reports
}

func getSourceImages(sourceIds []string) map[string]string {
	sourceImages := make(map[string]string)
	var params Params
	params.Fields.Include = []string{"logo.url"}
	params.Filter.Field = "id"
	params.Filter.Value = sourceIds
	body := queryRWApi("sources", params)
	source := Source{}
	err := json.Unmarshal(body, &source)
	if err != nil {
		log.Fatal(err)
	}
	for _, data := range source.Data {
		for _, v := range sourceIds {
			sourceImages[v] = data.Fields.Logo.URL
		}
	}
	return sourceImages
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

func main() {
	reports := getReports()
	reportsJson, err := json.Marshal(reports)
	if err != nil {
		panic(err)
	}
	updateGist("most_read", config.GistId, reportsJson)
}
