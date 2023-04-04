package config

import (
	"BookingProject/pkg/models"
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

//holds the application config

type AppConfig struct {
	TemplateCache map[string]*template.Template
	UseCache      bool
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProd        bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}
