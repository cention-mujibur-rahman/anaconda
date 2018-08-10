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
