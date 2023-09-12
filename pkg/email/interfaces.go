package email

type EmailClient interface {
	SendMail(email Email) error
}
