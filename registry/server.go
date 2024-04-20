package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const ServicePort = ":3000"
const ServiceURL = "http://localhost" + ServicePort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.Mutex
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

var reg = registry{registrations: make([]Registration, 0), mutex: new(sync.Mutex)}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Request received")
	switch req.Method {
	case http.MethodPost:
		dec := json.NewDecoder(req.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = reg.add(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %v", r.ServiceName, r.ServiceUrl)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}
