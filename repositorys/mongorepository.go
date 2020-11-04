package repositorys

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/andromedha/golang-testapi/dataclasses"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//MongoRepository inherited from Repository
type MongoRepository struct {
	client     mongo.Client
	connected  bool
	connection dataclasses.Connection
}

//NewMongoRepository Create a new Repository for MongoDB
func NewMongoRepository() MongoRepository {
	mongorep, err := Connect()
	if err != nil {
		log.Fatalf("Can not connect to Mongo Server! %v", err)
	}
	return mongorep
}

//Connect to the Mongo DB
func Connect() (MongoRepository, error) {
	repo := MongoRepository{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mongoclient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	defer cancel()
	if err != nil {
		return repo, err
	}
	repo.client = *mongoclient
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		return repo, err
	}
	repo.connected = true
	return repo, nil
}

//CloseConnection shut down the connection
func CloseConnection(repo *MongoRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := repo.client.Disconnect(ctx)
	return err
}

//GetDataBaseList returns a List of Databases
func (repo *MongoRepository) GetDataBaseList() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	databases, err := repo.client.ListDatabaseNames(ctx, nil, &options.ListDatabasesOptions{})
	if err != nil {
		return databases, err
	}
	return databases, nil
}

//GetCollection return all collections of a Database
func (repo *MongoRepository) GetCollection(database string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	collectioncursor, err := repo.client.Database(database).ListCollections(ctx, nil, &options.ListCollectionsOptions{})
	defer collectioncursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	var result []string
	for collectioncursor.Next(ctx) {
		elem := &bson.D{}
		if err := collectioncursor.Decode(&elem); err != nil {
			return nil, err
		}
		elemMap := elem.Map()
		result = append(result, fmt.Sprintf("%v", elemMap["name"]))
	}

	return result, nil
}

//SetDatabase set the connections values for operation on the mongodb
func (repo *MongoRepository) SetDatabase(databasedata dataclasses.Connection) {
	repo.connection = databasedata
}

//CreateFile create a new document on the database
func (repo *MongoRepository) CreateFile(file dataclasses.Textfile) (int, error) {
	data, errdata := bson.Marshal(file)
	if errdata != nil {
		return 0, errdata
	}
	collection, errcol := connectToCollection(repo)
	if errcol != nil {
		return 0, errcol
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	insert, err := collection.InsertOne(ctx, data)
	if err != nil {
		return 0, err
	}
	if id, errid := insert.InsertedID.(int32); errid {
		return int(id), nil
	}

	return 0, errors.New("Fail by creating document on mongodb")
}

//GetFile return a textfile from the database specified by the documentid
func (repo *MongoRepository) GetFile(id int) (dataclasses.Textfile, error) {
	var result dataclasses.Textfile

	collection, errcol := connectToCollection(repo)
	if errcol != nil {
		return result, errcol
	}
	filter := bson.M{"_id": id}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errfile := collection.FindOne(ctx, filter).Decode(&result)
	if errfile != nil {
		return result, errfile
	}
	return result, nil
}

//UpdateFile update a recent file on the database
func (repo *MongoRepository) UpdateFile(file dataclasses.Textfile) (dataclasses.Textfile, error) {

	collection, errcol := connectToCollection(repo)
	if errcol != nil {
		return dataclasses.Textfile{}, errcol
	}

	data, errdata := bson.Marshal(file)
	if errdata != nil {
		return dataclasses.Textfile{}, errdata
	}

	filter := bson.M{"_id": file.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	returndocument := options.ReturnDocument(1)
	var result dataclasses.Textfile
	errupdate := collection.FindOneAndReplace(ctx, filter, data, &options.FindOneAndReplaceOptions{ReturnDocument: &returndocument}).Decode(&result)
	if errupdate != nil {
		return dataclasses.Textfile{}, errupdate
	}
	return result, nil
}

//DeleteFile remove a textfile from the database
func (repo *MongoRepository) DeleteFile(id int) (bool, error) {

	collection, errcol := connectToCollection(repo)
	if errcol != nil {
		return false, errcol
	}

	filter := bson.M{"_id": id}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, errresult := collection.DeleteOne(ctx, filter)
	if errresult != nil {
		return false, errresult
	}
	if result.DeletedCount > 0 {
		return true, nil
	}
	return false, nil
}

func connectToCollection(repo *MongoRepository) (*mongo.Collection, error) {
	if repo.connection.Collection == "" || repo.connection.Database == "" {
		return nil, errors.New("No database and collection is set")
	}
	result := repo.client.Database(repo.connection.Database).Collection(repo.connection.Collection)
	return result, nil
}
