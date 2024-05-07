package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type registry struct {
	mutex         *sync.RWMutex
	registrations []Registration
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

var reg = registry{
	mutex:         new(sync.RWMutex),
	registrations: make([]Registration, 0),
}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Registry Service: request received")
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
		log.Printf("Added %v", r.ServiceName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
