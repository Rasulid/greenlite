package main

import (
	"errors"
	"fmt"
	"net/http"

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
		Activate: false,
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

	err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		fmt.Println(err, "______________________________")
		app.serverStatusError(w,r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"users": user}, nil)
	if err != nil {
		app.serverStatusError(w,r,err)
	}

}