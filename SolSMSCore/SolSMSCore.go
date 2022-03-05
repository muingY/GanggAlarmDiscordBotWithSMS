package SolSMSCore

import (
	"github.com/solapi/solapi-go"
	"github.com/solapi/solapi-go/types"
)

type SolSMSCore struct {
	client *solapi.Client
}

func (sol *SolSMSCore) Initialize() {
	sol.client = solapi.NewClient()
}

func (sol *SolSMSCore) SendSMS(to string, from string, msg string) (types.SimpleMessage, error) {
	message := make(map[string]interface{})
	message["to"] = to
	message["from"] = from
	message["text"] = msg
	message["type"] = "SMS"

	params := make(map[string]interface{})
	params["message"] = message

	result, err := sol.client.Messages.SendSimpleMessage(params)
	return result, err
}
