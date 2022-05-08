package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"go.elastic.co/apm/module/apmlogrus"
	"go.elastic.co/apm/module/apmsql"
	_ "go.elastic.co/apm/module/apmsql/sqlite3"
)

var db *sql.DB

func helloHandler(w http.ResponseWriter, req *http.Request) {
	name := mux.Vars(req)["name"]
	log := log.WithFields(apmlogrus.TraceContext(req.Context()))
	log.WithField("name", name).Info("handling hello request")

	cnt, err := updateRequestCount(req.Context(), name, log)
	if err != nil {
		log.WithError(err).Error("failed to update request count")
		http.Error(w, "failed to update request count", 500)
		return
	}
	fmt.Fprintf(w, "Hello, %s!. You visited %d times\n", name, cnt)
}

func updateRequestCount(ctx context.Context, name string, log *logrus.Entry) (int, error) {
	span, ctx := apm.StartSpan(ctx, "updateRequestCount", "custom")
	defer span.End()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	row := tx.QueryRowContext(ctx, "SELECT count FROM stats WHERE name=:name", sql.Named("name", name))
	var count int
	switch err := row.Scan(&count); err {
	case nil:
		count++
		if _, err := tx.ExecContext(ctx, "UPDATE stats SET count=:count WHERE name=:name", sql.Named("count", count), sql.Named("name", name)); err != nil {
			return -1, err
		}
		log.WithField("name", name).Infof("updated count to %d", count)
	case sql.ErrNoRows:
		count = 1
		if _, err := tx.ExecContext(ctx, "INSERT INTO stats (name, count) VALUES (:name, :count)", sql.Named("name", name), sql.Named("count", count)); err != nil {
			return -1, err
		}
		log.WithField("name", name).Info("count initialized to 1")
	default:
		return -1, err
	}
	return count, tx.Commit()
}

func main() {
	var err error

	db, err = apmsql.Open("sqlite3", ":memory:")
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}
	if _, err = db.Exec("CREATE TABLE stats (name TEXT PRIMARY KEY, count INTEGER);"); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/hello/{name}", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", apmhttp.Wrap(r)))
}
