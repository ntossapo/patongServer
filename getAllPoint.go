package main

import (
	"fmt"
	"math/rand"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "strings"
    "strconv"
    "net/http"
    //"time"
    "io/ioutil"
    "encoding/json"
)

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

type Position struct{
	Lat float64			`bson:"lat"			json:"lat"`
	Long float64		`bson:"long"		json:"long"`
	Count int 			`bson:"count"		json:"count"`
}

type Count struct{
	C int 				`bson:"count"		json:"count"`
}

func randInt(min int, max int) int {
    return min + rand.Intn(max-min)
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
		
		abResult = append(abResult, Position{float64(lat)/1E5, float64(lng)/1E5, 1})
	}

	return abResult
}

func FloatToString(inputFloat float64) string{
	return strconv.FormatFloat(inputFloat, 'f', 10, 64)
}

func main(){
	var originLat float64
	var originLong float64
	var destLat float64
	var destLong float64
	for ; true; {
		originLat = 7.7577217 + (float64(randInt(0, 4419259))/10000000.0)
		destLat = 7.7577217 + (float64(randInt(0, 4419259))/10000000.0)
		destLong = 98.2582515  + (float64(randInt(0, 189851))/1000000.0)
		originLong = 98.2582515  + (float64(randInt(0, 189851))/1000000.0)
		fmt.Println("Random : ", originLat, ", ", originLong, " to ", destLat, ", ", destLong)
		if (originLat > 7.9355061) && (originLat < 7.9145934){
			fmt.Println("Error originLat random")
			continue
		}
		if (destLat > 7.9355061) && (destLat < 7.9145934){
			fmt.Println("Error destLat random")
			continue
		}
		if(originLong > 98.310958) && (originLong < 98.2919729){
			fmt.Println("Error originLong random")
			continue
		}
		if(destLong > 98.310958) && (destLong < 98.2919729){
			fmt.Println("Error destLong random")
			continue
		}
		fmt.Println("URL : ", FloatToString(originLat), ", ", FloatToString(originLong), " to ", FloatToString(destLat), ", ", FloatToString(destLong))
		stringBuilder := []string{}
		stringBuilder = append(stringBuilder, "http://maps.googleapis.com/maps/api/directions/json")
		stringBuilder = append(stringBuilder, "?origin=")
		stringBuilder = append(stringBuilder, FloatToString(originLat))
		stringBuilder = append(stringBuilder, ",")
		stringBuilder = append(stringBuilder, FloatToString(originLong))
		stringBuilder = append(stringBuilder, "&destination=")
		stringBuilder = append(stringBuilder, FloatToString(destLat))
		stringBuilder = append(stringBuilder, ",")
		stringBuilder = append(stringBuilder, FloatToString(destLong))
		stringBuilder = append(stringBuilder, "&sensor=false&mode=driving&alternatives=true")
		url := strings.Join(stringBuilder,"")

		res,_ := http.Get(url)
		defer res.Body.Close()

		content, _ := ioutil.ReadAll(res.Body)

		var googleRoute GoogleRoute
		json.Unmarshal([]byte(content), &googleRoute)

		session, _ := mgo.Dial("localhost")
		defer session.Close()

		collection := session.DB("graph").C("point")
		pc := session.DB("graph").C("path")
		pos := Position{}
		for _, i := range googleRoute.Routes{
			result := DecodingPloyline(i.OverviewPolyline.Points)
			for _, j := range result{
				found := collection.Find(bson.M{"lat":j.Lat, "long":j.Long}).One(&pos)
				if found == nil{
					collection.Update(bson.M{"lat":j.Lat, "long":j.Long}, bson.M{"$inc" : bson.M{"count":1}})
					//fmt.Println("Increse ", j.Lat, ", ", j.Long, " Update to ", pos.Count+1)
				}else{
					j.Count = 1
					collection.Insert(&j)
					//fmt.Println("Insert ", j.Lat, ",", j.Long)
				}
				pc.Update(bson.M{}, bson.M{"currentpath":i.OverviewPolyline.Points})
			}
		}
		var c Count
		ccollection := session.DB("graph").C("count")
		ccollection.Update(bson.M{}, bson.M{"$inc" : bson.M{"count":1}})
		ccollection.Find(bson.M{}).One(&c)
		fmt.Println("Query Count : ", c.C)
		//time.Sleep(5*time.Second)
	}
}