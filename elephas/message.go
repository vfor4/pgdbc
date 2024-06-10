package elephas

// https://www.postgresql.org/docs/current/protocol-message-formats.html
const (
	commandCompleteMsg = 'C'
	errorResponseMsg   = 'E'
	noticeResponseMsg  = 'N'
	parameterStatusMsg = 'S'
	authMsgType        = 'R'
	backendKeyDataMsg  = 'K'
	noDataMsg          = 'n'
	passwordMessageMsg = 'p'
	terminateMsg       = 'X'

	SASL         = 10
	SASLContinue = 11
	SASLComplete = 12
)
