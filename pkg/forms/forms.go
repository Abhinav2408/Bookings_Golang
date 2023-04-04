package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Form struct {
	url.Values
	Errors errors
}

//initializes a form

func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
		//this is empty error map
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

//does the form have a given field and is not empty

func (f *Form) Has(field string, req *http.Request) bool {
	x := req.Form.Get(field)
	return x != ""
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

//checks minimum length of input

func (f *Form) MinLength(field string, length int, req *http.Request) bool {
	x := req.Form.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be atleast %d characters long", length))
		return false
	}
	return true
}

// checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid Email Address")
	}
}
