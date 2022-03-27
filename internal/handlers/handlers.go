package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Pleum-Jednipit/bookings/helpers"
	"github.com/Pleum-Jednipit/bookings/internal/config"
	"github.com/Pleum-Jednipit/bookings/internal/driver"
	"github.com/Pleum-Jednipit/bookings/internal/forms"
	"github.com/Pleum-Jednipit/bookings/internal/models"
	"github.com/Pleum-Jednipit/bookings/internal/render"
	"github.com/Pleum-Jednipit/bookings/internal/repository"
	"github.com/Pleum-Jednipit/bookings/internal/repository/dbrepo"
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
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

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

	m.App.Session.Put(r.Context(),"reservation",reservation)
	http.Redirect(w,r,"/reservation-summary",http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r , "generals.page.tmpl", &models.TemplateData{})
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
	// start := r.Form.Get("start")
	// end := r.Form.Get("end")
	w.Write([]byte("Post Search Routes"))
}


type jsonResponse struct {
	OK bool `json:"ok"`
	Message string `json:"message"`
}
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK: true,
		Message: "Availalable!",
	}
	
	out, err := json.MarshalIndent(resp, "", "   ")
	if err != nil {
		helpers.ServerError(w,err)
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(out)
}

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
	render.Template(w, r , "reservation-summary.page.tmpl" ,&models.TemplateData{})
}

