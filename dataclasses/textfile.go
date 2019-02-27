package dataclasses

//Textfile represnent a document on the database
type Textfile struct {
	ID    int32 `bson:"_id"`
	Title string
	Text  string
}
