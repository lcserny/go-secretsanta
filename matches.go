package gosecretsanta

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const MatchesRoute = "matches"

func InitMatchesController(router *mux.Router) {
	controller := newMatchesController(newMatchesService())
	router.HandleFunc(fmt.Sprintf("/%s", MatchesRoute), controller.generateLinks).Methods("POST")
	router.HandleFunc(fmt.Sprintf("/%s/{from}/{token}", MatchesRoute), controller.findMatch).Methods("GET")
	router.HandleFunc(fmt.Sprintf("/%s", MatchesRoute), controller.clearMatches).Methods("DELETE")
}

type NamesRequest struct {
	Names map[string][]string `json:"names"`
}

type matchesController struct {
	service *matchesService
}

func newMatchesController(service *matchesService) *matchesController {
	return &matchesController{service}
}

func (c *matchesController) generateLinks(writer http.ResponseWriter, request *http.Request) {
	var namesRequest NamesRequest
	if err := json.NewDecoder(request.Body).Decode(&namesRequest); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request received to generate links with: %v\n", namesRequest)

	matches, err := c.service.generateMatches(namesRequest.Names)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	var links []string

	protocol := "http"
	if request.TLS != nil {
		protocol = "https"
	}

	for _, pair := range matches {
		links = append(links, fmt.Sprintf("%s://%s/%s/%s/%s", protocol, request.Host, MatchesRoute, pair.name, pair.token))
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(links); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *matchesController) findMatch(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	from := vars["from"]
	token := vars["token"]
	log.Printf("Request received to find match with from: %s and token: %s\n", from, token)

	target, found := c.service.findTarget(token)
	if !found {
		if err := c.showPage("assets/invalid.html", writer, nil); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := MatchesPageData{
		From:   from,
		Target: target,
	}

	if err := c.showPage("assets/index.html", writer, data); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *matchesController) clearMatches(http.ResponseWriter, *http.Request) {
	log.Println("Request received to clear matches")
	c.service.clearMatches()
}

func (c *matchesController) showPage(templatePath string, writer http.ResponseWriter, data any) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(writer, data); err != nil {
		return err
	}
	return nil
}

type MatchesPageData struct {
	From   string
	Target string
}

type matchesService struct {
	ttMap *tokenTargetMap
}

func newMatchesService() *matchesService {
	return &matchesService{newTokenTargetMap()}
}

func (s *matchesService) clearMatches() {
	s.ttMap.clear()
}

func (s *matchesService) findTarget(token string) (string, bool) {
	return s.ttMap.getOnce(token)
}

func (s *matchesService) generateMatches(names map[string][]string) ([]matchPair, error) {
	if err := s.validateExcludes(names); err != nil {
		return nil, err
	}

	var matchPairs []matchPair
	for range 10 {
		matchPairs = s.generateMatchesInternal(names)
		if len(matchPairs) == len(names) {
			break
		}
	}

	if len(matchPairs) != len(names) {
		return nil, fmt.Errorf("could not generate correct matches from input given")
	}

	return matchPairs, nil
}

func (s *matchesService) generateMatchesInternal(names map[string][]string) []matchPair {
	var namesTaken []string
	var matches []matchPair

	for name, excludes := range names {
		var drawPool []string
		for n := range names {
			drawPool = append(drawPool, n)
		}

		drawPool = removeString(drawPool, name)
		drawPool = removeStrings(drawPool, excludes)
		drawPool = removeStrings(drawPool, namesTaken)

		if len(drawPool) == 0 {
			log.Printf("for name %s there are no options to draw from\n", name)
			continue
		}

		randomIndex := rand.Intn(len(drawPool))
		target := drawPool[randomIndex]

		namesTaken = append(namesTaken, target)

		token := uuid.New().String()
		s.ttMap.setVal(token, target)

		matches = append(matches, matchPair{name, token})
	}

	return matches
}

func (s *matchesService) validateExcludes(names map[string][]string) error {
	var allNames []string
	for name := range names {
		allNames = append(allNames, name)
	}
	sort.Strings(allNames)

	for name, excludes := range names {
		var allExcludes []string
		allExcludes = append(allExcludes, excludes...)
		allExcludes = append(allExcludes, name)
		sort.Strings(allExcludes)

		if reflect.DeepEqual(allNames, allExcludes) {
			return fmt.Errorf("exclude list for %s contains all names", name)
		}
	}

	return nil
}

func removeString(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func removeStrings(slice []string, values []string) []string {
	for _, val := range values {
		slice = removeString(slice, val)
	}
	return slice
}

type matchPair struct {
	name  string
	token string
}

type tokenTargetMap struct {
	sync.Mutex
	data map[string]string
}

func newTokenTargetMap() *tokenTargetMap {
	return &tokenTargetMap{
		data: make(map[string]string),
	}
}

func (sm *tokenTargetMap) getOnce(key string) (string, bool) {
	sm.Lock()
	defer sm.Unlock()
	value, ok := sm.data[key]
	delete(sm.data, key)
	return value, ok
}

func (sm *tokenTargetMap) setVal(key string, value string) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[key] = value
}

func (sm *tokenTargetMap) clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.data = map[string]string{}
}
