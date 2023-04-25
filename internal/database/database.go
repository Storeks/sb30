package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

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
		next := ul.users[pos].Id
		u.Id = ul.delpos
		u.isDel = false
		ul.users[pos] = u
		ul.delpos = next
		k = u.Id
	} else {
		ul.users = append(ul.users, u)
		k = len(ul.users)
		ul.users[k-1].Id = k
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
	ul.users[pos].Id = ul.delpos
	ul.delpos = id
	// fmt.Println(ul.users[pos].Id, ul.delpos)
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
		if val.Id == id {
			return val, nil
		}
	}
	return nil, fmt.Errorf("user with id: %d not found", id)
}

func (ul *UserArray) Map(f func(Record)) {
	for _, val := range ul.users {
		if !val.isDel {
			f(val)
		}
	}
}

func NewUserArray() *UserArray {
	return &UserArray{users: make([]*User, 0, 10)}
}

type User struct {
	isDel   bool   //`json:"-"`
	Id      int    `json:"-"`
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
