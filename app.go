package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	validationURL = "https://eshop.edalnice.cz/api/v3/charge_registrations"
	tokenURL      = "https://auth.edalnice.cz/auth/connect/token"
	clientID      = "eshop.client"
	clientSecret  = "5qz3XQuAm__bJUgADTCyP*"
	czechID       = "3906ba89-153c-4038-8e36-0ca1deb76076"
	port          = ":8080"
)

type ValidationResponse struct {
	Vehicle                    Vehicle  `json:"vehicle"`
	IsGivenExemption           bool     `json:"isGivenExemption"`
	PossibleExemptionReasonIDs []string `json:"possibleExemptionReasonIds"`
	Charges                    []Charge `json:"charges"`
}

type Vehicle struct {
	LicensePlate string `json:"licensePlate"`
	CountryID    string `json:"countryId"`
}

type Charge struct {
	PriceListItemID  string    `json:"priceListItemId"`
	ValidSince       time.Time `json:"validSince"`
	ValidUntil       time.Time `json:"validUntil"`
	IsCurrentlyValid bool      `json:"isCurrentlyValid"`
}

type Response struct {
	ID         string    `json:"id"`
	Valid      bool      `json:"valid"`
	ValidUntil time.Time `json:"validUntil"`
}

type App struct {
	Router *mux.Router
	Logger *zerolog.Logger
	Token  string
}

func (a *App) Initialize() {
	var err error
	l := &Logger{
		TimeFormat: time.RFC3339,
		Level:      zerolog.TraceLevel,
	}

	t := &Token{}
	a.Logger = l.loggerInitialize()
	a.Router = mux.NewRouter()
	a.Token, err = t.getToken(tokenURL, clientID, clientSecret)
	if err != nil {
		a.Logger.Fatal().Err(err)
	}
	a.initializeRoutes()
}

func (a *App) Run(port string) {
	fmt.Printf("Serving on http://localhost%s \n", port)
	a.Logger.Fatal().Err(http.ListenAndServe(port, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/validation/{id}", a.validateCar).Methods("GET")
}

func (a *App) validateCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	a.Logger.Info().Msg(id)

	URL := fmt.Sprintf("%s/%s/%s", validationURL, czechID, id)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		a.Logger.Info().Err(err)
	}

	defer res.Body.Close()
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		a.Logger.Info().Err(err)
	}

	// Parse the response
	var validationResponse ValidationResponse
	if err := json.Unmarshal(body, &validationResponse); err != nil {
		a.Logger.Info().Msg(fmt.Sprintf("response failed: %v", err))
	}

	var validTimesList []time.Time

	for _, charge := range validationResponse.Charges {
		if charge.IsCurrentlyValid {
			validTimesList = append(validTimesList, charge.ValidUntil)
		}
	}

	response := Response{
		ID:         validationResponse.Vehicle.LicensePlate,
		Valid:      isValid(validationResponse.Charges),
		ValidUntil: highestValidTime(time.Now(), validTimesList),
	}

	respondWithJSON(w, http.StatusOK, response)

}

func isValid(charges []Charge) bool {
	for _, charge := range charges {
		if charge.IsCurrentlyValid {
			return true
		}
	}
	return false
}

func highestValidTime(now time.Time, validTimes []time.Time) time.Time {

	highestValidTime := now
	for _, validTime := range validTimes {

		if validTime.After(highestValidTime) {
			highestValidTime = validTime
		}
	}
	return highestValidTime
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
