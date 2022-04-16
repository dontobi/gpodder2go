package apis

import (
	"os"
	"strconv"
	"time"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oxtyped/gpodder2go/pkg/data"

	"github.com/augurysys/timestamp"
)

type Pair struct {
	a, b interface{}
}

func (p Pair) String() string {
	output := fmt.Sprintf("[%q, %q]", p.a, p.b)
	return output
}

type PairArray struct {
	Pairs []Pair
}

func (p PairArray) String() string {
	astring := "["
	for idx, v := range p.Pairs {
		astring += v.String()
		if idx != (len(p.Pairs) - 1) {
			astring += ","
		}

	}
	astring += "]"

	return astring
}

// UserAPI
func (u *UserAPI) HandleLogin(w http.ResponseWriter, r *http.Request) {
	//db := u.Data

	//token, err := db.RetrieveLoginToken(username, password)
	//if err != nil {

	//	w.WriteHeader(400)
	//	return
	//}

	//cookie := &http.Cookie{
	//	Name:  "sessionid",
	//	Value: token,
	//}
	//http.SetCookie(w, cookie)

	return

}

// DeviceAPI
func (d *DeviceAPI) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid

	username := chi.URLParam(r, "username")
	deviceName := chi.URLParam(r, "deviceid")

	ddr := &DeviceDataRequest{}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body from payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	// onboard the new directory

	err = json.Unmarshal(payload, ddr)
	if err != nil {
		log.Printf("error decoding json payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	log.Printf("DDR is %#v and %#v %#v", ddr, username, deviceName)
	err = d.Data.AddDevice(username, deviceName, ddr.Caption, ddr.Type)
	if err != nil {
		log.Printf("error adding device: %#v", err)
		w.WriteHeader(400)
		return
	}

	// 200
	// 401
	// 404
	// 400
	w.WriteHeader(200)
	return

}

func (d *DeviceAPI) HandleGetDevices(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")
	_, err := d.Data.RetrieveDevices(username)
	if err != nil {
		w.WriteHeader(200)
		return
	}

	return

}

// TODO: Handle Device Subscription Change
func (s *SubscriptionAPI) HandleDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid
	// format
	w.WriteHeader(404)
	return

}

// API Endpoint: GET /api/2/subscriptions/{username}/{deviceid}.json
func (s *SubscriptionAPI) HandleGetDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid
	// format
	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")
	add := []string{}
	remove := []string{}

	since := r.URL.Query().Get("since")
	if since == "" {
		log.Println("error with since query params - expecting it not to be empty but got \"\"")
		w.WriteHeader(400)
		return

	}
	log.Printf("since is %#v", since)

	if format != "json" {
		log.Printf("error uploading device subscription changes as format is expecting JSON but got %#v", format)
		w.WriteHeader(400)
		return
	}

	subscriptionChanges := &SubscriptionChanges{
		Add:       add,
		Remove:    remove,
		Timestamp: timestamp.Now(),
	}

	db := s.Data
	tm := time.Time{}

	if since == "0" {
		tm = time.Unix(0, 0)
	} else {
		i, err := strconv.ParseInt(since, 10, 64)
		if err != nil {
			log.Printf("error parsing strconv: %#v", err)
			w.WriteHeader(400)
			return
		}

		tm = time.Unix(i, 0)
	}

	subs, err := db.RetrieveSubscriptionHistory(username, deviceId, tm)
	if err != nil {
		log.Printf("error retrieving subscription history: %#v", err)

		w.WriteHeader(400)
		return
	}

	add, remove = data.SubscriptionDiff(subs)

	subscriptionChanges.Add = add
	subscriptionChanges.Remove = remove

	log.Printf("add is %#v and remove is %#v", add, remove)

	outputPayload, err := json.Marshal(subscriptionChanges)
	if err != nil {
		log.Printf("error marshalling subscription changes into JSON string: %#v", err)
		w.WriteHeader(400)
		return
	}

	log.Printf("outputPayload is %#v", string(outputPayload))
	w.WriteHeader(200)
	w.Write(outputPayload)
	return

}

