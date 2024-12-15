package main

import (
	"strings"
	
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type pageOpt struct {
	contents []Node
	flashes  []Node
}

type OptFn func(*pageOpt)

func Index(contents ...Node) Node {
	return IndexWith(WithSection(contents...))
}

func WithSection(n ...Node) OptFn {
	return func(opt *pageOpt) {
		opt.contents = append(opt.contents, n...)
	}
}

func WithContactFlashes(flashes ...contactFlash) OptFn {
	return func(opt *pageOpt) {
		opt.flashes = append(opt.flashes, Map(flashes, flashContact))
	}
}

func IndexWith(opts ...OptFn) Node {
	var opt pageOpt
	for _, o := range opts {
		o(&opt)
	}
	
	body := []Node{
		hx.Boost("true"),
		Main(
			Header(
				H1(
					allCaps("contacts.app"),
					subTitle("A Demo Contacts Application"),
				),
			),
			Group(opt.flashes),
			Group(opt.contents),
		),
	}
	return HTML5(HTML5Props{
		Title:       "Contacts App",
		Description: "the contacts app",
		Language:    "en",
		Head: []Node{
			stylesheet("/static/css/missing-1.1.1.min.css"),
			stylesheet("/static/css/site.css"),
			Script(Src("/static/js/htmx-1.9.6.min.js")),
		},
		Body: body,
	})
}

func contactsSection(q string, contacts ...Contact) Node {
	return Group{
		contactSearchForm(q),
		Br(),
		contactsTable(q, contacts...),
		P(A(Href("/contacts/new"), Text("Add Contact"))),
	}
}

func contactSearchForm(q string) Node {
	return Form(
		Action("/contacts"), Method("get"), Class("tool-bar"),
		Label(For("search"), Text("Search Term")),
		Input(ID("search"), Type("search"), Name("q"), Value(q)),
		Input(Type("submit"), Value("Search")),
	)
}

func contactsTable(q string, contacts ...Contact) Node {
	return Table(
		THead(
			Tr(th("First"), th("Last"), th("Phone"), th("email")),
		),
		TBody(Map(contacts, func(c Contact) Node {
			hide := q != "" && !(
				strings.Contains(c.First, q) ||
					strings.Contains(c.Last, q) ||
					strings.Contains(c.Phone, q) ||
					strings.Contains(c.Email, q))
			return If(!hide, Tr(
				td(c.First),
				td(c.Last),
				td(c.Phone),
				td(c.Email),
				Td(
					A(Href("/contacts/"+c.ID+"/edit"), Text("Edit")),
					A(Href("/contacts/"+c.ID), Text("View")),
				),
			))
		})),
	)
}

func addContactSection(contact Contact, action string, errors map[string]string) Node {
	nodes := Group{
		Form(
			Action(action), Method("post"),
			FieldSet(
				Legend(Text("Contact Values")),
				P(
					Label(For("email"), Text("Email")),
					Input(Name("email"), ID("email"), Type("email"), Placeholder("dodgers@stink.com"), Value(contact.Email)),
					Span(Class("error"), Text(errors["email"])),
				),
				P(
					Label(For("first_name"), Text("First Name")),
					Input(Name("first_name"), ID("first_name"), Type("text"), Placeholder("First Name"), Value(contact.First)),
					Span(Class("error"), Text(errors["first_name"])),
				),
				P(
					Label(For("last_name"), Text("Last Name")),
					Input(Name("last_name"), ID("last_name"), Type("text"), Placeholder("Last Name"), Value(contact.Last)),
					Span(Class("error"), Text(errors["last_name"])),
				),
				P(
					Label(For("phone"), Text("Phone")),
					Input(Name("phone"), ID("phone"), Type("tel"), Placeholder("XXX-XXX-XXXX"), Value(contact.Phone)),
					Span(Class("error"), Text(errors["phone"])),
				),
				Button(Text("Save")),
			),
		),
	}
	
	if newAction, ok := strings.CutSuffix(action, "/edit"); ok {
		nodes = append(nodes, Form(
			Action(newAction+"/delete"), Method("post"),
			Button(Text("Delete contact")),
		))
	}
	
	nodes = append(nodes, P(A(Href("/contacts"), Text("Back"))))
	
	return nodes
}

func showContactSection(c Contact) Node {
	return Group{
		H1(Text(c.First + " " + c.Last)),
		Div(
			Div(Text("Phone: "+c.Phone)),
			Div(Text("Email: "+c.Email)),
		),
		P(
			A(Href("/contacts/"+c.ID+"/edit"), Text("Edit")),
			A(Href("/contacts"), Text("Back")),
		),
	}
}

func stylesheet(ref string) Node {
	return Link(Rel("stylesheet"), Href(ref))
}

func allCaps(text string) Node {
	return Span(Class("all-caps"), Text(text))
}

func subTitle(text string) Node {
	return El("sub-title", Text(text))
}

func flashContact(c contactFlash) Node {
	return Div(Class("flash"), Text(c.firstLast+" "+c.op+" "), A(Href("/contacts/"+c.id), Text("view")))
}

func th(text string) Node {
	return Th(Text(text))
}

func td(text string) Node {
	return Td(Text(text))
}
