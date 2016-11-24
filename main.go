package main

import (
  "encoding/base64"
	"fmt"
  "crypto"
  "crypto/rsa"
  "crypto/sha256"
  "crypto/rand"
	//"golang.org/x/oauth2"
	//"golang.org/x/oauth2/google"
	//"google.golang.org/api/analyticsreporting/v4"
	"log"
	"io/ioutil"
	"encoding/json"
  "time"
)

// Create jwt: header, claim set signature:

// Config contains the site configuration.
type Config struct {
	ServiceAccountID string
	KeyID            string
  JWTClaimSet      JWTClaimSet
	//YoutubeApiKey string   `json:"YoutubeApiKey"`
	//GAToken   string   `json:"GAToken"`
	//Categories    []string `json:"Categories"`
}

type JWTClaimSet struct {
  Iss string `json:"iss"`
  Scope string `json:"scope"`
  Aud string `json:"aud"`
  Exp int64 `json:"exp"`
  Iat int64 `json:"iat"`
  Sub string `json:"sub"`
}

var config Config

// init read the configuration file and initialize github SHAs

// Initialize oauth and analytics client
// make query to analytics, get the top 10
// get reports from rwapi
// write reports to gist

// Initialize oauth and analytics client
func main() {
    fmt.Println("here")

	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Fatal("Cannot read configuration file.")
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Invalid configuration file.")
	}
  // header
  headerByte := []byte("{\"alg\":\"RS256\",\"typ\":\"JWT\"}")
  header := base64.StdEncoding.EncodeToString(headerByte)
  fmt.Println(header)
  jwtClaimSet := config.JWTClaimSet
  jwtClaimSet.Exp = time.Now().Unix()
  jwtClaimSet.Iat = time.Now().Add(time.Hour).Unix()
  fmt.Println(jwtClaimSet)
  // claim set
  claimSetJson, err := json.Marshal(jwtClaimSet)
  if err != nil {
		log.Fatal("Failed marshaling claimset.")
  }
  claimSet := base64.StdEncoding.EncodeToString(claimSetJson)
  fmt.Println(claimSet)
  // JWS
  rng := rand.Reader
  message := []byte(fmt.Sprint("{", header, "}.{", claimSet, "}"))
  hashed := sha256.Sum256(message)
  //rsa.PrivateKey
  signature, err := rsa.SignPKCS1v15(rng, config.KeyID, crypto.SHA256, hashed[:])
  if err != nil {
		log.Fatal("Failed signing.")
  }
  fmt.Println(signature)






	//_ = oauth2.NewClient(oauth2.NoContext, ts)
	//tc := oauth2.NewClient(oauth2.NoContext, ts)
	//client := github.NewClient(tc)
	//ref, _, err := client.Git.GetRef("rwapps", "video_backups", "heads/master")
	//if err != nil {
	//	log.Fatal("git getref error")
	//}
}


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
