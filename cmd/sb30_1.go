// Lesson 30
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Database interface {
	Add(Record) (int, error)
	Delete(string) error
	Search(string) (Record, error)
}

type Record interface {
	JSONLoad([]byte) error
}

// Users storage in memory.
type UserArray struct {
	delpos int
	users  []*User
}

// Add() append user based on deleted position
func (ul *UserArray) Add(r Record) (int, error) {
	var k int
	u, ok := r.(*User)
	if !ok {
		return 0, fmt.Errorf("unexpected type for %v", u)
	}
	if ul.delpos > 0 {
		pos := ul.delpos - 1
		next := ul.users[pos].id
		u.id = ul.delpos
		u.isDel = false
		ul.users[pos] = u
		ul.delpos = next
		k = u.id
	} else {
		ul.users = append(ul.users, u)
		k = len(ul.users)
		ul.users[k-1].id = k
	}
	return k, nil
}

// Delete() find record via id and marks it as deleted
func (ul *UserArray) Delete(sid string) error {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return err
	}
	if id < 1 || id > len(ul.users) {
		return errors.New("record not found")
	}
	pos := id - 1
	if ul.users[pos].isDel {
		return errors.New("record allready delete")
	}
	ul.users[pos].isDel = true
	ul.users[pos].id = ul.delpos
	ul.delpos = id
	// fmt.Println(ul.users[pos].id, ul.delpos)
	return nil
}

func (ul *UserArray) Search(sid string) (Record, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return nil, err
	}
	for _, val := range ul.users {
		if val.isDel {
			continue
		}
		if val.id == id {
			return val, nil
		}
	}
	return nil, fmt.Errorf("user with id: %d not found", id)
}

func NewUserArray() *UserArray {
	return &UserArray{users: make([]*User, 0, 10)}
}

type User struct {
	isDel   bool   //`json:"-"`
	id      int    //`json:"id"`
	Name    string `json:"name"`
	Age     string `json:"age"`
	Friends []int  `json:"friends"`
}

func NewUser(aName string, aAge int) *User {
	return &User{
		Name:    aName,
		Age:     strconv.FormatInt(int64(aAge), 10),
		Friends: make([]int, 0),
	}
}

func NewEmptyUser() *User {
	return &User{
		Friends: make([]int, 0),
	}
}

// JSONLoad() prepare User record from json
func (u *User) JSONLoad(b []byte) error {
	err := json.Unmarshal(b, u)
	return err
}

// Business logic
type application struct {
	list Database

	infoLog  *log.Logger
	errorLog *log.Logger
}

// POST request For Create User
func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", 405)
		app.errorLog.Println("метод запрещен")
		return
	}

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	usr := NewEmptyUser()
	err = usr.JSONLoad(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(usr)
	id, err := app.list.Add(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(struct {
		Id string `json:"id"`
	}{strconv.FormatInt(int64(id), 10)})

	if err != nil {
		app.errorLog.Println(err)
	}

	app.infoLog.Println(string(b))

	w.Write(b)
}

// DELETE request for delete User
func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		http.Error(w, "Метод запрещен!", 405)
		app.errorLog.Println("метод запрещен")
		return
	}

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	var delUser struct {
		Id string `json:"target_id"`
	}
	err = json.Unmarshal(content, &delUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}

	usr, err := app.list.Search(delUser.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	user := usr.(*User)
	err = app.list.Delete(delUser.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	// TODO :: delete in friends

	app.infoLog.Println("delete user:", delUser.Id, user.Name)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Name))
}

// GET show all friends
func (app *application) friends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод запрещен!", 405)
		app.errorLog.Println("метод запрещен")
		return
	}

	var (
		id  string
		err error
	)
	s := strings.Split(strings.ToLower(r.URL.Path), "/")
	idx := -1
	for i, val := range s {
		if val == "friends" {
			idx = i + 1
			break
		}
	}
	if idx > 0 && len(s) > idx {
		id = s[idx]
		// id, err = strconv.Atoi(s[idx])
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	app.errorLog.Println(err)
		// 	return
		// }
	} else {
		err = errors.New("id not found")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	usr, err := app.list.Search(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	user := usr.(*User)

	// TODO :: Вывести всех друзей

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Name))
}

// PUT Change Age
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		http.Error(w, "Метод запрещен!", 405)
		app.errorLog.Println("метод запрещен")
		return
	}

	s := strings.Split(strings.ToLower(r.URL.Path), "/")
	id := s[1]
	// id, err := strconv.Atoi(s[1])
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	app.errorLog.Println(err)
	// 	return
	// }

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	var newage struct {
		NewAge string `json:"new agee"`
	}
	err = json.Unmarshal(content, &newage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}

	usr, err := app.list.Search(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	user := usr.(*User)

	// TODO :: update AGE

	user.Age = newage.NewAge

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("возраст пользователя успешно обновлён"))
}

// POST link two users as friend
func (app *application) makeFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", 405)
		app.errorLog.Println("метод запрещен")
		return
	}

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	var friend struct {
		Sid string `json:"source_id"`
		Tid string `json:"target_id"`
	}
	err = json.Unmarshal(content, &friend)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}

	susr, err := app.list.Search(friend.Sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	suser := susr.(*User)
	tusr, err := app.list.Search(friend.Tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	tuser := tusr.(*User)

	// TODO :: make friend

	app.infoLog.Printf("%s и %s теперь друзья", suser.Name, tuser.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s и %s теперь друзья", suser.Name, tuser.Name)))
}

func main() {
	addr := ":4000"

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		list:     NewUserArray(),
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/friends/", app.friends)
	mux.HandleFunc("/create", app.createUser)
	mux.HandleFunc("/delete", app.deleteUser)
	mux.HandleFunc("/make_friends", app.makeFriends)

	infoLog.Println("Запуск веб-сервера на http://127.0.0.1:4000")
	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
