package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

type Feedback struct {
	// XMLName         xml.Name        `xml:"feedback"`
	ReportMetadata  ReportMetadata  `xml:"report_metadata"`
	PolicyPublished PolicyPublished `xml:"policy_published"`
	Records         []Record        `xml:"record"`
}

type ReportMetadata struct {
	OrgName          string    `xml:"org_name"`
	Email            string    `xml:"email"`
	ExtraContactInfo string    `xml:"extra_contact_info"`
	ReportID         string    `xml:"report_id"`
	DateRange        DateRange `xml:"date_range"`
}

type DateRange struct {
	Begin int `xml:"begin"`
	End   int `xml:"end"`
}

func (d DateRange) FormattedDate() string {
	return fmt.Sprintf("01 Jan, 2006")
}

type PolicyPublished struct {
	XMLName xml.Name `xml:"policy_published"`
	Domain  string   `xml:"domain"`
	Adkim   string   `xml:"adkim"`
	Aspf    string   `xml:"aspf"`
	P       string   `xml:"p"`
	Sp      string   `xml:"sp"`
	Pct     int      `xml:"pct"`
}

type Record struct {
	Row         Row         `xml:"row"`
	Identifiers Identifiers `xml:"identifiers"`
	AuthResults AuthResults `xml:"auth_results"`
}

type Row struct {
	SourceIP        string          `xml:"source_ip"`
	Count           int             `xml:"count"`
	PolicyEvaluated PolicyEvaluated `xml:"policy_evaluated"`
}
type PolicyEvaluated struct {
	Disposition string `xml:"disposition"`
	Dkim        string `xml:"dkim"`
	Spf         string `xml:"spf"`
	Reason      Reason `xml:"reason"`
}
type Reason struct {
	Type    string `xml:"type"`
	Comment string `xml:"comment"`
}
type AuthResults struct {
	Spf   Spf    `xml:"spf"`
	Dkims []Dkim `xml:"dkim"`
}
type Spf struct {
	Domain string `xml:"domain"`
	Result string `xml:"result"`
}
type Dkim struct {
	Domain   string `xml:"domain"`
	Result   string `xml:"result"`
	Selector string `xml:"selector"`
}
type Identifiers struct {
	HeaderFrom string `xml:"header_from"`
}

func main() {
	var feedback Feedback
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(feedback *Feedback) {
		defer wg.Done()

		f, err := os.Open("tmp.xml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		err = xml.Unmarshal(b, feedback)
		if err != nil {
			panic(err)
		}
	}(&feedback)

	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("*.html")),
	}
	e.Renderer = renderer
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "dmarc.html", map[string]interface{}{})
	})
	wg.Wait()

	fmt.Println(feedback.PolicyPublished.Domain)
	fmt.Println(feedback.PolicyPublished.Adkim)
	fmt.Println(feedback.PolicyPublished.Aspf)
	fmt.Println(feedback.PolicyPublished.P)
	fmt.Println(feedback.PolicyPublished.Sp)
	fmt.Println(feedback.PolicyPublished.Pct)
	for _, record := range feedback.Records {
		fmt.Println(record.Row.SourceIP)
		fmt.Println(record.Row.Count)
	}

	e.Logger.Fatal(e.Start(":4000"))
}
