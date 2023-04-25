// Lesson 30
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/storeks/sb30/internal/database"
	"github.com/storeks/sb30/internal/logic"
)

// Business logic
type application struct {
	l *logic.Logic

	infoLog  *log.Logger
	errorLog *log.Logger
}

// POST request For Create User
func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	w.Header().Set("Allow", http.MethodPost)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}

	usr := database.NewEmptyUser()
	err = usr.JSONLoad(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(usr)
	id, err := app.l.AddUser(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
	// if r.Method != http.MethodDelete {
	// 	w.Header().Set("Allow", http.MethodDelete)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println("delete user:", userName)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(userName))
}

// GET show all friends
func (app *application) friends(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	w.Header().Set("Allow", http.MethodGet)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }
	//
	// var (
	// 	id  string
	// 	err error
	// )
	// s := strings.Split(strings.ToLower(r.URL.Path), "/")
	// idx := -1
	// for i, val := range s {
	// 	if val == "friends" {
	// 		idx = i + 1
	// 		break
	// 	}
	// }
	// if idx > 0 && len(s) > idx {
	// 	id = s[idx]
	// 	// id, err = strconv.Atoi(s[idx])
	// 	// if err != nil {
	// 	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	// 	app.errorLog.Println(err)
	// 	// 	return
	// 	// }
	// } else {
	// 	err = errors.New("id not found")
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	app.errorLog.Println(err)
	// 	return
	// }
	id := chi.URLParam(r, "idUser")

	out, err := app.l.FriendsList(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(out)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}

// PUT Change Age
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPut {
	// 	w.Header().Set("Allow", http.MethodPut)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }
	//
	// s := strings.Split(strings.ToLower(r.URL.Path), "/")
	// id := s[1]
	// id, err := strconv.Atoi(s[1])
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	app.errorLog.Println(err)
	// 	return
	// }

	id := chi.URLParam(r, "idUser")

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("возраст пользователя успешно обновлён"))
}

// POST link two users as friend
func (app *application) makeFriends(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	w.Header().Set("Allow", http.MethodPost)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }

	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Printf(info)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(info))
}

// GET show all users for test
func (app *application) showAll(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	w.Header().Set("Allow", http.MethodGet)
	// 	http.Error(w, "Метод запрещен!", 405)
	// 	app.errorLog.Println("метод запрещен")
	// 	return
	// }

	out, err := app.l.ShowAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(out)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}

// Entry point
func main() {
	addr := ":4000"

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		l:        logic.NewLogic(database.NewUserArray()),
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	rout := chi.NewRouter()
	rout.Get("/show", app.showAll)
	rout.Put("/{idUser}", app.home)
	rout.Delete("/delete", app.deleteUser)
	rout.Post("/make_friends", app.makeFriends)
	rout.Post("/create", app.createUser)
	rout.Get("/friends/{idUser}", app.friends)

	// mux := http.NewServeMux()
	// mux.HandleFunc("/", app.home)
	// mux.HandleFunc("/friends/", app.friends)
	// mux.HandleFunc("/create", app.createUser)
	// mux.HandleFunc("/delete", app.deleteUser)
	// mux.HandleFunc("/make_friends", app.makeFriends)
	// mux.HandleFunc("/show", app.showAll)

	infoLog.Println("Запуск веб-сервера на http://127.0.0.1:4000")
	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  rout, //mux,
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
