// sb30_tstServer
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Name    string `json:"name"`
	Age     string `json:"age"`
	Friends []int  `json:"friends"`
}

const sURL = "http://127.0.0.1:4000"

func main() {
	fmt.Println("Client for test.")

	Post(User{"John Doe", "15", []int{}})
	fmt.Println("------------------[POST")
	Post(User{"Morlon Pir", "35", []int{3}})
	fmt.Println("------------------[POST")
	Post(User{"VChif", "50", []int{}})
	fmt.Println("------------------[POST")
	Post(User{"Olegus Morky", "90", []int{3}})
	fmt.Println("------------------[DELETE")
	Delete(3)
	fmt.Println("------------------[POST")
	Post(User{"Lyto By", "17", []int{4, 2}})
	fmt.Println("------------------[POST")
	Post(User{"Konnon Varvar", "25", []int{}})
	fmt.Println("------------------[GET FR")
	GetFrieds("3")
}

func GetFrieds(userId string) {
	resp, err := http.Get(sURL + "/friends/" + userId)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Println("Status:", resp.Status)
	fmt.Println(resp.Header.Get("Content-Type"))
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", string(data))
}

func Post(u User) {
	json_data, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(sURL+"/create", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Println("Status:", resp.Status)
	fmt.Println(resp.Header.Get("Content-Type"))
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", string(data))
}

func Delete(UserId int) {
	client := &http.Client{Timeout: 30 * time.Second}
	post := `{"target_id":"` + strings.TrimSpace(strconv.FormatInt(int64(UserId), 10)) + `"}`
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, sURL+"/delete", strings.NewReader(post))
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	// if res.StatusCode != http.StatusOK {
	// 	panic(fmt.Sprintf("unexpected status: got %v", res.Status))
	// }
	fmt.Println("Status:", res.Status)
	fmt.Println(res.Header.Get("Content-Type"))
	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	// var data struct {
	// 	UserID    int    `json:"userId"`
	// 	ID        int    `json:"id"`
	// 	Title     string `json:"title"`
	// 	Completed bool   `json:"completed"`
	// }
	// err = json.NewDecoder(res.Body).Decode(&data)
	// if err != nil {
	// 	panic(err)
	// }
	fmt.Printf("%+v\n", string(data))
}
