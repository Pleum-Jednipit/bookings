package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Pleum-Jednipit/bookings/helpers"
	"github.com/Pleum-Jednipit/bookings/internal/config"
	"github.com/Pleum-Jednipit/bookings/internal/driver"
	"github.com/Pleum-Jednipit/bookings/internal/forms"
	"github.com/Pleum-Jednipit/bookings/internal/models"
	"github.com/Pleum-Jednipit/bookings/internal/render"
	"github.com/Pleum-Jednipit/bookings/internal/repository"
	"github.com/Pleum-Jednipit/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB: dbrepo.NewPostgreaRepo(db.SQL,a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.Template(w,r , "home.page.tmpl" ,  &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// send data to the template
	render.Template(w,r , "about.page.tmpl" ,  &models.TemplateData{
		StringMap: stringMap,
	})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomById(res.RoomID)
	if err != nil {
		helpers.ServerError(w,err)
	}

	res.Room = room

	m.App.Session.Put(r.Context(),"reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w,errors.New("Can't get from session"))
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")         
	reservation.Phone =  r.Form.Get("phone")
	reservation.Email = r.Form.Get ("email")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(),"reservation",reservation)

	roomRestriction := models.RoomRestriction{
		StartDate: reservation.StartDate,
		EndDate: reservation.EndDate,
		RoomID: reservation.RoomID,
		ReservationID: newReservationId,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(roomRestriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(),"reservation",reservation)
	http.Redirect(w,r,"/reservation-summary",http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r , "generals.page.tmpl", &models.TemplateData{

	})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r , "majors.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w,r , "search-availability.page.tmpl", &models.TemplateData{})
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
    if err != nil {
        log.Fatal(err)
    }
	
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	

	layout := "2006-01-02"

	startDate, err := time.Parse(layout,start)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	endDate, err := time.Parse(layout,end)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}


	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate,endDate)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(),"error","No availability")
		http.Redirect(w,r,"/search-availability",http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate: endDate,
	}

	m.App.Session.Put(r.Context(),"reservation",res)
	
	render.Template(w, r , "choose-room.page.tmpl" ,&models.TemplateData{
		Data: data,
	})
}


type jsonResponse struct {
	OK bool `json:"ok"`
	Message string `json:"message"`
	RoomID string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"

	startDate,_ := time.Parse(layout, sd)
	endDate,_:= time. Parse(layout, ed)

	roomID,_ := strconv.Atoi(r.Form.Get("room_id"))

	fmt.Printf("%s %s %d",startDate,endDate,roomID)

	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	
// 	sd := r.Form.Get("start")
// 	ed := r.Form.Get("end")

// 	fmt.Printf("%s %s",sd,ed)

// 	layout := "2006-01-02"

// 	startDate,_ := time.Parse(layout, sd)
// 	endDate,_ := time.Parse(layout, ed)

// 	roomID,_ := strconv.Atoi(r.Form.Get("room_id"))

// 	fmt.Printf("%s %s %d",startDate,endDate,roomID)

// 	available, _ := m.DB.SearchAvailabilityByDatesByRoomId(startDate,endDate,roomID)
	
// 	resp := jsonResponse{
// 		OK: available,
// 		Message: "",
// 		RoomID: strconv.Itoa(roomID),
// 		StartDate: sd,
// 		EndDate: ed,
		
// 	}

// 	out, err := json.MarshalIndent(resp, "", "   ")
// 	if err != nil {
// 		helpers.ServerError(w,err)
// 	}
// 	w.Header().Set("Content-Type","application/json")
// 	w.Write(out)
// }

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r , "contact.page.tmpl" ,&models.TemplateData{})
}

// Reservation summary
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(r.Context(),"error","Can't get reservation from session")
		http.Redirect(w,r,"/",http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(),"reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate. Format ("2006-01-02")
	ed := reservation.EndDate. Format ("2006-01-02")

	stringMap := make(map[string]string)
	stringMap ["start_date"] = sd
	stringMap ["end_date"] = ed

	render.Template(w, r , "reservation-summary.page.tmpl" ,&models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}


// ChooseRoom display information of rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r,"id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomId
	m.App.Session.Put(r.Context(), "reservation",res)


	http.Redirect(w,r,"/make-reservation",http.StatusSeeOther)

}

// BookRoom takes URL parameters, build a sessional variable and take user to make reservation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomId, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	var res models.Reservation

	layout := "2006-01-02"

	startDate, _ := time.Parse(layout,sd)
	endDate, _ := time.Parse(layout,ed)
	
	room, err := m.DB.GetRoomById(res.RoomID)
	if err != nil {
		helpers.ServerError(w,err)
	}

	res.Room = room
	res.RoomID = roomId
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(),"reservation",res)

	http.Redirect(w,r,"/make-reservation",http.StatusSeeOther)
}