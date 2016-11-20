package db

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"../models"

	r "gopkg.in/dancannon/gorethink.v2"
)

var session *r.Session

const db = "brownbag"
const table = "votes"

func connect() {
	fmt.Println("Establishing connection ...")

	hosts := []string{"localhost:28015"}

	if os.Getenv("DB_HOSTS") != "" {
		hosts = strings.Split(os.Getenv("DB_HOSTS"), ",")
	}

	fmt.Println("Running RethinkDB on hosts", hosts)

	// New session connection
	var connectionErr error
	session, connectionErr = r.Connect(r.ConnectOpts{
		Addresses: hosts,
		Database:  db,
	})

	if connectionErr != nil {
		panic(connectionErr.Error())
	}

	fmt.Println("Connection established ...")
}

func close() {
	fmt.Println("Closing connection ...")

	closeErr := session.Close()

	if closeErr != nil {
		panic(closeErr.Error())
	}

	fmt.Println("Connection closed ...")
}

func create() {
	fmt.Println("Creating initial votes ...")
	connect()

	initVotes := []models.Vote{models.Vote{Name: "darkSide", Count: 0}, models.Vote{Name: "lightSide", Count: 0}}
	_, createErr := r.Table(table).Insert(initVotes).Run(session)

	if createErr != nil {
		panic(createErr.Error())
	}

	close()
}

func Init() {
	connect()

	// Check if database exists and create it if not
	dbCursor, dbErr := r.DBList().Contains(db).Do(func(term r.Term) r.Term {
		return r.Branch(term, r.WriteResponse{DBsCreated: 1}, r.DBCreate(db))
	}).Run(session)

	if dbErr != nil {
		panic(dbErr.Error())
	}

	// Get cursor documents
	var dbResult interface{}
	dbCursor.One(&dbResult)

	// Convert documents to map of string keys and interface values
	dbResultMap, _ := dbResult.(map[string]interface{})

	// Get array of interfaces for config_changes key
	dbConfigChange := dbResultMap["config_changes"].([]interface{})

	if len(dbConfigChange) > 0 {
		fmt.Println("Database created ...")
	}

	// Determine how many replicas you want to use for table
	replicas := 1

	if os.Getenv("DB_REPLICAS") != "" {
		var atoiErr error
		replicas, atoiErr = strconv.Atoi(os.Getenv("DB_REPLICAS"))

		if atoiErr != nil {
			panic(atoiErr.Error())
		}
	}

	// Check if table exists and create it if not
	tableCursor, tableErr := r.TableList().Contains(table).Do(func(term r.Term) r.Term {
		return r.Branch(term, r.WriteResponse{TablesCreated: 1}, r.TableCreate(table, r.TableCreateOpts{Replicas: replicas}))
	}).Run(session)

	if tableErr != nil {
		panic(tableErr.Error())
	}

	// Get cursor documents
	var tableResult interface{}
	tableCursor.One(&tableResult)

	// Convert documents to map of string keys and interface values
	tableResultMap, _ := tableResult.(map[string]interface{})

	// Get array of interfaces for config_changes key
	tableConfigChange := tableResultMap["config_changes"].([]interface{})

	if len(tableConfigChange) > 0 {
		fmt.Println("Table created ...")
	}

	// Reconfigure table replicating if database and table already exists and replicas number is more than 1
	if replicas > 1 {
		fmt.Println("Reconfiguring number of replicas for table " + table)
		_, reconfigureErr := r.Table(table).Reconfigure(r.ReconfigureOpts{Shards: 1, Replicas: replicas}).Run(session)
		if reconfigureErr != nil {
			panic(reconfigureErr.Error())
		}
	}

	close()

	votes := Read()
	// Create initial data only if table is empty
	if len(votes) == 0 {
		create()
	}
}

func Read() []models.Vote {
	connect()

	votes := []models.Vote{}

	fmt.Println("Fetching data ...")
	res, fetchErr := r.Table(table).Run(session)

	if fetchErr != nil {
		panic(fetchErr.Error())
	}

	mapErr := res.All(&votes)

	if mapErr != nil {
		panic(mapErr.Error())
	}

	fmt.Printf("Data fetched: %v \n", votes)

	close()

	return votes
}

func Update(newVote models.Vote) models.Vote {
	connect()
	fmt.Println("Updating data ...")

	updateCursor, updateErr := r.Table(table).Filter(r.Row.Field("name").Eq(newVote.Name)).Update(newVote, r.UpdateOpts{ReturnChanges: true}).RunWrite(session)

	close()

	if updateErr != nil {
		panic(updateErr.Error())
	}

	// Cursor with changes
	updateChanges := updateCursor.Changes
	// New value from update
	newValue := updateChanges[0].NewValue.(map[string]interface{})
	count := newValue["count"].(float64)
	name := newValue["name"].(string)
	vote := models.Vote{Name: name, Count: int(count)}

	fmt.Println("Data updated", vote)

	return vote
}
