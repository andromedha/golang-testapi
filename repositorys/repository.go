package repositorys

import "github.com/andromedha/golang-testapi/dataclasses"

//Repository Abstract Interface
type Repository interface {
	CreateFile(dataclasses.Textfile) (int, error)
	GetDataBaseList() ([]string, error)
	GetCollection(string) ([]string, error)
	SetDatabase(dataclasses.Connection) error
	GetFile(int) (dataclasses.Textfile, error)
	UpdateFile(dataclasses.Textfile) (dataclasses.Textfile, error)
	DeleteFile(int) (bool, error)
}
