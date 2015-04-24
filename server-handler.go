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
	Pos Position 		`bson:"pos"			json:"pos"`
	Type string			`bson:"atype" 		json:"atype"`
	Name string			`bson:"name" 		json:"name"`
	Tel string			`bson:"tel" 		json:"tel"`
	Desc string			`bson:"desc" 		json:"desc"`
	DateTime string		`bson:"dateTime" 	json:"dateTime"`
}

type RouteReq struct{
	Origin	Position    `bson:"origin"		json:"origin"`
	Destination Position`bson:"destination"	json:"destination"`
}

type Position struct{
	Lat float64			`bson:"lat"			json:"lat"`
	Long float64		`bson:"long"		json:"long"`
}

type Block struct{
	Pos Position 		`bson:"position"			json:"position"`
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
	_MGOSERVER = _Loopback
	_Database = "patong"

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
		Position{
			StringToFloat(r.FormValue(_LatRq)),
			StringToFloat(r.FormValue(_LngRq))},
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

func getMongoCollection(s *mgo.Session, cname string) *mgo.Collection{
	return (*s).DB(_Database).C(cname)
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
	stringBuilder = append(stringBuilder, FloatToString(r.Origin.Lat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(r.Origin.Long))
	stringBuilder = append(stringBuilder, "&destination=")
	stringBuilder = append(stringBuilder, FloatToString(r.Destination.Lat))
	stringBuilder = append(stringBuilder, ",")
	stringBuilder = append(stringBuilder, FloatToString(r.Destination.Long))
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
	http.HandleFunc("/block", addBlock)
	http.HandleFunc("/allblock", getBlock)
}

func getAccidentPosition(w http.ResponseWriter, r *http.Request) {
	var list []Event
	ip := getIP(r)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, "event")

	collection.Find(bson.M{}).All(&list)
	ErrorHandler(err)

	result, _ := json.Marshal(list)

	fmt.Println("\t[LOG] -> Get All Event Called from :",ip)
	fmt.Fprintf(w, "%s", string(result))
	
}

func getBlock(w http.ResponseWriter, r *http.Request){
	var list []Block
	ip := getIP(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, "block")
	collection.Find(bson.M{}).All(&list)
	ErrorHandler(err)

	result, _ := json.Marshal(list)
	fmt.Println("\t[LOG] -> Get All Block Called from :",ip)
	fmt.Fprintf(w, "%s", string(result))
}

//{"origin":{"lat":?, "long":?}, "destination":{"lat":?, "long":?}}
func getBestPath(w http.ResponseWriter, r *http.Request){

	stringReqParam := r.FormValue("data")
	reqParam := &RouteReq{}

	w.Header().Set("Content-Type", "text/html; charset=utf-8r")

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

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, "block")
	end := false
	pos := Position{}

	fmt.Println("\t[LOG] -> ", "Get Routes " , len(googleRoute.Routes) , " Ways")
	for c, j := range googleRoute.Routes{
		if end {
			break
		}
		result := DecodingPloyline(j.OverviewPolyline.Points)
		size := len(result)
		for i, k := range result {
			fmt.Printf("\t[LOG] -> lat:%f-%f, long:%f-%f\n", (k.Lat-0.00005), (k.Lat+0.00005), (k.Long-0.00005), (k.Long+0.00005))
			found := collection.Find(bson.M{"position.lat":bson.M{"$gt":(k.Lat-0.00005), "$lt":(k.Lat+0.00005)}, "position.long":bson.M{"$gt":(k.Long-0.00005), "$lt":(k.Long+0.00005)}}).One(&pos)
			if found == nil{
				fmt.Println("\t[LOG] -> ", k.Lat, ", ", k.Long, " In Range. Next Ways")
				break
			}else if i == size-1{
				fmt.Println("\t[LOG] -> ", c, " is Work!!")
				fmt.Fprintf(w, "<html>")
				fmt.Fprintf(w, "<head>")
				fmt.Fprintf(w, "<META HTTP-EQUIV=\"Refresh\" CONTENT=\"0;URL=http://localhost/map.html?%s\">", j.OverviewPolyline.Points)
				fmt.Fprintf(w, "</head>")
				fmt.Fprintf(w, "</html>")
				end = true
				break
			}else if found.Error() == "not found"{
				continue
			}
		}
		
	}
	
}

//{position:{lat:?, long:?}}
func addBlock(w http.ResponseWriter, r *http.Request){
	jsonRequestPosition := r.FormValue("position")
	position := Position{}
	json.Unmarshal([]byte(jsonRequestPosition), &position)
	block := Block{position}

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, "block")
	collection.Insert(&block);
	ErrorHandler(err);
	fmt.Printf("\t[LOG] -> add new block mode 1 {lat:%f, long:%f}\n",
	 block.Pos.Lat,
	  block.Pos.Long)

	printJsonBool(&w, true)
}


func addAccidentPosition(w http.ResponseWriter, r *http.Request){
	event := extractDataFromRequest(r)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, "event")
	collection.Insert(&event)
	ErrorHandler(err);

	fmt.Printf("\t[LOG] -> add new accident \n{lat : %f, long : %f, aType : %d , name : \"%s\" ,tel : \"%s\" , desc : \"%s\" dateTime : \"%s\"}\n",
		event.Pos.Lat,
		event.Pos.Long,
		event.Type,
		event.Name,
		event.Tel,
		event.Desc,
		event.DateTime)

	printJsonBool(&w, true)
}