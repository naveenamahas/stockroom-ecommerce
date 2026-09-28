package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
	"time"
)

func signupHandler(w http.ResponseWriter, r *http.Request) {
	var request SignupRequest
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		errorJSON(w, 400, "invalid JSON")
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	request.Email = normalizedEmail(request.Email)
	if request.Name == "" || request.Email == "" || len(request.Password) < 6 {
		errorJSON(w, 400, "name, email and password (6+ chars) required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var existing User
	err := users.FindOne(ctx, bson.M{"email": request.Email}).Decode(&existing)
	if err == nil {
		errorJSON(w, 409, "email already registered")
		return
	}
	if err != mongo.ErrNoDocuments {
		errorJSON(w, 500, "MongoDB error")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		errorJSON(w, 500, "password hashing failed")
		return
	}
	if _, err = users.InsertOne(ctx, User{Name: request.Name, Email: request.Email, PasswordHash: string(hash)}); err != nil {
		errorJSON(w, 500, "signup failed")
		return
	}
	writeJSON(w, 201, map[string]string{"message": "signup successful"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		errorJSON(w, 400, "invalid JSON")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var user User
	if err := users.FindOne(ctx, bson.M{"email": normalizedEmail(request.Email)}).Decode(&user); err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)) != nil {
		errorJSON(w, 401, "invalid email or password")
		return
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		errorJSON(w, 500, "session creation failed")
		return
	}
	token := hex.EncodeToString(bytes)
	if err := rdb.Set(ctx, "session:"+token, user.ID.Hex(), 24*time.Hour).Err(); err != nil {
		errorJSON(w, 500, "Redis session failed")
		return
	}
	isSecure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true") || strings.HasPrefix(r.Host, "localhost") == false
	if isSecure {
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteNoneMode, Secure: true, MaxAge: 86400})
	} else {
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	}
	writeJSON(w, 200, map[string]string{"message": "login successful"})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ctx, cancel, ok := sessionUserID(w, r)
	if !ok {
		return
	}
	defer cancel()
	var user User
	if err := users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		errorJSON(w, 404, "user not found")
		return
	}
	writeJSON(w, 200, map[string]string{"name": user.Name, "email": user.Email})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_id"); err == nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_ = rdb.Del(ctx, "session:"+cookie.Value).Err()
	}
	isSecure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true") || !strings.HasPrefix(r.Host, "localhost")
	if isSecure {
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteNoneMode, Secure: true})
	} else {
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
	}
	writeJSON(w, 200, map[string]string{"message": "logout successful"})
}
