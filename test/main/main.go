package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	engines "github.com/tapvanvn/godbengine"
	"github.com/tapvanvn/godbengine/engine"
	"github.com/tapvanvn/godbengine/engine/adapter"
	"github.com/tapvanvn/gosession"
	"github.com/tapvanvn/gosession/test"
	"github.com/tapvanvn/goutil"
)

var Config *test.Config = nil
var Provider *gosession.Provider = nil
var Validator *gosession.Validator = nil

var ChnSession chan string = make(chan string, 1000)
var PrepareSession chan string = make(chan string, 1000)

var TestSessionNum = 5
var SessionNum = 0

var TotalQuota = int64(100)

func startEngine(eng *engine.Engine) {

	//read redis define from env
	var redis *adapter.RedisPool = nil

	if Config.MemDB != nil {

		redisConnectString := Config.MemDB.ConnectionString

		redisPool := adapter.RedisPool{}

		err := redisPool.Init(redisConnectString)

		if err != nil {

			fmt.Println("cannot init redis")
		}
		redis = &redisPool
	}
	var documentDB engine.DocumentPool = nil

	if Config.DocumentDB != nil {

		connectString := Config.DocumentDB.ConnectionString
		databaseName := Config.DocumentDB.DatabaseName

		if Config.DocumentDB.Provider == "mongodb" {

			mongoPool := &adapter.MongoPool{}

			err := mongoPool.InitWithDatabase(connectString, databaseName)

			if err != nil {

				log.Fatal("cannot init mongo")
			}
			documentDB = mongoPool

		} else {

			firestorePool := adapter.FirestorePool{}
			firestorePool.Init(connectString)
			documentDB = &firestorePool
		}
	}

	eng.Init(redis, documentDB, nil)
}

func processAction() {
	for {
		sessionString := <-ChnSession
		action := rand.Intn(5)
		fmt.Println(sessionString, action)
		err := Validator.ValidateAction(sessionString, "testAgent", action)
		if err == nil {
			PrepareSession <- sessionString
		} else {
			fmt.Println(sessionString, err)
		}
	}
}

func forwardAction() {
	for {
		sessionString := <-PrepareSession
		ChnSession <- sessionString
	}
}

func autoGen() {
	if SessionNum > TestSessionNum {
		return
	}
	SessionNum++
	if sessionString, err := Provider.IssueSessionString("testAgent"); err != nil {

		log.Fatal(err)

	} else {

		fmt.Println(sessionString)

		if err := Validator.Validate(sessionString, "testAgent"); err != nil {

			log.Fatal(err)
		}
		ChnSession <- sessionString
	}
}

func main() {

	file, err := os.Open("config.jsonc")
	if err != nil {
		log.Fatal(err)
	}
	configData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	configData = goutil.TripJSONComment(configData)
	config := &test.Config{}
	err = json.Unmarshal(configData, config)
	if err != nil {
		log.Fatal(err)
	}
	Config = config

	engines.InitEngineFunc = startEngine
	eng := engines.GetEngine()

	err = gosession.Init(nil, eng)

	if err != nil {
		log.Fatal(err)
	}

	provider, err := gosession.NewProvider()
	if err != nil {
		log.Fatal(err)
	}
	Provider = provider

	validator, err := gosession.NewValidator(TotalQuota)
	if err != nil {
		log.Fatal(err)
	}

	Validator = validator
	validator.AddActionQuota(0, 5)
	validator.AddActionQuota(1, 5)
	validator.AddActionQuota(2, 5)
	validator.AddActionQuota(3, 5)
	validator.AddActionQuota(4, -1)

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {

		if sessionString, err := Provider.IssueSessionString("testAgent"); err != nil {
			w.Write([]byte(sessionString))
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	//case Test Generate session
	goutil.Schedule(autoGen, time.Millisecond*100)

	go processAction()
	go forwardAction()

	log.Fatal(http.ListenAndServe(":9000", nil))
}
