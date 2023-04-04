package main

import (
	"BookingProject/pkg/config"
	"BookingProject/pkg/driver"
	"BookingProject/pkg/handlers"
	"BookingProject/pkg/helpers"
	"BookingProject/pkg/models"
	"BookingProject/pkg/render"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
)

const portNum = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {

	//what am i going to put in session
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)

	fmt.Println("Starting Mail Listener")
	listenForMail()

	srv := &http.Server{
		Addr:    portNum,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()

	log.Fatal(err)

	_ = http.ListenAndServe(portNum, nil)
}

func run() (*driver.DB, error) {

	//what i am going to put in session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProd = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProd

	app.Session = session

	//connect to database
	log.Println("Connecting to Database")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=abhi2811sharma$$$")
	if err != nil {
		log.Fatal("Cannot connect to DB")
	}

	log.Println("Connected to Database")

	tmplcache, err := render.CreateTemplateCache()

	if err != nil {
		log.Println("Cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tmplcache
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)
	return db, nil
}
