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

type Event struct{
	Lat float64			`bson:"lat" 		json:"lat"`
	Lng float64			`bson:"lng" 		json:"lng"`
	Type string		`bson:"atype" 		json:"atype"`
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

const(
	_Loopback = "localhost"
	_Database = "patong"
	_Collection = "accident"

	_LatRq = "lat"
	_LngRq = "long"
	_ATypeRq = "aType"
	_NameRq = "name"
	_TelRq = "tel"
	_DesRq = "desc"
	_DTRq = "dateTime"

)


func FloatToString(inputFloat float64) string{
	return strconv.FormatFloat(inputFloat, 'f', 6, 64)
}

func StringToFloat(s string) float64{
	result, err := strconv.ParseFloat(s, 64)
	ErrorHandler(err)
	return result
}

func extractDataFromRequest(r *http.Request) Event{
	return Event{
		StringToFloat(r.FormValue(_LatRq)),
		StringToFloat(r.FormValue(_LngRq)),
	  	r.FormValue(_ATypeRq),
	  	r.FormValue(_NameRq),
	  	r.FormValue(_TelRq),
	  	r.FormValue(_DesRq),
	  	r.FormValue(_DTRq)}
}

func ErrorHandler(err error){
	if err != nil{
		panic(err)
	}
}

func mongoDial() *mgo.Session{
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	return session
}

func getMongoCollection(s *mgo.Session) *mgo.Collection{
	return (*s).DB(_Database).C(_Collection)
}

func printJsonBool(w *http.ResponseWriter, b bool){
	result, err := json.Marshal(true)
	ErrorHandler(err)
	fmt.Fprintf(*w, "%s", string(result))
}

func getIP(r *http.Request) string{
	result, _, _ := net.SplitHostPort(r.RemoteAddr)
	return result
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

func ProgramLabel(){
	fmt.Println("Fire Engine Navigation System Version 0.1")
	fmt.Println("[MAIN] -> Server Start @ port", 8080)
}

func BuildApiUrl(r *RouteReq) string{
	stringBuilder := []string{}
	stringBuilder = append(stringBuilder, "http://maps.googleapis.com/maps/api/directions/json")
	stringBuilder = append(stringBuilder, "?origin=")
	stringBuilder = append(stringBuilder, FloatToString(r.OriLat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(r.OriLong))
	stringBuilder = append(stringBuilder, "&destination=")
	stringBuilder = append(stringBuilder, FloatToString(r.DestLat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(r.DestLong))
	stringBuilder = append(stringBuilder, "&sensor=false&mode=driving&alternatives=true")
	return strings.Join(stringBuilder,"")
}

/**
	MAIN AND ROUTING FUNCTION
**/
func main() {
	ProgramLabel()
	Rounting()
	log.Fatal(http.ListenAndServe(":8080", nil))
}	

func Rounting(){
	http.HandleFunc("/add", addAccidentPosition)
	http.HandleFunc("/get", getAccidentPosition)
	http.HandleFunc("/route", getBestPath)
}

func getAccidentPosition(w http.ResponseWriter, r *http.Request) {
	var List []Event
	ip := getIP(r)
	
	session := mongoDial()
	collection := getMongoCollection(session)

	err := collection.Find(bson.M{}).All(&List)
	ErrorHandler(err)

	result, _ := json.Marshal(List)

	fmt.Println("\t[LOG] -> Get Data Called from :",ip)
	fmt.Fprintf(w, "%s", string(result))
	
}

func getBestPath(w http.ResponseWriter, r *http.Request){

	stringReqParam := r.FormValue("data")
	reqParam := &RouteReq{}

	err := json.Unmarshal([]byte(stringReqParam), &reqParam)
	ErrorHandler(err)

	url := BuildApiUrl(reqParam)

	fmt.Println("\t[LOG] -> get route url : " + url)

	res,err := http.Get(url)
	ErrorHandler(err)
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	ErrorHandler(err)

	var googleRoute GoogleRoute
	json.Unmarshal([]byte(content), &googleRoute)
	ErrorHandler(err)

	/* show all route
	for i, j := range googleRoute.Routes{
		result := DecodingPloyline(j.OverviewPolyline.Points)
		fmt.Fprintf(w, "%d===================\n", i)
		for l, k := range result {
			fmt.Fprintf(w, "%d : %f, %f\n", l, k.Lat, k.Long)
		}
		fmt.Fprintf(w, "===================\n")
	}
	*/
}


func addAccidentPosition(w http.ResponseWriter, r *http.Request){
	event := extractDataFromRequest(r)
	session := mongoDial()
	collection := getMongoCollection(session)
	err := collection.Insert(&event)
	ErrorHandler(err);

	fmt.Printf("\t[LOG] -> add new accident \n{lat : %f, long : %f, aType : %d , name : \"%s\" ,tel : \"%s\" , desc : \"%s\" dateTime : \"%s\"}\n",
		event.Lat,
		event.Lng,
		event.Type,
		event.Name,
		event.Tel,
		event.Desc,
		event.DateTime)

	printJsonBool(&w, true)
}