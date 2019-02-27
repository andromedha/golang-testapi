package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/andromedha/golang-testapi/dataclasses"
	"github.com/andromedha/golang-testapi/repositorys"

	"github.com/gorilla/mux"
)

func main() {
	repo := repositorys.NewMongoRepository()
	defer repositorys.CloseConnection(&repo)
	r := mux.NewRouter()
	fmt.Printf("Server started on port 10000...\n")
	r.HandleFunc("/mongo/databases", GetDataBasesHandler(&repo)).Methods("GET")
	r.HandleFunc("/mongo/collections/{database}", GetCollectionHandler(&repo)).Methods("GET")
	r.HandleFunc("/mongo/connection", SetDatabaseData(&repo)).Methods("POST")
	r.HandleFunc("/mongo/create", CreateDocument(&repo)).Methods("POST")
	r.HandleFunc("/mongo/find/{id}", FindDocument(&repo)).Methods("GET")
	r.HandleFunc("/mongo/update", UpdateDocument(&repo)).Methods("POST")
	r.HandleFunc("/mongo/delete/{id}", DeleteDocument(&repo)).Methods("GET")
	http.ListenAndServe(":10000", r)
}

//GetDataBasesHandler show all Databases
func GetDataBasesHandler(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := repo.GetDataBaseList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json, _ := json.Marshal(result)
		w.Write(json)
	}
}

//GetCollectionHandler show all Collections
func GetCollectionHandler(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		result, errresult := repo.GetCollection(vars["database"])
		if errresult != nil {
			http.Error(w, errresult.Error(), http.StatusInternalServerError)
			return
		}
		json, _ := json.Marshal(result)
		w.Write(json)
	}
}

//SetDatabaseData save database and collection name for future request
func SetDatabaseData(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var connection dataclasses.Connection
		errdecode := json.NewDecoder(r.Body).Decode(&connection)
		if errdecode != nil {
			http.Error(w, errdecode.Error(), http.StatusInternalServerError)
			return
		}
		repo.SetDatabase(connection)
		w.Write([]byte("Sucessfully set the connection data"))
	}
}

//CreateDocument create a new document on the database
func CreateDocument(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var file dataclasses.Textfile
		errdecode := json.NewDecoder(r.Body).Decode(&file)
		if errdecode != nil {
			http.Error(w, errdecode.Error(), http.StatusInternalServerError)
			return
		}
		result, errresult := repo.CreateFile(file)
		if errresult != nil {
			http.Error(w, errresult.Error(), http.StatusInternalServerError)
			return
		}
		json, _ := json.Marshal(result)
		w.Write(json)
	}
}

//FindDocument find and return one document from the database
func FindDocument(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, errconv := strconv.Atoi(vars["id"])
		if errconv != nil {
			http.Error(w, errconv.Error(), http.StatusInternalServerError)
			return
		}
		result, errresult := repo.GetFile(id)
		if errresult != nil {
			http.Error(w, errresult.Error(), http.StatusInternalServerError)
			return
		}
		json, _ := json.Marshal(result)
		w.Write(json)
	}
}

//UpdateDocument update one given document on the database
func UpdateDocument(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var file dataclasses.Textfile
		errdecode := json.NewDecoder(r.Body).Decode(&file)
		if errdecode != nil {
			http.Error(w, errdecode.Error(), http.StatusInternalServerError)
			return
		}
		result, errresult := repo.UpdateFile(file)
		if errresult != nil {
			http.Error(w, errresult.Error(), http.StatusInternalServerError)
			return
		}
		json, _ := json.Marshal(result)
		w.Write(json)
	}
}

//DeleteDocument remove one document from the database
func DeleteDocument(repo repositorys.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, errdecode := strconv.Atoi(vars["id"])
		if errdecode != nil {
			http.Error(w, errdecode.Error(), http.StatusInternalServerError)
			return
		}
		result, errresult := repo.DeleteFile(id)
		if errresult != nil {
			http.Error(w, errresult.Error(), http.StatusInternalServerError)
			return
		}
		if result {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write([]byte("No matching document found and deleted"))
		return

	}
}