// API Endpoint: POST /api/2/subscriptions/{username}/{deviceid}.{format}
func (s *SubscriptionAPI) HandleUploadDeviceSubscriptionChange(w http.ResponseWriter, r *http.Request) {

	// username
	// deviceid
	// format
	// add (slice)
	// remove (slice)
	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")

	if format != "json" {
		log.Printf("error uploading device subscription changes as format is expecting JSON but got %#v", format)
		w.WriteHeader(400)
		return
	}

	subscriptionChanges := &SubscriptionChanges{}
	err := json.NewDecoder(r.Body).Decode(&subscriptionChanges)
	if err != nil {

		log.Printf("error decoding json payload: %#v", err)
		w.WriteHeader(400)
		return
	}

	log.Printf("subscription changes is %#v", subscriptionChanges)

	addSlice := subscriptionChanges.Add
	removeSlice := subscriptionChanges.Remove

	ts := data.CustomTimestamp{}
	ts.Time = time.Now()

	db := s.Data

	pairz := []Pair{}
	for _, v := range addSlice {
		sub := data.Subscription{
			User:      username,
			Device:    deviceId,
			Podcast:   v,
			Timestamp: ts,
			Action:    "SUBSCRIBE",
		}
		pair := Pair{v, v}
		pairz = append(pairz, pair)
		err := db.AddSubscriptionHistory(sub)
		if err != nil {
			log.Printf("error adding subscription: %#v", err)
		}
	}

	for _, v := range removeSlice {
		sub := data.Subscription{
			User:      username,
			Device:    deviceId,
			Podcast:   v,
			Timestamp: ts,
			Action:    "UNSUBSCRIBE",
		}
		pair := Pair{v, v}
		pairz = append(pairz, pair)
		db.AddSubscriptionHistory(sub)
	}

	pp := PairArray{pairz}

	subscriptionChangeOutput := &SubscriptionChangeOutput{
		Timestamp:  timestamp.Time(ts.Time),
		UpdateUrls: json.RawMessage(pp.String()),
	}

	outputBytes, err := json.Marshal(subscriptionChangeOutput)
	if err != nil {
		log.Printf("error marshalling output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	log.Printf("outputbytes is %#v", string(outputBytes))
	w.Write(outputBytes)
	return

}

func (s *SubscriptionAPI) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")
	xml, err := s.Data.RetrieveAllDeviceSubscriptions(username)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	log.Printf("output: %#v", xml)
	w.Write([]byte(xml))

}

func (s *SubscriptionAPI) HandleGetDeviceSubscription(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")

	xml, err := s.Data.RetrieveDeviceSubscriptions(username, deviceId)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Write([]byte(xml))
	w.WriteHeader(200)

}

// API Endpoint: POST and PUT /subscriptions/{username}/{deviceid}.{format}
func (s *SubscriptionAPI) HandleUploadDeviceSubscription(w http.ResponseWriter, r *http.Request) {
	// username
	// deviceid
	// format

	username := chi.URLParam(r, "username")
	deviceId := chi.URLParam(r, "deviceid")
	format := chi.URLParam(r, "format")
	ts := data.CustomTimestamp{}
	ts.Time = time.Now()

	log.Printf("%v, %v, %v", username, deviceId, format)

	switch r.Method {
	case "POST":
		log.Println("Receive a POST")
	case "PUT":
		// Upload entire subscriptions

		// TODO: need to handle all the different formats, json, xml, text etc

		log.Println("Receive a PUT")
		log.Printf("Saving subscription...")

		b, _ := ioutil.ReadAll(r.Body)

		var arr []string
		err := json.Unmarshal(b, &arr)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(fmt.Sprintf("%s-%s.%s", username, deviceId, format))
		if err != nil {
			log.Printf("error saving file: %#v", err)
			w.WriteHeader(400)
			return
		}
		defer f.Close()

		// start to write each line by line
		for _, v := range arr {

			sub := data.Subscription{
				User:      username,
				Device:    deviceId,
				Podcast:   v,
				Action:    "SUBSCRIBE",
				Timestamp: ts,
			}
			s.Data.AddSubscriptionHistory(sub)
			f.WriteString(v + "\n")
		}

		w.WriteHeader(200)
		return
	default:
		w.WriteHeader(400)
		return

	}

}

// EpisodeAPI

func (e *EpisodeAPI) HandleEpisodeAction(w http.ResponseWriter, r *http.Request) {

	//username
	//format - defaulting to "json" as per spec
	//username := chi.URLParam(r, "username")

	// body:
	// podcast (string) optional
	// device (string) optional
	// since (int) optional also, if no actions, then release all
	// aggregated (bool)

	episodeActionOutput := &EpisodeActionOutput{
		Actions:   []data.EpisodeAction{},
		Timestamp: timestamp.Now(),
	}

	episodeActionOutputBytes, err := json.Marshal(episodeActionOutput)
	if err != nil {
		log.Printf("error marshalling episodes actions output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	log.Printf("%#v", string(episodeActionOutputBytes))
	w.Write(episodeActionOutputBytes)
	return

}

// POST /api/2/episodes/{username}.json
func (e *EpisodeAPI) HandleUploadEpisodeAction(w http.ResponseWriter, r *http.Request) {
	//username

	username := chi.URLParam(r, "username")
	ts := time.Now()

	b, _ := ioutil.ReadAll(r.Body)

	log.Printf("output is strings: %#v", string(b))
	var arr []data.EpisodeAction

	pairz := []Pair{}

	err := json.Unmarshal(b, &arr)
	if err != nil {
		log.Printf("error unmarshalling: %#v", err)
		w.WriteHeader(400)
		return
	}

	for _, data := range arr {
		log.Printf("episode user: %#v and data: %#v AND REAL DATA: %#v \n", username, data, e.Data)
		err := e.Data.AddEpisodeActionHistory(username, data)
		if err != nil {
			log.Printf("error adding episode action into history: %#v", err)
		}
		pair := Pair{
			data.Episode, data.Episode,
		}
		pairz = append(pairz, pair)
	}

	// format

	pp := PairArray{pairz}
	subscriptionChangeOutput := &SubscriptionChangeOutput{
		Timestamp:  timestamp.Time(ts),
		UpdateUrls: json.RawMessage(pp.String()),
	}

	outputBytes, err := json.Marshal(subscriptionChangeOutput)
	if err != nil {
		log.Printf("error marshalling output: %#v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	log.Printf("outputbytes is %#v", string(outputBytes))
	w.Write(outputBytes)
}