package constants

import (
	"fmt"
)

const AppName string = "makisama"

var SessionName string = fmt.Sprintf("%s_session", AppName)
var FlashSessionKey string = fmt.Sprintf("%s_flash_k7Qm2vL9xP4d", AppName)
var SessionManagerSessionKey string = fmt.Sprintf("%s_authid_T3nR8cX5qZ1w", AppName)

const MailtrapUrl string = "https://send.api.mailtrap.io/api/send"
const TurnstileCaptchaSiteVerify = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
const TurnstileCaptchaFormName = "cf-turnstile-response"
const TemplateDataContextKey string = "templateData"

const PerPageConstant int = 10
const BcryptCost int = 12

const AlphaNumericLowerCharset string = "abcdefghijklmnopqrstuvwxyz0123456789"
const AlphaNumericUpperCharset string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const AlphaNumericCharset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const NumericCharset string = "0987654321"

const MaxImageUploadSize int = 2 << 20 // 2MiB
var BoardUploadImageDirectory string = "./images/board"

var LogDirectory string = "./logs"

var AllowedImageMime = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/bmp":  {},
}

// Version 2, make this amp
const (
	Error400Title   string = "Bad Request"
	Error400Message string = "The submitted form could not be processed. Please check your input and try again."
	Error401Title   string = "Unauthorized"
	Error401Message string = "You need to be authorized to view this page"
	Error403Title   string = "Forbidden"
	Error403Message string = "You do not have permission to perform this action"
	Error404Title   string = "Not Found"
	Error404Message string = "The page you are looking for might have been removed, had its name changed, is temporarily unavailable or simply does not exist."
	Error405Title   string = "Method Not Allowed"
	Error405Message string = "The HTTP request method used for this URL path is not permitted by our server rules."
	Error429Title   string = "Too Many Requests"
	Error429Message string = "You are making requests too rapidly. Please pause for a moment, wait a minute, and try again."
	Error500Title   string = "Server Error"
	Error500Message string = "Server error. Try again later."
)

const BannedMessage string = "Email address associated with this user is currently banned"

const (
	FirstnameMaxLength           int = 50
	FirstnameMinLength           int = 2
	LastnameMaxLength            int = 50
	LastnameMinLength            int = 2
	PasswordMaxLength            int = 70
	PasswordMinLength            int = 8
	AuthidMaxLength              int = 64
	AuthidMinLength              int = 32
	PidMaxLength                 int = 16
	PidMinLength                 int = 32
	EmailMaxLength               int = 255
	EmailMinLength               int = 5
	FeedbackSlugMinLength        int = 50
	FeedbackSlugMaxLength        int = 100
	FeedbackTitleMinLength       int = 3
	FeedbackTitleMaxLength       int = 100
	FeedbackDescriptionMaxLength int = 10000
	BoardTitleMinLength          int = 5
	BoardTitleMaxLength          int = 60
	BoardWebsiteMinLength        int = 0
	BoardWebsiteMaxLength        int = 200
	BoardDescriptionMaxLength    int = 1000
	ReportCommentMaxLength       int = 10000
)

const RateLimit int = 60

var FeedbackResolutions = map[string]struct{}{
	"under discussion": {},
	"in progress":      {},
	"resolved":         {},
	"out of scope":     {},
	"rejected":         {},
	"duplicate":        {},
	"won't fix":        {},
}

var FeedbackScopes = map[string]struct{}{
	"suggestion":  {},
	"issue":       {},
	"improvement": {},
}

var ReservedWords = map[string]struct{}{
	"dashboard":      {},
	"account":        {},
	"accounts":       {},
	"auth":           {},
	"auths":          {},
	"authenticate":   {},
	"authenticates":  {},
	"board":          {},
	"boards":         {},
	"api":            {},
	"apis":           {},
	"feedbacks":      {},
	"feedback":       {},
	"report":         {},
	"reports":        {},
	"new":            {},
	"news":           {},
	"company":        {},
	"about":          {},
	"about_us":       {},
	"contact":        {},
	"contact_us":     {},
	"verify_account": {},
	"verify":         {},
	"search":         {},
	"searches":       {},
}

var ReportReasons = map[string]struct{}{
	"Spam":                                {},
	"Harassment or Bullying":              {},
	"Hate Speech":                         {},
	"Violence or Threats":                 {},
	"Self-Harm or Suicide Content":        {},
	"Misinformation or False Information": {},
	"Sexual Content":                      {},
	"Child Safety Concerns":               {},
	"Privacy Violations":                  {},
	"Impersonation":                       {},
	"Fraud or Scams":                      {},
	"Copyright Infringement":              {},
	"Trademark Infringement":              {},
	"Illegal Activity":                    {},
	"Malware or Phishing":                 {},
	"Graphic or Disturbing Content":       {},
	"Others":                              {},
}
