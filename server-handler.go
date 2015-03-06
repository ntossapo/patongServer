package main

import(
	"fmt"
	"net/http"
	"log"
	"strconv"
	"encoding/json"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "net"
    "strings"
    "io/ioutil"
)

type Accident struct{
	Lat float64			`bson:"lat" 		json:"lat"`
	Lng float64			`bson:"lng" 		json:"lng"`
	Atype string		`bson:"atype" 		json:"atype"`
	Name string			`bson:"name" 		json:"name"`
	Tel string			`bson:"tel" 		json:"tel"`
	Desc string			`bson:"desc" 		json:"desc"`
	DateTime string		`bson:"dateTime" 	json:"dateTime"`
}

type RouteReq struct{
	OriLat float64		`bson:"orilat" 		json:"orilat"`
	OriLong float64		`bson:"orilong" 	json:"orilong"`
	DestLat float64		`bson:"destlat" 	json:"destlat"`
	DestLong float64	`bson:"destlong 	json:"destlong"`
}

type Position struct{
	Lat float64
	Long float64
}

type GoogleRoute struct {
	Routes []struct {
		Bounds struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"southwest"`
		} `json:"bounds"`
		Copyrights string `json:"copyrights"`
		Legs       []struct {
			Distance struct {
				Text  string  `json:"text"`
				Value float64 `json:"value"`
			} `json:"distance"`
			Duration struct {
				Text  string  `json:"text"`
				Value float64 `json:"value"`
			} `json:"duration"`
			EndAddress  string `json:"end_address"`
			EndLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"end_location"`
			StartAddress  string `json:"start_address"`
			StartLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"start_location"`
			Steps []struct {
				Distance struct {
					Text  string  `json:"text"`
					Value float64 `json:"value"`
				} `json:"distance"`
				Duration struct {
					Text  string  `json:"text"`
					Value float64 `json:"value"`
				} `json:"duration"`
				EndLocation struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"end_location"`
				HtmlInstructions string `json:"html_instructions"`
				Polyline         struct {
					Points string `json:"points"`
				} `json:"polyline"`
				StartLocation struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"start_location"`
				TravelMode string `json:"travel_mode"`
			} `json:"steps"`
			ViaWaypoint []interface{} `json:"via_waypoint"`
		} `json:"legs"`
		OverviewPolyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
		Summary       string        `json:"summary"`
		Warnings      []interface{} `json:"warnings"`
		WaypointOrder []interface{} `json:"waypoint_order"`
	} `json:"routes"`
	Status string `json:"status"`
}


func FloatToString(inputFloat float64) string{
	return strconv.FormatFloat(inputFloat, 'f', 6, 64)
}

func addAccidentPosition(w http.ResponseWriter, r *http.Request){
	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 32)
	lng, _ := strconv.ParseFloat(r.FormValue("lng"), 32)
	aType := r.FormValue("aType")
	name := r.FormValue("name")
	tel := r.FormValue("tel")
	desc := r.FormValue("desc")
	dateTime := r.FormValue("dateTime")

	session, err := mgo.Dial("localhost")
	if err != nil{
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("patong").C("accident")
	err = c.Insert(&Accident{lat, lng, aType, name, tel, desc, dateTime})

	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("add new accident \n{lat : %f, long : %f, aType : %d , name : \"%s\" , tel : \"%s\" , desc : \"%s\" dateTime : \"%s\"}\n", lat, lng, aType, name, tel, desc, dateTime)
	bolResult, _ := json.Marshal(true)
	fmt.Fprintf(w, "%s", string(bolResult))
}

func getAccidentPosition(w http.ResponseWriter, r *http.Request) {
	var accident []Accident
	ip ,_ ,_ := net.SplitHostPort(r.RemoteAddr)
	session, err := mgo.Dial("localhost")
	defer session.Close();

	c := session.DB("patong").C("accident")
	err = c.Find(bson.M{}).All(&accident)
	if err != nil{
		panic(err)
	}
	dataOut, _ := json.Marshal(accident)
	fmt.Println("Get Data Called from :",ip)
	fmt.Fprintf(w, "%s", string(dataOut))
	
}

func getBestPath(w http.ResponseWriter, r *http.Request){
	/*stringOriginLat := r.FormValue("oriLat");
	stringOriginLong  := r.FormValue("oriLong");
	stringDestLat := r.FormValue("destLat");
	stringDestLong := r.FormValue("destLong");*/

	stringReqParam := r.FormValue("data")
	reqParam := &RouteReq{}
	if err := json.Unmarshal([]byte(stringReqParam), &reqParam) ; err != nil {
		panic(err)
	}

	fmt.Println(reqParam.OriLat)
		
	stringBuilder := []string{}
	stringBuilder = append(stringBuilder, "http://maps.googleapis.com/maps/api/directions/json")
	stringBuilder = append(stringBuilder, "?origin=")
	stringBuilder = append(stringBuilder, FloatToString(reqParam.OriLat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(reqParam.OriLong))
	stringBuilder = append(stringBuilder, "&destination=")
	stringBuilder = append(stringBuilder, FloatToString(reqParam.DestLat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(reqParam.DestLong))
	stringBuilder = append(stringBuilder, "&sensor=false&mode=driving&alternatives=true")
	
	url := strings.Join(stringBuilder,"")
	fmt.Println("get route url : " + url)

	res, _ := http.Get(url)
	defer res.Body.Close()
	content, _ := ioutil.ReadAll(res.Body)

	var googleRoute GoogleRoute

	if err := json.Unmarshal([]byte(content), &googleRoute) ; err != nil{
		panic(err)
	}

	for i, j := range googleRoute.Routes{
		result := DecodingPloyline(j.OverviewPolyline.Points)
		fmt.Fprintf(w, "%d===================\n", i)
		for l, k := range result {
			fmt.Fprintf(w, "%d : %f, %f\n", l, k.Lat, k.Long)
		}
		fmt.Fprintf(w, "===================\n")
	}
}

func DecodingPloyline(polylineString string) []Position{
	var abResult []Position
	index := 0
	len := len(polylineString)
	lat := 0
	lng := 0

	for ; index < len; {
		var b int
		var shift uint
		var result int
		for ; true ; {
			b = int(polylineString[index]) - 63;
			index++
			result |= (b & 0x1f) << shift
			shift += 5
			if !(b >= 0x20) {break}
		}
		var dlat int
		if (result & 1) != 0{
			dlat = ^(result >> 1)
		}else{
			dlat = (result >> 1)
		}
		lat += dlat

		shift = 0
		result = 0
		for ; true ; {
			b = int(polylineString[index]) - 63;
			index++
			result |= (b & 0x1f) << shift;
			shift += 5;
			if !(b >= 0x20) {break}
		}
		var dlng int
		if (result & 1) != 0 {
			dlng = ^(result >> 1)
		}else{
			dlng = (result >> 1)
		}
		lng += dlng
		
		abResult = append(abResult, Position{float64(lat)/1E5, float64(lng)/1E5})
	}

	return abResult
}

func main() {
	fmt.Println("Server Start @ port", 8080)
	http.HandleFunc("/add", addAccidentPosition)
	http.HandleFunc("/get", getAccidentPosition)
	http.HandleFunc("/route", getBestPath)
	log.Fatal(http.ListenAndServe(":8080", nil))
}	