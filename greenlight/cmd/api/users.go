package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"greenlight.rasulabduvaitov.net/internal/data"
	"greenlight.rasulabduvaitov.net/internal/validator"
)


func (app *application) registrUserHendler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}


	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverStatusError(w,r,err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch{
		case errors.Is(err, data.ErrorDublicateEmail):
			v.AddErrors("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverStatusError(w,r,err)
		}
		return
	}

	token, err := app.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActiation)
	if err != nil {
		app.serverStatusError(w,r,err)
		return
	}

	app.background(func(){


		data := map[string]interface{}{
			"activationToken": token.PlainText,
			"userID": user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
		
	})
		

	err = app.writeJSON(w, http.StatusCreated, envelope{"users": user}, nil)
	if err != nil {
		app.serverStatusError(w,r,err)
	}

}



func (app *application) activateUserHendler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		TokenPlainText string  `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestError(w,r, err)
		return 
	}

	v := validator.New()

	if data.ValidateTokenPlainText(v, input.TokenPlainText); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActiation, input.TokenPlainText)
	if err != nil {
		fmt.Println("user:", user)
		fmt.Println("error:", err)
		switch{
		case errors.Is(err, data.ErrorRecordNotFound):
			v.AddErrors("token", "invalid or expired token")
			app.failedValidationResponse(w,r, v.Errors)
		default:
			app.serverStatusError(w,r,err)
		}

		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch{
		case errors.Is(err, data.ErrorEditConflict):
			app.editConflictResponse(w,r)
		default:
			app.serverStatusError(w,r,err)
		}
		return
	}


	err = app.models.Token.DeleteAllForUser(user.ID, data.ScopeActiation)
	if err != nil {
		app.serverStatusError(w,r,err)
		return
	}

	err = app.writeJSON(w, http.StatusOK ,envelope{"user": user}, nil)
	if err != nil {
		app.serverStatusError(w,r, err)
	}


}