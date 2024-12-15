package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	
	. "maragu.dev/gomponents"
	ghttp "maragu.dev/gomponents/http"
)

//go:embed static
var assets embed.FS

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	
	contacts := Contacts{
		v: []Contact{
			{
				ID:    "1",
				First: "John",
				Last:  "Saddle",
				Phone: "555-251-1111",
				Email: "dodgers@stink.net",
			},
			{
				ID:    "2",
				First: "Omar",
				Last:  "Ibn al Khattab",
				Phone: "555-123-4567",
				Email: "oneof@greats.com",
			},
			{
				ID:    "3",
				First: "Khadijah",
				Last:  "Bint Khuwaylid",
				Phone: "555-765-4321",
				Email: "salaams@unlimited.com",
			},
		},
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/contacts", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /contacts", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		section := contactsSection(r.URL.Query().Get("q"), contacts.v...)
		contactFlashes := getContactFlashes(w, r)
		return IndexWith(WithSection(section), WithContactFlashes(contactFlashes...)), nil
	}))
	mux.HandleFunc("GET /contacts/new", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		return Index(addContactSection(Contact{}, r.URL.Path, nil)), nil
	}))
	mux.HandleFunc("POST /contacts/new", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		c := Contact{
			ID:    rand.Text(),
			Email: r.FormValue("email"),
			First: r.FormValue("first_name"),
			Last:  r.FormValue("last_name"),
			Phone: r.FormValue("phone"),
		}
		if errs := contactValid(c); len(errs) > 0 {
			return Index(addContactSection(c, r.URL.Path, errs)), nil
		}
		if !contacts.Add(c) {
			errs := map[string]string{
				"email": "email exists for: " + c.Email,
			}
			return Index(addContactSection(c, r.URL.Path, errs)), nil
		}
		
		setContactFlash(w, c, "created")
		
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		
		return nil, nil
	}))
	mux.HandleFunc("GET /contacts/{contact_id}", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		c, ok := contacts.Find(r.PathValue("contact_id"))
		if !ok {
			http.Redirect(w, r, "/contacts", http.StatusSeeOther)
			return nil, nil
		}
		return Index(showContactSection(c)), nil
	}))
	mux.HandleFunc("GET /contacts/{contact_id}/edit", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		cid := r.PathValue("contact_id")
		c, ok := contacts.Find(cid)
		if !ok {
			http.Redirect(w, r, "/contacts", http.StatusSeeOther)
			return nil, nil
		}
		return Index(addContactSection(c, r.URL.Path, nil)), nil
	}))
	mux.HandleFunc("POST /contacts/{contact_id}/edit", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		c := Contact{
			ID:    r.PathValue("contact_id"),
			Email: r.FormValue("email"),
			First: r.FormValue("first_name"),
			Last:  r.FormValue("last_name"),
			Phone: r.FormValue("phone"),
		}
		if _, ok := contacts.Find(c.ID); !ok {
			http.Redirect(w, r, "/contacts", http.StatusSeeOther)
			return nil, nil
		}
		if errs := contactValid(c); len(errs) > 0 {
			return Index(addContactSection(c, r.URL.Path, errs)), nil
		}
		if !contacts.Update(c) {
			setContactFlash(w, c, "failed to find")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return nil, nil
		}
		
		setContactFlash(w, c, "updated")
		
		http.Redirect(w, r, "/contacts/"+c.ID, http.StatusSeeOther)
		
		return nil, nil
	}))
	mux.HandleFunc("POST /contacts/{contact_id}/delete", ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (Node, error) {
		contacts.Rm(r.PathValue("contact_id"))
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return nil, nil
	}))
	mux.HandleFunc("GET /static/{folder}/", func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("folder") {
		case "js":
			w.Header().Set("Content-Type", "text/javascript")
		case "css":
			w.Header().Set("Content-Type", "text/css")
		}
		http.FileServerFS(assets).ServeHTTP(w, r)
	})
	
	port := ":8080"
	logger.Info("listening on http://localhost" + port)
	http.ListenAndServe(port, mux)
}

type Contacts struct {
	mu sync.Mutex
	v  []Contact
}

func (c *Contacts) Add(contact Contact) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, v := range c.v {
		if contact.Email == v.Email {
			return false
		}
	}
	
	c.v = append(c.v, contact)
	
	return true
}

func (c *Contacts) Find(id string) (Contact, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, v := range c.v {
		if v.ID == id {
			return v, true
		}
	}
	
	return Contact{}, false
}

func (c *Contacts) Rm(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, v := range c.v {
		if v.ID == id {
			c.v = append(c.v[:i], c.v[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Contacts) Update(contact Contact) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, v := range c.v {
		if v.ID == contact.ID {
			c.v[i] = contact
			return true
		}
	}
	return false
}

type Contact struct {
	ID    string
	First string
	Last  string
	Phone string
	Email string
}

func contactValid(c Contact) map[string]string {
	out := make(map[string]string)
	if c.Email == "" {
		out["email"] = "email is required"
	}
	if c.First == "" {
		out["first_name"] = "first name is required"
	}
	if c.Last == "" {
		out["last_name"] = "last name is required"
	}
	if c.Phone == "" {
		out["phone"] = "phone number is required"
	}
	return out
}

type contactFlash struct {
	id        string
	firstLast string
	op        string
}

func (c contactFlash) EncodeOp(op string) string {
	return fmt.Sprintf("contact:%s:%s:%s", op, c.id, c.firstLast)
}

func setContactFlash(w http.ResponseWriter, c Contact, op string) {
	f := contactFlash{
		id:        c.ID,
		firstLast: c.First + " " + c.Last,
		op:        "created",
	}
	setFlash(w, "contact", f.EncodeOp(op))
}

func getContactFlashes(w http.ResponseWriter, r *http.Request) []contactFlash {
	fm, err := getFlash(w, r, "contact")
	if err != nil {
		return nil
	}
	
	parts := strings.SplitN(fm, ":", 4)
	if len(parts) != 4 {
		return nil
	}
	
	c := contactFlash{
		id:        parts[2],
		firstLast: parts[3],
		op:        parts[1],
	}
	
	return []contactFlash{c}
}
