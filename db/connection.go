package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var (
	dbConn *sql.DB
	err    error
)

func init() {
	// dbConn, err = sql.Open("postgres", "user=danny password=danny dbname=gopostgres sslmode=disable") // local
	dbConn, err = sql.Open("postgres", "host=dpg-cp5h5df79t8c73eu79ag-a.oregon-postgres.render.com port=5432 user=danny password=K7GFIvEnOmJ68MHckRoHZMe6Yub9YsCT dbname=gopostgres") // render

	if err != nil {
		panic(err.Error())
	}

}

/*
Terminal to Render : PGPASSWORD=K7GFIvEnOmJ68MHckRoHZMe6Yub9YsCT psql -h dpg-cp5h5df79t8c73eu79ag-a.oregon-postgres.render.com -U danny gopostgres


vscode :
Hostname: dpg-cp5h5df79t8c73eu79ag-a.oregon-postgres.render.com
Port: 5432
Username: danny
Password: K7GFIvEnOmJ68MHckRoHZMe6Yub9YsCT

ssl 要打開

*/
