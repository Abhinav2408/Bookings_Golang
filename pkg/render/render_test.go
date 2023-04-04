package render

import (
	"BookingProject/pkg/models"
	"log"
	"net/http"
	"testing"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Flash value of 123 not found in session")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil

}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		log.Println("line 45")
		t.Error(err)
	}

	app.TemplateCache = tc
	r, err := getSession()
	if err != nil {
		log.Println("line 52")
		t.Error(err)
	}

	var ww myWriter

	err = Template(&ww, r, "home.page.html", &models.TemplateData{})

	if err != nil {
		log.Println("line 61")
		t.Error("Error writing template to browser")
	}

	err = Template(&ww, r, "non-existent.page.html", &models.TemplateData{})
	log.Println(err)
	if err == nil {
		t.Error("Rendered non existent template from cache")
	}

}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)

}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
