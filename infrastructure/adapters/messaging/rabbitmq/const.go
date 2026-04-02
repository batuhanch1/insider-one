package rabbitmq

const (
	Exchange_CreateEmail  = "Insider.One.Notification.Create.Email"
	Exchange_CancelEmail  = "Insider.One.Notification.Cancel.Email"
	Exchange_EmailCreated = "Insider.One.Notification.Email.Created"

	Exchange_CreatePush  = "Insider.One.Notification.Create.Push"
	Exchange_CancelPush  = "Insider.One.Notification.Cancel.Push"
	Exchange_PushCreated = "Insider.One.Notification.Push.Created"

	Exchange_CreateSms  = "Insider.One.Notification.Create.Sms"
	Exchange_CancelSms  = "Insider.One.Notification.Cancel.Sms"
	Exchange_SmsCreated = "Insider.One.Notification.Sms.Created"
)

const (
	Queue_CreateEmail_Generic  = "Insider.One.Notification.Create.Email.%s"
	Queue_EmailCreated_Generic = "Insider.One.Notification.Email.Created.%s"

	Queue_CreatePush_Generic  = "Insider.One.Notification.Create.Push.%s"
	Queue_PushCreated_Generic = "Insider.One.Notification.Push.Created.%s"

	Queue_CreateSms_Generic  = "Insider.One.Notification.Create.Sms.%s"
	Queue_SmsCreated_Generic = "Insider.One.Notification.Sms.Created.%s"

	Queue_CancelEmail = "Insider.One.Notification.Cancel.Email"
	Queue_CancelPush  = "Insider.One.Notification.Cancel.Push"
	Queue_CancelSms   = "Insider.One.Notification.Cancel.Sms"
)

const (
	RoutingKey_Asterisk = "*"
	RoutingKey_High     = "HIGH"
	RoutingKey_Medium   = "MEDIUM"
	RoutingKey_Low      = "LOW"
	RoutingKey_Generic  = "%s"
)

func IsPriorityRoutingKeyValid(key string) bool {
	switch key {
	case RoutingKey_High:
		return true
	case RoutingKey_Medium:
		return true
	case RoutingKey_Low:
		return true
	default:
		return false
	}
}
