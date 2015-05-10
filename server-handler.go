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

type Event2 struct{
	Pos Position 		`bson:"pos"			json:"pos"`
	Type string			`bson:"type" 		json:"type"`
	Name string			`bson:"name" 		json:"name"`
	InformerId string	`bson:"informerid" 	json:"informerid"`
	Telnum string 		`bson:"telnum" 		json:"telnum"`
	Desc string			`bson:"desc" 		json:"desc"`
	DateTime string		`bson:"dateTime" 	json:"dateTime"`
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Status string 		`bson:"status" 		json:"status"`
	Option string 		`bson:"option" 		json:"option"`
	Address string 		`bson:"address" 	json:"address"`
	Place string 		`bson:"place" 		json:"place"`
	Image string 		`bson:"image" 		json:"image"`
}

type Event struct{
	Pos Position 		`bson:"pos"			json:"pos"`
	Type string			`bson:"type" 		json:"type"`
	Name string			`bson:"name" 		json:"name"`
	InformerId string	`bson:"informerid" 	json:"informerid"`
	Desc string			`bson:"desc" 		json:"desc"`
	DateTime string		`bson:"dateTime" 	json:"dateTime"`
	Status string 		`bson:"status" 		json:"status"`
	Option string 		`bson:"option" 		json:"option"`
	Telnum string 		`bson:"telnum" 		json:"telnum"`
	Address string 		`bson:"address" 	json:"address"`
	Place string 		`bson:"place" 		json:"place"`
	Image string 		`bson:"image" 		json:"image"`
}

type RouteReq struct{
	Origin	Position    `bson:"origin"		json:"origin"`
	Destination Position`bson:"destination"	json:"destination"`
}

type Position struct{
	Lat float64			`bson:"lat"			json:"lat"`
	Long float64		`bson:"long"		json:"long"`
}

//{"pos":{"lat":1.0, "long":1.0}, "score":1.0}
type Block struct{
	Pos Position 		`bson:"pos"				json:"pos"`
	Score int			`bson:"score"			json:"score"`
}

type DataVersion struct{
	EventVersion int 	`bson:"eventversion"		json:"eventversion"`
	BlockVersion int 	`bson:"blockversion"		json:"blockversion"`
}

type ServerStatusResponse struct{
	ServerStatus string 		`bson:"status"	json:"status"`
	DataVersion DataVersion 	`bson:"dataversion"		json:"dataversion"`
}

type UserResponse struct{
	ServerStatus string 	`bson:"status"	json:"status"`
	UserData User2			`bson:"user"	json:"user"`
}

type PathResponse struct{
	Status string 				`bson:"status"	json:"status"`
	Paths Path 					`bson:"paths"	json:"path"`
}

type Path struct{
	Score int 					`bson:"score"	json:"score"`
	Path string 				`bson:"path"	json:"path"`
}

type EventRespose struct{
	Status string 			`bson:"status"	json:"status"`
	Eve Event2 				`bson:"event"	json:"event"`
}

type User struct{
	Telnum string 		`bson:"telnum"	json:"telnum"`
}

type User2 struct{
	Telnum	string 		`bson:"telnum"	json:"telnum"`
	Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
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

	_MgoCollectionEvent = "event"
	_MgoCollectionBlock = "block"
	_MgoCollectionDataVersion = "dataversion"
	_MgoCollectionUser = "user"
)


func FloatToString(inputFloat float64) string{
	return strconv.FormatFloat(inputFloat, 'f', 6, 64)
}

