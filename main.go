package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mohamedattahri/mail"
	"github.com/versine/loginauth"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

type handler struct {
	Config *config
}

func (h *handler) formatForm(w http.ResponseWriter, r *http.Request) error {
	n := strings.Join([]string{
		r.Form.Get("voornaam"),
		r.Form.Get("tussenvoegsels"),
		r.Form.Get("achternaam"),
	}, " ")
	regexp.MustCompile(`\s+`).ReplaceAllString(n, " ")
	r.Form.Set("naam", n)

	return nil
}

func (h *handler) mailForm(w http.ResponseWriter, r *http.Request) error {
	msg := mail.NewMessage()

	msg.SetFrom(&mail.Address{"Vlag & Wimpel", "contact@vlagenwimpel.com"})
	msg.SetSubject("Aanmelding lidmaatschap `" + r.Form.Get("naam") + "`")
	msg.SetContentType("text/plain")

	fl := []string{
		"naam",
		"mail",
		"telefoon",
		"socials",
		"motivatie",
	}
	for _, f := range fl {
		fmt.Fprintln(msg.Body, "# "+strings.ToUpper(f))
		fmt.Fprintln(msg.Body, r.Form.Get(f))
		fmt.Fprintln(msg.Body)
	}

	auth := loginauth.New(h.Config.MailUser, h.Config.MailPass, h.Config.
		MailHost)
	return smtp.SendMail(h.Config.MailHost+":"+strconv.Itoa(h.Config.MailPort),
		auth, h.Config.MailUser, []string{h.Config.MailUser}, msg.Bytes())
}

func main() {
	c, err := parseConfig()
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h := &handler{Config: c}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", 405)
			return
		}

		if err = r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		rc, err := recaptcha.NewReCAPTCHA(c.ReCAPTCHAKey, recaptcha.V3, 10*time.
			Second)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if err := rc.Verify(r.Form.Get("token")); err != nil {
			http.Error(w, err.Error(), 429)
			return
		}

		if err = h.formatForm(w, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err = h.mailForm(w, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, "/", 301)
	})
	if err := http.ListenAndServe(c.Listen, nil); err != nil {
		log.Fatalln(err)
	}
}
