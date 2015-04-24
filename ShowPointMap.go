package main

import(
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "net"
)


func main() {
	http.HandleFunc("/point", getPoint)
	http.HandleFunc("/path", getPath)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

type Position struct{
	Lat float64			`bson:"lat"			json:"lat"`
	Long float64		`bson:"long"		json:"long"`
	Count int 			`bson:"count"		json:"count"`
}

type Count struct{
	C int 				`bson:"count"		json:"count"`
}

type Path struct{
	Cpath string		`bson:"currentpath"		json:"currentpath"`
}

func getIP(r *http.Request) string{
	result, _, _ := net.SplitHostPort(r.RemoteAddr)
	return result
}

func getPoint(w http.ResponseWriter, r *http.Request){
	var list []Position
	ip := getIP(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	session, _ := mgo.Dial("localhost")
	defer session.Close()

	var c Count
	ccollection := session.DB("graph").C("count")
	ccollection.Find(bson.M{}).One(&c)

	countSelection := int(float64(c.C)*0.05)

	fmt.Println("\t[LOG] -> Data Called gte", countSelection, " from ip ", ip, " count Query = ", c.C)
	collection := session.DB("graph").C("point")
	collection.Find(bson.M{"count":bson.M{"$gte":10, "$lte":30}}).All(&list)
	fmt.Printf("\t[LOG] -> Result Length ", len(list))
	result, _ := json.Marshal(list)
	fmt.Fprintf(w, "%s", string(result))
}

func getPath(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	session, _ := mgo.Dial("localhost")
	defer session.Close()
	collection := session.DB("graph").C("path")
	var p Path
	collection.Find(bson.M{}).One(&p)
	fmt.Printf("\t[LOG] -> Get Path ", p.Cpath)
	result, _ := json.Marshal(p)
	fmt.Fprintf(w, "%s", string(result))
}

