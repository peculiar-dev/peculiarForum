package email

type EmailConnector interface {
	SendNotification(to, body string) error
}