func StringToFloat(s string) float64{
	result, err := strconv.ParseFloat(s, 64)
	ErrorHandler(err)
	return result
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
	fmt.Println("Fire Engine Navigation System Version 1.0")
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

func UpdateDataVersion(v string){
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	dataVersionCollection := getMongoCollection(session, _MgoCollectionDataVersion)
	if v == "event"{
		dataVersionCollection.Update(bson.M{}, bson.M{"$inc":bson.M{"eventversion":1}})
	}else if v == "block"{
		dataVersionCollection.Update(bson.M{}, bson.M{"$inc":bson.M{"blockversion":1}})
	}
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
	//server status
	http.HandleFunc("/server", checkServerStatus)

	//user
	http.HandleFunc("/useradd", adduser)

	//event
	http.HandleFunc("/eventget", getevent)
	http.HandleFunc("/eventadd", addevent)
	http.HandleFunc("/eventupdate", updateEvent)
	http.HandleFunc("/eventremove", deleteEvent)

	http.HandleFunc("/route", getBestPath)
	http.HandleFunc("/blockadd", addBlock)
	http.HandleFunc("/blockget", getBlock)
}

func deleteEvent(w http.ResponseWriter, r *http.Request){
	jsonString := r.FormValue("data")
	event := Event2{}
	err := json.Unmarshal([]byte(jsonString), &event)
	ErrorHandler(err)
	
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	eventCollection := getMongoCollection(session, _MgoCollectionEvent)
	notfound := eventCollection.Remove(bson.M{"_id":event.Id})
	ErrorHandler(notfound)
	if notfound != nil{
		result, _ := json.Marshal(EventRespose{"error : event not found", Event2{}})
		fmt.Fprintf(w, "%s", string(result))
	} else {
		result, _ := json.Marshal(EventRespose{"success : remove success", event})
		fmt.Fprintf(w, "%s", string(result))
	}
	fmt.Println("\t[LOG] -> deleteEvent Called")
	UpdateDataVersion("event")
}

//{"Pos":{"Lat":1,"Long":1},"Type":"1","Name":"test","InformerId":"test","Desc":"test","DateTime":"29/4/2558 08:43","id":"55403b9730a438619bee54bf","Status":"test"}
func updateEvent(w http.ResponseWriter, r *http.Request){
	jsonString := r.FormValue("data")
	event := Event2{}
	temp := Event2{}
	err := json.Unmarshal([]byte(jsonString), &event)
	ErrorHandler(err)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	eventCollection := getMongoCollection(session, _MgoCollectionEvent)
	notfound := eventCollection.Find(bson.M{"_id":event.Id}).One(&temp)

	fmt.Println(notfound, event.Id)
	if notfound != nil{
		result, _ := json.Marshal(EventRespose{"error : event not found", Event2{}})
		fmt.Fprintf(w, "%s", string(result))
	} else {
		eventCollection.Update(bson.M{"_id":event.Id}, bson.M{"pos":bson.M{"lat":event.Pos.Lat, "long":event.Pos.Long}, "type":event.Type, "name":event.Name, "informid":event.InformerId, "desc":event.Desc, "datetime":event.DateTime, "status":event.Status})
		eventCollection.Find(bson.M{"_id":event.Id}).One(&temp)
		result, _ := json.Marshal(EventRespose{"success : update success", temp})
		fmt.Fprintf(w, "%s", string(result))
	}
	fmt.Println("\t[LOG] -> updateEvent Called")
	UpdateDataVersion("event")
}

//{telnum"?"}
func adduser(w http.ResponseWriter, r *http.Request){
	jsonString := r.FormValue("data")
	user := User{}
	temp := User2{}
	json.Unmarshal([]byte(jsonString), &user)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	userCollection := getMongoCollection(session, _MgoCollectionUser)
	found := userCollection.Find(bson.M{"telnum":user.Telnum}).One(&temp)
	fmt.Println(found)
	if found != nil{
		userCollection.Insert(&user)
		user2 := User2{}
		userCollection.Find(bson.M{"telnum":user.Telnum}).One(&user2)
		result, _ := json.Marshal(UserResponse{"success : register success", user2})
		fmt.Fprintf(w, "%s", string(result))
	} else {
		result, _ := json.Marshal(UserResponse{"success : already register", temp})
		fmt.Fprintf(w, "%s", string(result))
	}
	fmt.Println("\t[LOG] -> adduser Called")
}

func checkServerStatus(w http.ResponseWriter, r *http.Request){
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	var dataversion DataVersion
	dataVersionCollection := getMongoCollection(session, _MgoCollectionDataVersion)
	found := dataVersionCollection.Find(bson.M{}).One(&dataversion)
	if found != nil {
		result, _ := json.Marshal(ServerStatusResponse{"error : database error", DataVersion{-1, -1}})
		fmt.Fprintf(w, "%s", string(result))
	} else {
		result, _ := json.Marshal(ServerStatusResponse{"online", dataversion})
		fmt.Fprintf(w, "%s", string(result))
	}
	fmt.Println("\t[LOG] -> checkServerStatus Called")
}

func getevent(w http.ResponseWriter, r *http.Request) {
	var list []Event2
	ip := getIP(r)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, _MgoCollectionEvent)

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

	collection := getMongoCollection(session, _MgoCollectionEvent)
	collection.Find(bson.M{}).All(&list)
	ErrorHandler(err)

	result, _ := json.Marshal(list)
	fmt.Println("\t[LOG] -> Get All Block Called from :",ip)
	fmt.Fprintf(w, "%s", string(result))
}

