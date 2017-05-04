package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/couchbase/gocb"
	"github.com/gorilla/mux"
	hashids "github.com/speps/go-hashids"
)

type MyUrl struct {
	ID       string `json:"id,omitempty"`
	LongUrl  string `json:"longUrl,omitempty"`
	ShortUrl string `json:"ShortUrl,omitempty"`
}

var bucket *gocb.Bucket
var bucketName string

func ExpandEndpoint(w http.ResponseWriter, req *http.Request) {

}
func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	var url MyUrl
	_ = json.NewDecoder(req.Body).Decode(&url)
	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, url.LongUrl)
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE longUrl = $1")
	rows, err := bucket.ExecuteN1qlQuery(query, n1qlParams)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	var row MyUrl
	rows.One(&row)
	if row == (MyUrl{}) {
		hd := hashids.NewData()
		h := hashids.NewWithData(hd)
		now := time.Now()
		url.ID, _ = h.Encode([]int{int(now.Unix())})
		url.ShortUrl = "http://localhost:12345/" + url.ID
		bucket.Insert(url.ID, url, 0)
	} else {
		url = row
	}
	json.NewEncoder(w).Encode(url)
}
func RootEndpoint(w http.ResponseWriter, req *http.Request) {

}

func main() {
	router := mux.NewRouter()
	cluster, _ := gocb.Connect("couchbase://127.0.0.1")
	bucketName = "default"
	bucket, _ = cluster.OpenBucket(bucketName, "")
	router.HandleFunc("/create", CreateEndpoint).Methods("POST")
	router.HandleFunc("/expand/", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":12345", router))
}
