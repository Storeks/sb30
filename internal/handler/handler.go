package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/storeks/sb30/internal/database"
	"github.com/storeks/sb30/internal/logic"
)

// Business logic
type Handler struct {
	l        *logic.Logic
	infoLog  *log.Logger
	errorLog *log.Logger
	H        *chi.Mux
}

func NewHandler(i, e *log.Logger) *Handler {
	return &Handler{
		l:        logic.NewLogic(database.NewUserArray()),
		infoLog:  i,
		errorLog: e,
		H:        chi.NewRouter(),
	}
}

func (h *Handler) Init() *chi.Mux {
	h.H.Get("/show", h.ShowAll)
	return h.InitApi()
}

func (h *Handler) InitApi() *chi.Mux {
	h.H.Put("/{idUser}", h.Home)
	h.H.Delete("/delete", h.DeleteUser)
	h.H.Post("/make_friends", h.MakeFriends)
	h.H.Post("/create", h.CreateUser)
	h.H.Get("/friends/{idUser}", h.Friends)
	return h.H
}

// POST request For Create User
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	usr := database.NewEmptyUser()
	err = usr.JSONLoad(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errorLog.Println(err)
		return
	}

	h.infoLog.Println(usr)
	id, err := h.l.AddUser(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(struct {
		Id string `json:"id"`
	}{strconv.FormatInt(int64(id), 10)})

	if err != nil {
		h.errorLog.Println(err)
	}

	h.infoLog.Println(string(b))

	w.Write(b)
}

// DELETE request for delete User
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	var delUser struct {
		Id string `json:"target_id"`
	}
	err = json.Unmarshal(content, &delUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errorLog.Println(err)
		return
	}

	userName, err := h.l.DeleteUser(delUser.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	h.infoLog.Println("delete user:", userName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(userName))
}

// GET show all friends
func (h *Handler) Friends(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "idUser")

	out, err := h.l.FriendsList(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}
	h.infoLog.Println(out)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}

// PUT Change Age
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "idUser")

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	var newage struct {
		NewAge string `json:"new agee"`
	}
	err = json.Unmarshal(content, &newage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errorLog.Println(err)
		return
	}

	err = h.l.ChangeUserAge(id, newage.NewAge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("возраст пользователя успешно обновлён"))
}

// POST link two users as friend
func (h *Handler) MakeFriends(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}

	var friend struct {
		Sid string `json:"source_id"`
		Tid string `json:"target_id"`
	}

	err = json.Unmarshal(content, &friend)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errorLog.Println(err)
		return
	}

	info, err := h.l.MakeUsersFrends(friend.Sid, friend.Tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}
	h.infoLog.Printf(info)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(info))
}

// GET show all users for test
func (h *Handler) ShowAll(w http.ResponseWriter, r *http.Request) {
	out, err := h.l.ShowAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		h.errorLog.Println(err)
		return
	}
	h.infoLog.Println(out)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}
