package main

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analyticsreporting/v4"
	"log"
	//"io/ioutil"
	//"encoding/json"
)

// Create jwt: header, claim set signature:

// Config contains the site configuration.
type Config struct {
	ServiceAccountID string
	KeyID            string
	//YoutubeApiKey string   `json:"YoutubeApiKey"`
	//GAToken   string   `json:"GAToken"`
	//Categories    []string `json:"Categories"`
}

var config Config

// init read the configuration file and initialize github SHAs

func main() {
    fmt.Println("here")
	ctx := oauth2.NoContext
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/devstorage.readonly")
	if err != nil {
    fmt.Printf("the error is %s", err)
	}
	httpClient := oauth2.NewClient(ctx, ts)
	service, err := analyticsreporting.New(httpClient)
	reportRequest := analyticsreporting.ReportRequest{ViewId: "75062", PageSize: 10}
	getReportsRequest := analyticsreporting.GetReportsRequest{}
	getReportsRequest.ReportRequests = append(getReportsRequest.ReportRequests, &reportRequest)
	reportsService := analyticsreporting.NewReportsService(service)
	reportsBatchGetCall := reportsService.BatchGet(&getReportsRequest)
	response, err := reportsBatchGetCall.Do()
	if err != nil {
		log.Fatal("Cannot create defaultTokenSource.")
	}
	fmt.Println(response)

	//data, err := ioutil.ReadFile("config/config.json")
	//if err != nil {
	//	log.Fatal("Cannot read configuration file.")
	//}
	//err = json.Unmarshal(data, &config)
	//if err != nil {
	//	log.Fatal("Invalid configuration file.")
	//}
	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: config.GAToken},
	//)

	//  {
	//    "iss":"761326798069-r5mljlln1rd4lrbhg75efgigp36m78j5@developer.gserviceaccount.com",
	//    "scope":"https://www.googleapis.com/auth/devstorage.readonly",
	//    "aud":"https://www.googleapis.com/oauth2/v4/token",
	//    "exp":1328554385,
	//    "iat":1328550785
	//  }
	//_ = oauth2.NewClient(oauth2.NoContext, ts)
	//tc := oauth2.NewClient(oauth2.NoContext, ts)
	//client := github.NewClient(tc)
	//ref, _, err := client.Git.GetRef("rwapps", "video_backups", "heads/master")
	//if err != nil {
	//	log.Fatal("git getref error")
	//}
}

// Initialize oauth and analytics client
// make query to analytics, get the top 10
// get reports from rwapi
// write reports to gist

//  /*
//   * Initialize the connection to Google Analtycs account
//   */
//  function initializeAnalytics()
//  {
//      $KEY_FILE_LOCATION = __DIR__ . '/credentials.json';
//      $client = new Google_Client();
//      $client->setAuthConfig($KEY_FILE_LOCATION);
//      $client->setScopes(['https://www.googleapis.com/auth/analytics.readonly']);
//      $analytics = new Google_Service_Analytics($client);
//      return $analytics;
//  }
//  /*
//   * Return the most read reports in the Google Analytics format
//   */
//  function getResults($analytics, $profileId, $startDate, $endDate, $numberOfResults)
//  {
//      return $analytics->data_ga->get(
//          'ga:' . $profileId,
//  	$startDate,
//          $endDate,
//          'ga:sessions, ga:pageviews, ga:socialInteractions',
//           array('dimensions'    => 'ga:pagePath, ga:pageTitle',
//  	       'sort'          => '-ga:pageviews',
//  	       'filters'       => 'ga:dimension1==Report',
//  	       'samplingLevel' => 'HIGHER_PRECISION',
//  	       'max-results'   => $numberOfResults
//  		)
//  	);
//  }
//  /*
//   * Use the RW API to get the reports with the url_alias provided by Google Analytics
//   */
//  function getReports($results, $webhost, $apiEndPoint)
//  {
//      $jsons = array();
//      foreach($results["rows"] as $result)
//      {
//          $url = urlencode($webhost . $result[0]);
//          $content = file_get_contents($apiEndPoint . "?filter[field]=url_alias&filter[value]=" . $url);
//  	$content = json_decode($content, true);
//  	$href = $content["data"][0]["href"];
//  	$jsons[] = file_get_contents($href);
//      }
//      return $jsons;
//  }
//  $analytics = initializeAnalytics();
//  $results = getResults($analytics, $profileId, $startDate, $endDate, $numberOfResults);
//  $reports = getReports($results, $webhost, $apiEndPoint);
//  /*
//   * Write the reports in the gist
//   */
//  $data   = array("files" => array($gistFilename => array( "content" => json_encode($reports))));
//  $header = array("Authorization: token " . $gistToken, "User-Agent: ReliefWeb API");
//  $curl = curl_init();
//  curl_setopt($curl, CURLOPT_URL, $gist);
//  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
//  curl_setopt($curl, CURLOPT_CUSTOMREQUEST, 'PATCH');
//  curl_setopt($curl, CURLOPT_POSTFIELDS, json_encode($data));
//  curl_setopt($curl, CURLOPT_HTTPHEADER, $header);
//  curl_exec($curl);
//  curl_close($curl);
