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
)

type Accident struct{
	Lat float64			`bson:"lat" json:"lat"`
	Lng float64			`bson:"lng" json:"lng"`
	Atype string			`bson:"atype" json:"atype"`
	Name string			`bson:"name" json:"name"`
	Tel string			`bson:"tel" json:"tel"`
	Desc string			`bson:"desc" json:"desc"`
	DateTime string		`bson:"dateTime" json:"dateTime"`
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

func main() {
	fmt.Println("Server Start @ port", 8080)
	http.HandleFunc("/add", addAccidentPosition)
	http.HandleFunc("/get", getAccidentPosition)
	log.Fatal(http.ListenAndServe(":8080", nil))
}	