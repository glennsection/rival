package controllers

import (
	"fmt"
	"net/http"

	"bloodtales/system"
	"bloodtales/models"
)

func InsertUser(w http.ResponseWriter, r *http.Request, application *system.Application) {
	// parse parameters
	email, password := r.FormValue("email"), r.FormValue("password")
	if email == "" || password == "" {
		fmt.Fprintf(w, "Invalid request params")
		return
	}

	// check existing user
	user, err := models.GetUserByEmail(application.DB, email)
	if user != nil {
		fmt.Fprintf(w, "User already exists with email: " + email)
		return
	}

	// create user
	user = &models.User {
		Username: email,
		Email:    email,
	}
	user.HashPassword(password)

	// insert user
	if err = models.InsertUser(application.DB, user); err != nil {
		panic(err)
	}

	fmt.Fprint(w, "User inserted successfully")
}

func GetUser(w http.ResponseWriter, r *http.Request, application *system.Application) {
	// parse parameters
	email := r.FormValue("email")
	if email == "" {
		fmt.Fprintf(w, "Invalid request params")
		return
	}

	// get user
	user, _ := models.GetUserByEmail(application.DB, email)
	if user != nil {
		fmt.Fprintf(w, "Found user: %v (%v)", user.Username, user.Email)
	} else {
		fmt.Fprintf(w, "Failed to find User with email: %v", email)
	}
}

func LoginUser(w http.ResponseWriter, r *http.Request, application *system.Application) {
	// parse parameters
	email, password := r.FormValue("email"), r.FormValue("password")
	if email == "" || password == "" {
		fmt.Fprintf(w, "Invalid request params")
		return
	}

	// authenticate user
	user, _ := models.LoginUser(application.DB, email, password)
	if user != nil {
		fmt.Fprintf(w, "Authenticated user: %v (%v)", user.Username, user.Email)
	} else {
		fmt.Fprintf(w, "Failed authenticate User with email: %v", email)
	}
}