package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const ServicePort = ":3000"
const ServiceURL = "http://localhost" + ServicePort + "/services"

func RegisterService(r Registration) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		return err
	}
	res, err := http.Post(ServiceURL, "application/json", buf)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to register service: %v, got status code: %v", r.ServiceName, res.StatusCode)
	}
	return nil
}
