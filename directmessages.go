package anaconda

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

//GetDirectMessagesList Returns all Direct Message events (both sent and received) within the last 30 days.
//Sorted in reverse-chronological order.
//https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/list-events
func (a TwitterApi) GetDirectMessagesList(v url.Values) (messages DMEventList, err error) {
	responseCh := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/events/list.json", v, &messages, _GET, responseCh}
	return messages, (<-responseCh).err
}

//GetDirectMessagesSent deprecated
func (a TwitterApi) GetDirectMessagesSent(v url.Values) (messages []DirectMessage, err error) {
	response_ch := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/sent.json", v, &messages, _GET, response_ch}
	return messages, (<-response_ch).err
}

//GetDirectMessagesShow Returns a single Direct Message event by the given id.
//https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/get-event
func (a TwitterApi) GetDirectMessagesShow(v url.Values) (message DirectMessage, err error) {
	response_ch := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/events/show.json", v, &message, _GET, response_ch}
	return message, (<-response_ch).err
}

//PostDMToScreenName deprecated
// https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/new-message
func (a TwitterApi) PostDMToScreenName(text, screenName string) (message DirectMessage, err error) {
	v := url.Values{}
	v.Set("screen_name", screenName)
	v.Set("text", text)
	return a.postDirectMessagesImpl(v)
}

//PostDMToUserId deprecated
// https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/new-message
func (a TwitterApi) PostDMToUserId(text string, userId int64) (message DirectMessage, err error) {
	v := url.Values{}
	v.Set("user_id", strconv.FormatInt(userId, 10))
	v.Set("text", text)
	return a.postDirectMessagesImpl(v)
}

// DeleteDirectMessage will destroy (delete) the direct message with the specified ID.
// https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/delete-message-event
func (a TwitterApi) DeleteDirectMessage(id int64, includeEntities bool) (message DirectMessage, err error) {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	response_ch := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/events/destroy.json", v, &message, _POST, response_ch}
	return message, (<-response_ch).err
}

//postDirectMessagesImpl un-used
func (a TwitterApi) postDirectMessagesImpl(v url.Values) (message DirectMessage, err error) {
	response_ch := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/new.json", v, &message, _POST, response_ch}
	return message, (<-response_ch).err
}

// IndicateTyping will create a typing indicator
// https://developer.twitter.com/en/docs/direct-messages/typing-indicator-and-read-receipts/api-reference/new-typing-indicator
func (a TwitterApi) IndicateTyping(id int64) (err error) {
	v := url.Values{}
	v.Set("recipient_id", strconv.FormatInt(id, 10))
	response_ch := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/direct_messages/indicate_typing.json", v, nil, _POST, response_ch}
	return (<-response_ch).err
}

//NewDirectMessage Publishes a new message_create event resulting in a Direct Message sent to a specified user from the authenticating user.
//Returns an event if successful. Supports publishing Direct Messages with optional Quick Reply and media attachment.
//Replaces behavior currently provided by POST direct_messages/new.
//https://developer.twitter.com/en/docs/direct-messages/sending-and-receiving/api-reference/new-event
func (a TwitterApi) NewDirectMessage(jd []byte) (rsBody []byte, err error) {
	result := make(map[string]interface{})
	res, err := a.Do(a.HttpClient, a.baseUrl+"/direct_messages/events/new.json", jd)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	result = decodeRespBody(res.Body)
	rsBody, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return rsBody, nil
}

//Do execute the send query by http client
// It will return the result as *http.Response
func (a TwitterApi) Do(client *http.Client, urlStr string, jd []byte) (*http.Response, error) {
	req, err := a.doHttpReq(client, urlStr, "POST", jd)
	if err != nil{
		return nil, err
	}
	return client.Do(req)
}

func decodeRespBody(wr io.Reader) (body map[string]interface{}) {
	body = make(map[string]interface{})
	json.NewDecoder(wr).Decode(&body)
	return body
}

func (a TwitterApi) doHttpReq(client *http.Client, URL, method string, reader []byte)(*http.Request, error){
	rb :=  bytes.NewReader(reader)
	req, err := http.NewRequest(method, URL, rb)
	if err != nil {
		return nil, err
	}
	if req.URL.RawQuery != "" {
		return nil, errors.New("oauth: url must not contain a query string")
	}
	c := a.oauthClient
	for k, v := range c.Header {
		req.Header[k] = v
	}
	auth := c.AuthorizationHeader(a.Credentials, method, req.URL, nil)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	if client == nil {
		client = http.DefaultClient
	}
	return req, nil
}

//GetDirectMessagesMedia fetch direct messages media
func (a TwitterApi) GetDirectMessagesMedia(mediaURL string, v url.Values) (*http.Response, error) {
	client := a.HttpClient
	req, err := a.doHttpReq(client, mediaURL, "GET", nil)
	if err != nil{
		return nil, err
	}
	return client.Do(req)
}
