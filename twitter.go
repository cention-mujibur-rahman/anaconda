//Package anaconda provides structs and functions for accessing version 1.1
//of the Twitter API.
//
//Successful API queries return native Go structs that can be used immediately,
//with no need for type assertions.
//
//Authentication
//
//If you already have the access token (and secret) for your user (Twitter provides this for your own account on the developer portal), creating the client is simple:
//
//  anaconda.SetConsumerKey("your-consumer-key")
//  anaconda.SetConsumerSecret("your-consumer-secret")
//  api := anaconda.NewTwitterApi("your-access-token", "your-access-token-secret")
//
//
//Queries
//
//Executing queries on an authenticated TwitterApi struct is simple.
//
//  searchResult, _ := api.GetSearch("golang", nil)
//  for _ , tweet := range searchResult {
//      fmt.Print(tweet.Text)
//  }
//
//Certain endpoints allow separate optional parameter; if desired, these can be passed as the final parameter.
//
//  v := url.Values{}
//  v.Set("count", "30")
//  result, err := api.GetSearch("golang", v)
//
//
//Endpoints
//
//Anaconda implements most of the endpoints defined in the Twitter API documentation: https://dev.twitter.com/docs/api/1.1.
//For clarity, in most cases, the function name is simply the name of the HTTP method and the endpoint (e.g., the endpoint `GET /friendships/incoming` is provided by the function `GetFriendshipsIncoming`).
//
//In a few cases, a shortened form has been chosen to make life easier (for example, retweeting is simply the function `Retweet`)
//
//More detailed information about the behavior of each particular endpoint can be found at the official Twitter API documentation.
package anaconda

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	_GET  = iota
	_POST = iota
)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "http://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "http://api.twitter.com/oauth/authenticate",
	TokenRequestURI:               "http://api.twitter.com/oauth/access_token",
}

type TwitterApi struct {
	Credentials *oauth.Credentials
	queryQueue  chan query
	delay       time.Duration
	delay_mutex sync.Mutex
}

type query struct {
	url         string
	form        url.Values
	data        interface{}
	method      int
	response_ch chan response
}

type response struct {
	data interface{}
	err  error
}

const DEFAULT_DELAY = 10 * time.Second

//NewTwitterApi takes an user-specific access token and secret and returns a TwitterApi struct for that user.
//The TwitterApi struct can be used for accessing any of the endpoints available.
func NewTwitterApi(access_token string, access_token_secret string) *TwitterApi {
	//TODO figure out how much to buffer this channel
	//A non-buffered channel will cause blocking when multiple queries are made at the same time
	queue := make(chan query)
	c := &TwitterApi{&oauth.Credentials{Token: access_token, Secret: access_token_secret}, queue, DEFAULT_DELAY, sync.Mutex{}}
	go c.throttledQuery()
	return c
}

//SetConsumerKey will set the application-specific consumer_key used in the initial OAuth process
//This key is listed on https://dev.twitter.com/apps/YOUR_APP_ID/show
func SetConsumerKey(consumer_key string) {
	oauthClient.Credentials.Token = consumer_key
}

//SetConsumerSecret will set the application-specific secret used in the initial OAuth process
//This secret is listed on https://dev.twitter.com/apps/YOUR_APP_ID/show
func SetConsumerSecret(consumer_secret string) {
	oauthClient.Credentials.Secret = consumer_secret
}

// SetDelay will set the delay between throttled queries
// To turn of throttling, set it to 0 seconds
func (c *TwitterApi) SetDelay(t time.Duration) {
	c.delay_mutex.Lock()
	c.delay = t
	c.delay_mutex.Unlock()
}

func (c *TwitterApi) GetDelay() time.Duration {
	c.delay_mutex.Lock()
	t := c.delay
	c.delay_mutex.Unlock()
	return t
}

//AuthorizationURL generates the authorization URL for the first part of the OAuth handshake.
//Redirect the user to this URL.
//This assumes that the consumer key has already been set (using SetConsumerKey).
func AuthorizationURL(callback string) (string, error) {
	tempCred, err := oauthClient.RequestTemporaryCredentials(http.DefaultClient, callback, nil)
	if err != nil {
		return "", err
	}
	return oauthClient.AuthorizationURL(tempCred, nil), nil
}

func cleanValues(v url.Values) url.Values {
	if v == nil {
		return url.Values{}
	}
	return v
}

// apiGet issues a GET request to the Twitter API and decodes the response JSON to data.
func (c TwitterApi) apiGet(urlStr string, form url.Values, data interface{}) error {
	resp, err := oauthClient.Get(http.DefaultClient, c.Credentials, urlStr, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, data)
}

// apiPost issues a POST request to the Twitter API and decodes the response JSON to data.
func (c TwitterApi) apiPost(urlStr string, form url.Values, data interface{}) error {
	resp, err := oauthClient.Post(http.DefaultClient, c.Credentials, urlStr, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, data)
}

// decodeResponse decodes the JSON response from the Twitter API.
func decodeResponse(resp *http.Response, data interface{}) error {
	if resp.StatusCode != 200 {
		// TODO don't ignore this error
		// TODO don't use ReadAll
		p, _ := ioutil.ReadAll(resp.Body)

		var twitterError TwitterErrorResponse
		err := json.Unmarshal(p, &twitterError)
		if err != nil {
			return fmt.Errorf("Could not even parse error. Get %s returned status %d, %s", resp.Request.URL, resp.StatusCode, p)
		}
		return twitterError
		//return fmt.Errorf("Get %s returned status %d, %s", resp.Request.URL, resp.StatusCode, p)
	}
	return json.NewDecoder(resp.Body).Decode(data)
}

//query executes a query to the specified url, sending the values specified by form, and decodes the response JSON to data
//method can be either _GET or _POST
func (c TwitterApi) execQuery(urlStr string, form url.Values, data interface{}, method int) error {
	switch method {
	case _GET:
		return c.apiGet(urlStr, form, data)
	case _POST:
		return c.apiPost(urlStr, form, data)
	default:
		return fmt.Errorf("HTTP method not yet supported")
	}
}

//throttledQuery executes queries and automatically throttles them according to SECONDS_PER_QUERY
func (c *TwitterApi) throttledQuery() {
	for q := range c.queryQueue {
		now := time.Now()
		url := q.url
		form := q.form
		data := q.data //This is where the actual response will be written
		method := q.method

		response_ch := q.response_ch

		err := c.execQuery(url, form, data, method)
		response_ch <- struct {
			data interface{}
			err  error
		}{data, err}

		//We have to hold a mutex because the delay time may have changed
		c.delay_mutex.Lock()
		delay := c.delay
		c.delay_mutex.Unlock()
		time.Sleep(delay - time.Since(now))
	}
}
