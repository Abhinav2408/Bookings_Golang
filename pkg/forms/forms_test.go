package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	req := httptest.NewRequest("POST", "/whatever", nil)
	form := New(req.PostForm)

	isvalid := form.Valid()
	if !isvalid {
		t.Error("got invalid, should be valid")
	}
}

func TestForm_Required(t *testing.T) {
	req := httptest.NewRequest("POST", "/whatever", nil)
	form := New(req.PostForm)
	form.Required("a", "b", "c")
	isvalid := form.Valid()
	if isvalid {
		t.Error("got valid, should be invalid")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ := http.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	isvalid = form.Valid()
	if !isvalid {
		t.Error("got invalid, should be valid")
	}
}

func TestForm_Has(t *testing.T) {
	req := httptest.NewRequest("POST", "/whatever", nil)
	form := New(req.PostForm)

	has := form.Has("whatever", req)
	if has {
		t.Error("no field present, still gave present")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")
	req.Form = postedData
	form = New(postedData)

	has = form.Has("a", req)
	if !has {
		t.Error("field present, still gave not present")
	}
}

func TestForm_MinLength(t *testing.T) {
	req := httptest.NewRequest("POST", "/whatever", nil)

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")
	req.PostForm = postedData
	form := New(req.PostForm)

	val := form.MinLength("a", 3, req)
	if len(form.Errors.Get("a")) == 0 {
		t.Error("Should have errors")
	}

	if val {
		t.Error("Wrong values")
	}

	form = New(req.PostForm)
	val = form.MinLength("a", 0, req)

	if len(form.Errors.Get("a")) != 0 {
		t.Error("Should not have errors")
	}

	if !val {
		t.Error("Wrong values")
	}

}

func TestForm_IsEmail(t *testing.T) {
	req := httptest.NewRequest("POST", "/whatever", nil)
	form := New(req.PostForm)

	form.IsEmail("X")
	if form.Valid() {
		t.Error("form shows valid email for non existent field")
	}

	postedValues := url.Values{}
	postedValues.Add("email", "me@here.com")

	form = New(postedValues)

	form.IsEmail("email")
	if !form.Valid() {
		t.Error("form shows invalid email for valid email")
	}

	postedValues = url.Values{}
	postedValues.Add("email", "x")

	form = New(postedValues)

	form.IsEmail("email")
	if form.Valid() {
		t.Error("form shows valid email for invalid email")
	}

}
