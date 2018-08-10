package anaconda

type DirectMessage struct {
	CreatedAt           string   `json:"created_at"`
	Entities            Entities `json:"entities"`
	Id                  int64    `json:"id"`
	IdStr               string   `json:"id_str"`
	Recipient           User     `json:"recipient"`
	RecipientId         int64    `json:"recipient_id"`
	RecipientScreenName string   `json:"recipient_screen_name"`
	Sender              User     `json:"sender"`
	SenderId            int64    `json:"sender_id"`
	SenderScreenName    string   `json:"sender_screen_name"`
	Text                string   `json:"text"`
}

//DMEvent represents the payload of direct message single event
type DMEvent struct {
	EventType        string `json:"type"`
	ID               string `json:"id"`
	CreatedTimestamp string `json:"created_timestamp"`
	MessageCreate    struct {
		Target struct {
			RecipientID string `json:"recipient_id"`
		} `json:"target"`
		SenderID    string      `json:"sender_id"`
		MessageData MessageData `json:"message_data"`
	} `json:"message_create"`
}

//MessageData is the event message_data
type MessageData struct {
	Text     string `json:"text"`
	Entities Entities
}

//DMEventList ...
type DMEventList struct {
	NextCursor string    `json:"next_cursor"`
	Events     []DMEvent `json:"events"`
}
