package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type Id = uuid.UUID

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Biography string `json:"biography"`
}

type UserResponse struct {
	Id   Id   `json:"id"`
	User User `json:"user"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJSON(w http.ResponseWriter, resp Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal json data", err)
		sendJSON(
			w,
			Response{Error: "something went wrong"},
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(statusCode)
	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write response", "error", err)
		return
	}
}

type ApplicationDB struct {
	Data map[Id]User
}

func NewHandler(db ApplicationDB) http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Post("/", handlePost(db))
	r.Get("/", handleGet(db))
	r.Get("/{id}", handleGetById(db))
	r.Put("/{id}", handlePut(db))
	r.Delete("/{id}", handleDelete(db))

	return r
}

func handlePost(db ApplicationDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		if user.FirstName == "" || user.LastName == "" || user.Biography == "" {
			sendJSON(w, Response{Error: "invalid body - missing data"}, http.StatusBadRequest)
		}

		newId := genNewId()
		db.Data[newId] = user

		sendJSON(w, Response{Data: UserResponse{User: user, Id: newId}}, http.StatusCreated)
	}
}

func handleGet(db ApplicationDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, Response{Data: db.Data}, http.StatusOK)
	}
}

func handleGetById(db ApplicationDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userId := uuid.MustParse(id)
		data, ok := db.Data[userId]
		if !ok {
			sendJSON(w, Response{Error: "user not found"}, http.StatusNotFound)
		}

		sendJSON(w, Response{Data: data}, http.StatusOK)
	}
}

func handlePut(db ApplicationDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userId := uuid.MustParse(id)

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		if user.FirstName == "" || user.LastName == "" || user.Biography == "" {
			sendJSON(w, Response{Error: "invalid body - missing data"}, http.StatusBadRequest)
		}

		if _, ok := db.Data[userId]; !ok {
			sendJSON(w, Response{Error: "user not found"}, http.StatusNotFound)
		}

		db.Data[userId] = user

		sendJSON(w, Response{Data: UserResponse{User: user, Id: userId}}, http.StatusCreated)
	}
}

func handleDelete(db ApplicationDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userId := uuid.MustParse(id)

		data, ok := db.Data[userId]

		if !ok {
			sendJSON(w, Response{Error: "user not found"}, http.StatusNotFound)
		}

		delete(db.Data, userId)

		sendJSON(w, Response{Data: data}, http.StatusOK)
	}
}

func genNewId() Id {
	return uuid.New()
}
