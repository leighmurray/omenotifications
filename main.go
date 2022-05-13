package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/leighmurray/omedb"
)

const (
	vapidPublicKey  = "BLopEDYPbeESlGuxRdCsXZxUFLTP1nY-bJsS8eu6TTkkYeJP7tQ9ZZijXOmFwSUcyE3vrUSu95tWHxZeWOyd8X4"
	vapidPrivateKey = "xfhegdXnCuSKa-66IJFERIpv5ylQzd4bjbVcUlowThg"
)

type StreamListResponse struct {
	Message		string `json:"message"`
	Response	[]string `json:"response"`
	StatusCode	int `json:"statusCode"`
}

type StreamInfoResponse struct {
	Message		string	`json:"message"`
	Response	struct {
		Connections		map[string]int	`json:"connections"`
		CreatedTime		time.Time	`json:"createdTime"`
		LastRecvTime		time.Time	`json:"lastRecvTime"`
		LastSentTime		time.Time	`json:"lastSentTime"`
		LastUpdatedTime		time.Time	`json:"lastUpdatedTime"`
		MaxTotalConnectionTime	time.Time	`json:"maxTotalConnectionTime"`
		MaxTotalConnections	int		`json:"maxTotalConnections"`
		TotalBytesIn		int		`json:"totalBytesIn"`
		TotalBytesOut		int		`json:"totalBytesOut"`
		TotalConnections	int		`json:"totalConnections"`
	} `json:"response"`
	StatusCode	int	`json:"statusCode"`
}

func sendNotifications(streamer string) {
	fmt.Println("We're gonna notify users about a new stream for: " + streamer)

	subscriptions := omedb.GetSubscriptions()

	for _, subscription := range subscriptions {

		resp, err := webpush.SendNotification([]byte(streamer), &subscription, &webpush.Options{
			Subscriber:		"test@leighmurray.com",
			VAPIDPublicKey:		vapidPublicKey,
			VAPIDPrivateKey:	vapidPrivateKey,
			TTL:			30,
		})

		if err != nil {
			fmt.Println(err)
			panic("couldn't send subscription")
		}

		defer resp.Body.Close()
	}
}

func main() {
	currentTime := time.Now()

	client := &http.Client{}

	req, _ := http.NewRequest("GET","http://127.0.0.1:8081/v1/vhosts/default/apps/app/streams", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic b21lLWFjY2Vzcy10b2tlbg==")

	resp, err := client.Do(req)

	if err != nil {
		panic("Couldn't access OME")
	}
	defer resp.Body.Close()

	var streamListResponse StreamListResponse
	err = json.NewDecoder(resp.Body).Decode(&streamListResponse)
	if err != nil {
		panic("Couldn't decode stream list response")
	}

	for _, streamer := range streamListResponse.Response {
	        req, _ := http.NewRequest("GET","http://127.0.0.1:8081/v1/stats/current/vhosts/default/apps/app/streams/" + streamer, nil)
	        req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Basic b21lLWFjY2Vzcy10b2tlbg==")

		resp, err := client.Do(req)

	        if err != nil {
		        panic("Couldn't access OME")
	        }
		defer resp.Body.Close()
		var streamInfoResponse StreamInfoResponse
		err = json.NewDecoder(resp.Body).Decode(&streamInfoResponse)
		if err != nil {
			fmt.Println(err)
			panic("Couldn't decode stream info response")
		}
		streamCreatedTime := streamInfoResponse.Response.CreatedTime
		if currentTime.Sub(streamCreatedTime).Seconds() < 60 {
			// if the stream was created within the last 60 seconds, create a notification
			sendNotifications(streamer)
		}
	}

}

