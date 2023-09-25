package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url": r.URL.String(),
	})
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int,
	massage interface{}) {
	env := envelope{"error": massage}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverStatusError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.PrintError(err, nil)

	massage := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, massage)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	massage := "the response could not be found"
	app.errorResponse(w, r, http.StatusNotFound, massage)
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	massage := fmt.Sprintf("the method %s is not allowed", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, massage)
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
	}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request){
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}


func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request){
	message := "rate limit exidet"
	app.errorResponse(w,r, http.StatusTooManyRequests, message)
}