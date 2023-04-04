package handlers

import (
	"BookingProject/pkg/config"
	"BookingProject/pkg/driver"
	"BookingProject/pkg/forms"
	"BookingProject/pkg/helpers"
	"BookingProject/pkg/models"
	"BookingProject/pkg/render"
	"BookingProject/pkg/repository"
	"BookingProject/pkg/repository/dbrepo"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

// Repository used by handlers
var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// creates new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// sets repository for handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, req *http.Request) {

	render.Template(w, req, "home.page.html", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, req *http.Request) {

	//perform some logic

	render.Template(w, req, "about.page.html", &models.TemplateData{})
}

func (m *Repository) Generals(w http.ResponseWriter, req *http.Request) {

	render.Template(w, req, "generals.page.html", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, req *http.Request) {

	render.Template(w, req, "majors.page.html", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, req *http.Request) {

	render.Template(w, req, "search-availability.page.html", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, req *http.Request) {

	start := req.Form.Get("start")
	end := req.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(req.Context(), "error", "No availability")
		http.Redirect(w, req, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(req.Context(), "reservation", res)

	render.Template(w, req, "reservation-summary.html", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:room_id`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// handles request for avail in general and major, and sends JSON response
func (m *Repository) JSONAvailability(w http.ResponseWriter, req *http.Request) {

	sd := req.Form.Get("start")
	ed := req.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(req.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	resp := jsonResponse{
		OK:        available,
		Message:   "Available",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, req *http.Request) {

	render.Template(w, req, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, req *http.Request) {

	res, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(req.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)

	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't find room")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(req.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)

	stringMap["start-date"] = sd
	stringMap["end-date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, req, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, req *http.Request) {

	err := req.ParseForm()
	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't parse form")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := req.Form.Get("start_date")
	ed := req.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)

	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't parse start date")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)

	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't parse end date")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(req.Form.Get("room_id"))

	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Invalid room id")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Invalid room id")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: req.Form.Get("first_name"),
		LastName:  req.Form.Get("last_name"),
		Phone:     req.Form.Get("phone"),
		Email:     req.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room:      room,
	}

	form := forms.New(req.PostForm)

	//form.Has("first_name", req)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, req)
	form.IsEmail("email")
	//if form is invalid, we don't want to lost that data
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "Invalid Data in form", http.StatusSeeOther)
		render.Template(w, req, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't add reservation")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(req.Context(), "reservation", reservation)

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(req.Context(), "error", "Can't insert restriction")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	//send notifications

	htmlMessage := fmt.Sprintf(`<strong>Reservation Confirmation</strong><br>
	Dear %s:,<br>
	This is to confirm your reservation from %s to %s`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:      reservation.Email,
		From:    "me@here.com",
		Subject: "Room Confirmation",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(req.Context(), "reservation", reservation)

	http.Redirect(w, req, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, req *http.Request) {
	reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get reservation from session")
		m.App.Session.Put(req.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(req.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	render.Template(w, req, "reservation-summary.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, req *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(req, "id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(req.Context(), "reservation", res)
	http.Redirect(w, req, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) BookRoom(w http.ResponseWriter, req *http.Request) {

	//id,s,e
	roomID, _ := strconv.Atoi(req.URL.Query().Get("id"))
	sd := req.URL.Query().Get("s")
	ed := req.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(req.Context(), "reservation", res)

	http.Redirect(w, req, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) ShowLogin(w http.ResponseWriter, req *http.Request) {
	render.Template(w, req, "login.page.html", &models.TemplateData{Form: forms.New(nil)})
}

func (m *Repository) PostShowLogin(w http.ResponseWriter, req *http.Request) {
	_ = m.App.Session.RenewToken(req.Context())
	//this prevents a session fixation attack. should be done on login/logout

	err := req.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := req.Form.Get("email")
	password := req.Form.Get("password")

	form := forms.New(req.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		//take user back to page
		render.Template(w, req, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)

	if err != nil {
		log.Println(err)
		m.App.Session.Put(req.Context(), "error", "invalid login credentials")
		http.Redirect(w, req, "/user/login", http.StatusSeeOther)
	}

	//user must now be logged in

	m.App.Session.Put(req.Context(), "user_id", id)
	m.App.Session.Put(req.Context(), "flash", "Logged in Successfully")
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, req *http.Request) {
	_ = m.App.Session.Destroy(req.Context())
	_ = m.App.Session.RenewToken(req.Context())

	http.Redirect(w, req, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, req *http.Request) {
	render.Template(w, req, "admin-dashboard.page.html", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, req *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, req, "admin-new-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, req *http.Request) {

	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, req, "admin-all-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, req *http.Request) {

	exploded := strings.Split(req.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)

	stringMap["src"] = src

	year := req.URL.Query().Get("y")
	month := req.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	//get data from DB

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, req, "admin-reservations-show.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, req *http.Request) {

	err := req.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(req.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)

	stringMap["src"] = src

	//get data from DB

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = req.Form.Get("first_name")
	res.LastName = req.Form.Get("last_name")
	res.Email = req.Form.Get("email")
	res.Phone = req.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(req.Context(), "flash", "Changes saved")

	month := req.Form.Get("month")
	year := req.Form.Get("year")

	if year == "" {
		http.Redirect(w, req, fmt.Sprintf("admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, req *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(req, "id"))
	src := chi.URLParam(req, "src")

	_ = m.DB.UpdateProcessedForReservation(id, 1)

	year := req.URL.Query().Get("y")
	month := req.URL.Query().Get("m")

	m.App.Session.Put(req.Context(), "flash", "Reservation Marked as processed")

	if year == "" {
		http.Redirect(w, req, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, req *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(req, "id"))
	src := chi.URLParam(req, "src")

	_ = m.DB.DeleteReservation(id)

	year := req.URL.Query().Get("y")
	month := req.URL.Query().Get("m")

	m.App.Session.Put(req.Context(), "flash", "Reservation Deleted")

	if year == "" {
		http.Redirect(w, req, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, req *http.Request) {

	now := time.Now()

	if req.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(req.URL.Query().Get("y"))
		month, _ := strconv.Atoi(req.URL.Query().Get("m"))

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)  //next month
	last := now.AddDate(0, -1, 0) // last month

	nextMonth := next.Format("01") //gives month
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)

	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	//get the first and last days of month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)

	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {

		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		//get all restrictions for the current room
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				//its reservation
				for d := y.StartDate; !d.After(y.EndDate); d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}

			} else {
				//its a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(req.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, req, "admin-reservations-calendar.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(req.Form.Get("y"))
	month, _ := strconv.Atoi(req.Form.Get("m"))

	//process

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(req.PostForm)

	for _, x := range rooms {
		//get the blockmap from session. loop through it. if we have an entry in map that is not in our posted data and restriction id >0, then we remove it
		curMap := m.App.Session.Get(req.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)

		for name, value := range curMap {
			//ok will be false if value is not in map
			if val, ok := curMap[name]; ok {

				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name), req) {
						//delete the restriction by id
						err := m.DB.DeleteBlockByID(value)

						if err != nil {
							log.Println(err)
						}

					}
				}
			}
		}
	}

	//now handle new blocks
	for name := range req.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			//insert new block

			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}

		}
	}

	m.App.Session.Put(req.Context(), "flash", "Changes Saved")
	http.Redirect(w, req, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)

}