//{"origin":{"lat":7.8931394, "long":98.3536982}, "destination":{"lat":7.8958174, "long":98.351649}}
func getBestPath(w http.ResponseWriter, r *http.Request){
	ip := getIP(r)
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

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, _MgoCollectionBlock)
	pos := Block{}
	var allPaths []Path
	score := 0

	for _, j := range googleRoute.Routes{
		result := DecodingPloyline(j.OverviewPolyline.Points)
		size := len(result)
		for i, k := range result {
			//fmt.Printf("\t[LOG] -> lat:%f-%f, long:%f-%f\n", (k.Lat-0.00005), (k.Lat+0.00005), (k.Long-0.00005), (k.Long+0.00005))
			found := collection.Find(bson.M{"pos.lat":bson.M{"$gt":(k.Lat-0.00005), "$lt":(k.Lat+0.00005)}, "pos.long":bson.M{"$gt":(k.Long-0.00005), "$lt":(k.Long+0.00005)}}).One(&pos)
			if found == nil{
				score += pos.Score
			}else if i == size-1{
				fmt.Println("Add way to result")
				allPaths = append(allPaths, Path{score, j.OverviewPolyline.Points})
			}else if found != nil{
				continue
			}
			fmt.Println(found, " score ", score)
		}
		fmt.Println("========================================================")
		score = 0;
	}

	result, _ := json.Marshal(allPaths)
	fmt.Println("\t[LOG] -> Path Called from :",ip)
	fmt.Println("\t[LOG] -> ", "Get Routes " , len(googleRoute.Routes) , " Ways")
	fmt.Fprintf(w, "%s", string(result))
}

//{"pos":{"lat":7.8952595, "long":98.3534837}, "score":1}
func addBlock(w http.ResponseWriter, r *http.Request){
	jsonRequestPosition := r.FormValue("data")
	block := Block{}
	err := json.Unmarshal([]byte(jsonRequestPosition), &block)
	fmt.Println(jsonRequestPosition)

	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, _MgoCollectionBlock)
	collection.Insert(&block);
	ErrorHandler(err);
	fmt.Printf("\t[LOG] -> add new block mode %d {lat:%f, long:%f}\n", block.Score,
	 block.Pos.Lat,
	  block.Pos.Long)

	printJsonBool(&w, true)
	UpdateDataVersion("block")

}

//{"pos":{"lat":?, "long":?}, "type":"?", "name":"?", "InformerId":"?", "desc":"?", "dateTime":"?", "status":"?""}
func addevent(w http.ResponseWriter, r *http.Request){
	jsonString := r.FormValue("data")
	fmt.Println(jsonString)
	event := Event{}
	temp := Event2{}
	err := json.Unmarshal([]byte(jsonString), &event)
	ErrorHandler(err)

	fmt.Println(event)
	session, err := mgo.Dial(_Loopback)
	ErrorHandler(err)
	defer session.Close()

	collection := getMongoCollection(session, _MgoCollectionEvent)
	collection.Insert(&event)
	collection.Find(event).One(&temp)
	fmt.Println(temp)
	result, _ := json.Marshal(EventRespose{"success : insert success", temp})
	fmt.Fprintf(w, "%s", string(result))
	fmt.Println("\t[LOG] -> addEvent Called")
	UpdateDataVersion("event")
}