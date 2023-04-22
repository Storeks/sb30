// Lesson 30
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
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
	l    *Logic

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

	userName, err := app.l.DeleteUser(delUser.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println("delete user:", userName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(userName))
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

	out, err := app.l.FriendsList(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(out)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
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

	err = app.l.ChangeUserAge(id, newage.NewAge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}

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

	info, err := app.l.MakeUsersFrends(friend.Sid, friend.Tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Printf(info)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(info))
}

// Business logic
type Logic struct {
	ds Database
}

func NewLogic(aDs Database) *Logic {
	return &Logic{ds: aDs}
}

func (l *Logic) AddUser() error {
	return nil
}

func (l *Logic) DeleteUser(id string) (string, error) {
	usr, err := l.ds.Search(id)
	if err != nil {
		return "", err
	}
	user := usr.(*User)
	name := user.Name

	if len(user.Friends) > 0 {
		for _, val := range user.Friends {
			u, err := l.ds.Search(strconv.FormatInt(int64(val), 10))
			if err == nil {
				if uu, ok := u.(*User); ok {
					index := -1
					for idx, v := range uu.Friends {
						if v == user.id {
							index = idx
							break
						}
					}
					if index > 0 {
						uu.Friends = append(uu.Friends[:index], uu.Friends[index+1:]...)
					}
				}
			}
		}
	}
	err = l.ds.Delete(id)
	return name, err
}

func (l *Logic) FriendsList(id string) (string, error) {
	usr, err := l.ds.Search(id)
	if err != nil {
		return "", err
	}
	user := usr.(*User)

	out := "Список друзей " + user.Name + ":\n"

	if len(user.Friends) > 0 {
		for _, val := range user.Friends {
			u, err := l.ds.Search(strconv.FormatInt(int64(val), 10))
			if err == nil {
				uu := u.(*User)
				out += "\tName: " + uu.Name + "\tAge: " + uu.Age + "\n"
			}
		}
	} else {
		out += "\t<< ПУСТО >>\n"
	}
	return out, nil
}

func (l *Logic) ChangeUserAge(id, age string) error {
	usr, err := l.ds.Search(id)
	if err != nil {
		return err
	}
	user := usr.(*User)
	user.Age = age
	return nil
}

func (l *Logic) MakeUsersFrends(sid, tid string) (string, error) {
	susr, err := l.ds.Search(sid)
	if err != nil {
		return "", err
	}
	suser := susr.(*User)
	tusr, err := l.ds.Search(tid)
	if err != nil {
		return "", err
	}
	tuser := tusr.(*User)
	for _, val := range suser.Friends {
		// Already friends
		if val == tuser.id {
			return "", fmt.Errorf("%s and %s already friends", suser.Name, tuser.Name)
		}
	}
	suser.Friends = append(suser.Friends, tuser.id)
	tuser.Friends = append(tuser.Friends, suser.id)
	return fmt.Sprintf("%s и %s теперь друзья", suser.Name, tuser.Name), nil
}

// Entry point
func main() {
	addr := ":4000"

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		list:     NewUserArray(),
		l:        NewLogic(NewUserArray()),
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
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errorLog.Fatal(err)
		}
	}()

	// Ctrl-C for canceled server work
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt)
	<-cancelChan

	infoLog.Println("Server shutdown ...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

}
