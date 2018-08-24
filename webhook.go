package anaconda

import (
	"net/url"
)

var (
	enterpriseAPITier = "enterprise"
	premiumAPITier    = "premium"
)

//GetAppActivityWebhooks represents the twitter account_activity webhook
//Returns all URLs and their statuses for the given app. Currently,
//only one webhook URL can be registered to an application.
//https://dev.twitter.com/webhooks/reference/get/account_activity/webhooks
func (a TwitterApi) GetAppActivityWebhooks(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/account_activity/all/webhooks.json", v, &u, _GET, responseCh}
	return u, (<-responseCh).err
}

func getWebhookURL(baseURL, apiTier, envName, webhookID string) string {
	if apiTier == premiumAPITier {
		//https://api.twitter.com/1.1/account_activity/all/webhooks.json
		return baseURL + "/account_activity/all/" + envName + "/subscriptions.json"
	}
	return baseURL + "/account_activity/webhooks/" + webhookID + "/subscriptions/all.json"
}

//CountAppActivityWebhooks Returns the count of subscriptions that are currently active on your account for all activities.
//Note that the /count endpoint requires application-only OAuth, so that you should make requests using a bearer token
//instead of user context.
func (a TwitterApi) CountAppActivityWebhooks(v url.Values) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/account_activity/subscriptions/count.json", v, &u, _GET, responseCh}
	return u, (<-responseCh).err
}

//WebHookCount represents the Get webhook responses
type WebHookCount struct {
	AccountName string `json:"account_name"`
	SubCountAll string `json:"subscriptions_count_all"`
	SubsCountDM bool   `json:"subscriptions_count_direct_messages"`
}

//WebHookResp represents the Get webhook responses
type WebHookResp struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Valid     bool   `json:"valid"`
	CreatedAt string `json:"created_at"`
}

//SetAppActivityWebhooks represents to set twitter account_activity webhook
//Registers a new webhook URL for the given application context.
//The URL will be validated via CRC request before saving. In case the validation fails,
//a comprehensive error is returned. message to the requester.
//Only one webhook URL can be registered to an application.
//https://api.twitter.com/1.1/account_activity/webhooks.json
func (a TwitterApi) SetAppActivityWebhooks(v url.Values, envName, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	a.queryQueue <- query{a.baseUrl + "/account_activity/all/" + envName + "/webhooks.json", v, &u, _POST, responseCh}
	return u, (<-responseCh).err
}

//DeleteAppActivityWebhooks Removes the webhook from the provided application’s configuration.
//https://dev.twitter.com/webhooks/reference/del/account_activity/webhooks
func (a TwitterApi) DeleteAppActivityWebhooks(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	URL := a.baseUrl + "/account_activity/all/" + envName + "/webhooks/" + webhookID + ".json"
	if apiTier == enterpriseAPITier {
		URL = a.baseUrl + "/account_activity/webhooks/" + webhookID + ".json"
	}
	a.queryQueue <- query{URL, v, &u, _DELETE, responseCh}
	return u, (<-responseCh).err
}

//PutAppActivityWebhooks update webhook which reenables the webhook by setting its status to valid.
//https://dev.twitter.com/webhooks/reference/put/account_activity/webhooks
func (a TwitterApi) PutAppActivityWebhooks(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	URL := a.baseUrl + "/account_activity/all/" + envName + "/webhooks/" + webhookID + ".json"
	if apiTier == enterpriseAPITier {
		URL = a.baseUrl + "/account_activity/webhooks/" + webhookID + ".json"
	}
	a.queryQueue <- query{URL, v, &u, _PUT, responseCh}
	return u, (<-responseCh).err
}

//SetWHSubscription Subscribes the provided app to events for the provided user context.
//When subscribed, all DM events for the provided user will be sent to the app’s webhook via POST request.
//https://developer.twitter.com/en/docs/accounts-and-users/subscribe-account-activity/api-reference
func (a TwitterApi) SetWHSubscription(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	whURL := getWebhookURL(a.baseUrl, apiTier, envName, webhookID)
	a.queryQueue <- query{whURL, v, &u, _POST, responseCh}
	return u, (<-responseCh).err
}

//GetWHSubscription Provides a way to determine if a webhook configuration is
//subscribed to the provided user’s Direct Messages.
//https://dev.twitter.com/webhooks/reference/get/account_activity/webhooks/subscriptions
func (a TwitterApi) GetWHSubscription(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	//EnterPrise not impelmented
	a.queryQueue <- query{a.baseUrl + "/account_activity/all/" + envName + "/subscriptions.json", v, &u, _GET, responseCh}
	return u, (<-responseCh).err
}

//GetWHSubscriptionList Provides a way to determine if a webhook configuration is
//subscribed to the provided user’s Direct Messages.
//https://dev.twitter.com/webhooks/reference/get/account_activity/webhooks/subscriptions
func (a TwitterApi) GetWHSubscriptionList(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	//EnterPrise not impelmented
	a.queryQueue <- query{a.baseUrl + "account_activity/all/" + envName + "/subscriptions/list.json", v, &u, _GET, responseCh}
	return u, (<-responseCh).err
}

//DeleteWHSubscription Deactivates subscription for the provided user context and app. After deactivation,
//all DM events for the requesting user will no longer be sent to the webhook URL..
//https://developer.twitter.com/en/docs/accounts-and-users/subscribe-account-activity/api-reference
func (a TwitterApi) DeleteWHSubscription(v url.Values, envName, webhookID, apiTier string) (u interface{}, err error) {
	v = cleanValues(v)
	responseCh := make(chan response)
	if apiTier == premiumAPITier {
		a.queryQueue <- query{a.baseUrl + "/account_activity/all/" + envName + "/subscriptions.json", v, &u, _DELETE, responseCh}
	} else {
		a.queryQueue <- query{a.baseUrl + "/account_activity/webhooks/" + webhookID + "/subscriptions/all.json", v, &u, _DELETE, responseCh}
	}
	return u, (<-responseCh).err
}
