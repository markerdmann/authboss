package authboss

import (
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Cfg is the singleton instance of Config
var Cfg *Config = NewConfig()

// Config holds all the configuration for both authboss and it's modules.
type Config struct {
	// MountPath is the path to mount authboss's routes at (eg /auth).
	MountPath string
	// ViewsPath is the path to search for overridden templates.
	ViewsPath string
	// RootURL is the scheme+host+port of the web application (eg https://www.happiness.com:8080) for url generation.  No trailing slash.
	RootURL string
	// BCryptCost is the cost of the bcrypt password hashing function.
	BCryptCost int

	// PrimaryID is the primary identifier of the user. Set to one of:
	// authboss.StoreEmail, authboss.StoreUsername (StoreEmail is default)
	PrimaryID string

	Layout          *template.Template
	LayoutHTMLEmail *template.Template
	LayoutTextEmail *template.Template
	LayoutDataMaker ViewDataMaker

	OAuth2Providers map[string]OAuthProvider

	// ErrorHandler handles would be 500 errors.
	ErrorHandler http.Handler
	// BadRequestHandler handles would be 400 errors.
	BadRequestHandler http.Handler
	// NotFoundHandler handles would be 404 errors.
	NotFoundHandler http.Handler

	AuthLoginOKPath   string
	AuthLoginFailPath string
	AuthLogoutOKPath  string

	RecoverOKPath        string
	RecoverTokenDuration time.Duration

	Policies      []Validator
	ConfirmFields []string

	ExpireAfter time.Duration

	LockAfter    int
	LockWindow   time.Duration
	LockDuration time.Duration

	EmailFrom          string
	EmailSubjectPrefix string
	SMTPAddress        string
	SMTPAuth           smtp.Auth

	XSRFName  string
	XSRFMaker XSRF

	Storer            Storer
	OAuth2Storer      OAuth2Storer
	CookieStoreMaker  CookieStoreMaker
	SessionStoreMaker SessionStoreMaker
	LogWriter         io.Writer
	Callbacks         *Callbacks
	Mailer            Mailer
}

func NewConfig() *Config {
	return &Config{
		MountPath:  "/",
		ViewsPath:  "/",
		RootURL:    "http://localhost:8080",
		BCryptCost: bcrypt.DefaultCost,

		PrimaryID: StoreEmail,

		Layout:          template.Must(template.New("").Parse(`<!DOCTYPE html><html><body>{{template "authboss" .}}</body></html>`)),
		LayoutHTMLEmail: template.Must(template.New("").Parse(`<!DOCTYPE html><html><body>{{template "authboss" .}}</body></html>`)),
		LayoutTextEmail: template.Must(template.New("").Parse(`{{template "authboss" .}}`)),

		AuthLoginOKPath:   "/",
		AuthLoginFailPath: "/",
		AuthLogoutOKPath:  "/",

		Policies: []Validator{
			Rules{
				FieldName:       "username",
				Required:        true,
				MinLength:       2,
				MaxLength:       4,
				AllowWhitespace: false,
			},
			Rules{
				FieldName: "password",
				Required:  true,
				MinLength: 4,
				MaxLength: 8,

				AllowWhitespace: false,
			},
		},
		ConfirmFields: []string{
			StorePassword, ConfirmPrefix + StorePassword,
		},

		ExpireAfter: 60 * time.Minute,

		RecoverOKPath:        "/",
		RecoverTokenDuration: time.Duration(24) * time.Hour,

		LogWriter: ioutil.Discard,
		Callbacks: NewCallbacks(),
		Mailer:    LogMailer(ioutil.Discard),
	}
}
