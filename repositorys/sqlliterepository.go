package repositorys

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/andromedha/golang-testapi/dataclasses"

	//needed because we open the sqllite db with this driver but never reference/use the driver explicit
	_ "github.com/mattn/go-sqlite3"
)

//SQLLiteRepository inherited from Repository
type SQLLiteRepository struct {
	client     *sql.DB
	connected  bool
	connection dataclasses.Connection
}

//NewSQLLiteRepository create new repository
func NewSQLLiteRepository() SQLLiteRepository {
	sqlLiteRepository, err := ConnectSQLLite()
	if err != nil {
		log.Fatalf("Can not connect to sqllite db! %v", err)
	}
	return sqlLiteRepository
}

//ConnectSQLLite handles connect to database
func ConnectSQLLite() (SQLLiteRepository, error) {
	repo := SQLLiteRepository{}
	dbpath := "./testdb.db"
	os.Remove(dbpath)
	db, err := sql.Open("sqlite3", dbpath)
	repo.client = db
	tables, collErr := repo.GetCollection("")
	if collErr != nil {
		return repo, err
	}
	if len(tables) < 1 || index(tables, "textfiles") < 0 {
		initalTableStatement := "create table textfiles (ID integer not null primary key, Title text,Name text)"
		_, err = db.Exec(initalTableStatement)
	}
	return repo, err
}

//CloseConnection shut down the connection
func (repo *SQLLiteRepository) CloseConnection() error {
	err := repo.client.Close()
	return err
}

//GetDataBaseList returns a List of Databases
func (repo *SQLLiteRepository) GetDataBaseList() ([]string, error) {
	databases := make([]string, 0)
	return databases, errors.New("By SqlLite only single database is supported")
}

//GetCollection return all collections of a Database
func (repo *SQLLiteRepository) GetCollection(database string) ([]string, error) {
	dbStatement := "SELECT name FROM sqlite_master WHERE type ='table' AND name NOT LIKE 'sqlite_%';"
	rows, err := repo.client.Query(dbStatement)
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var name string
		scanErr := rows.Scan(&name)
		if scanErr != nil {
			log.Fatal(scanErr)
		}
		result = append(result, name)
	}
	rowErr := rows.Err()
	if rowErr != nil {
		return result, rowErr
	}
	return result, err
}

//SetDatabase set the connections values for operation on the mongodb
func (repo *SQLLiteRepository) SetDatabase(databasedata dataclasses.Connection) error {
	return errors.New("Not supported by SqlLite")
}

//CreateFile create a new document on the database
func (repo *SQLLiteRepository) CreateFile(file dataclasses.Textfile) (int, error) {
	context, ctxErr := repo.client.Begin()
	if ctxErr != nil {
		return 0, ctxErr
	}
	statement, stmErr := context.Prepare("insert into textfiles(title,name) values (?,?)")
	if stmErr != nil {
		return 0, stmErr
	}
	defer statement.Close()
	result, resultErr := statement.Exec(file.Title, file.Text)
	if resultErr != nil {
		return 0, resultErr
	}
	commitErr := context.Commit()
	if commitErr != nil {
		return 0, commitErr
	}
	rowid, _ := result.LastInsertId()
	return int(rowid), nil
}

//GetFile return a textfile from the database specified by the documentid
func (repo *SQLLiteRepository) GetFile(id int) (dataclasses.Textfile, error) {
	var result dataclasses.Textfile
	dbstatement := "Select ID,Title,Name from textfiles where ID = ?"
	rows, rowsErr := repo.client.Query(dbstatement, id)
	if rowsErr != nil {
		return result, rowsErr
	}
	for rows.Next() {
		var id int32
		var title, name string
		scanErr := rows.Scan(&id, &name, &title)
		if scanErr != nil {
			log.Fatal(scanErr)
		}
		result.ID = id
		result.Text = name
		result.Title = title
	}
	if result.ID == 0 {
		return result, fmt.Errorf("No document with id %v", id)
	}
	return result, nil
}

//UpdateFile update a recent file on the database
func (repo *SQLLiteRepository) UpdateFile(file dataclasses.Textfile) (dataclasses.Textfile, error) {
	var result dataclasses.Textfile
	context, ctxErr := repo.client.Begin()
	if ctxErr != nil {
		return result, ctxErr
	}
	statement, stmErr := context.Prepare("Update textfiles set Title = ?, Name = ? where ID = ?")
	if stmErr != nil {
		return result, stmErr
	}
	defer statement.Close()
	dbResult, resultErr := statement.Exec(file.Title, file.Text, file.ID)
	if resultErr != nil {
		return result, resultErr
	}
	commitErr := context.Commit()
	if commitErr != nil {
		return result, commitErr
	}
	if rows, _ := dbResult.RowsAffected(); rows > 0 {
		return file, nil
	}
	return result, errors.New("No affected rows")
}

//DeleteFile remove a textfile from the database
func (repo *SQLLiteRepository) DeleteFile(id int) (bool, error) {
	context, ctxErr := repo.client.Begin()
	if ctxErr != nil {
		return false, ctxErr
	}
	statement, stmErr := context.Prepare("Delete from textfiles where ID = ?")
	if stmErr != nil {
		return false, stmErr
	}
	defer statement.Close()
	dbResult, resultErr := statement.Exec(id)
	if resultErr != nil {
		return false, resultErr
	}
	commitErr := context.Commit()
	if commitErr != nil {
		return false, commitErr
	}
	if rows, _ := dbResult.RowsAffected(); rows > 0 {
		return true, nil
	}
	return false, errors.New("No affected rows")

}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}
