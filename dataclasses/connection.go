package dataclasses

//Connection represent the name of the database and collection for the mongoDB
type Connection struct {
	Database   string
	Collection string
}
