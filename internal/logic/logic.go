// logic - buisness logic
package logic

import (
  "github.com/storeks/sb30/internal/database"
  "strconv"
  "fmt"
)


type Database interface {
	Add(database.Record) (int, error)
	Delete(string) error
	Search(string) (database.Record, error)
	Map(func(database.Record))
}

// Business logic
type Logic struct {
	ds Database
}

func NewLogic(aDs Database) *Logic {
	return &Logic{ds: aDs}
}

func (l *Logic) AddUser(u *database.User) (int, error) {
	return l.ds.Add(u)
}

func (l *Logic) DeleteUser(id string) (string, error) {
	usr, err := l.ds.Search(id)
	if err != nil {
		return "", err
	}
	user := usr.(*database.User)
	name := user.Name

	if len(user.Friends) > 0 {
		for _, val := range user.Friends {
			u, err := l.ds.Search(strconv.FormatInt(int64(val), 10))
			if err == nil {
				if uu, ok := u.(*database.User); ok {
					index := -1
					for idx, v := range uu.Friends {
						if v == user.Id {
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
	user := usr.(*database.User)

	out := "Список друзей " + user.Name + ":\n"

	if len(user.Friends) > 0 {
		for _, val := range user.Friends {
			u, err := l.ds.Search(strconv.FormatInt(int64(val), 10))
			if err == nil {
				uu := u.(*database.User)
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
	user := usr.(*database.User)
	user.Age = age
	return nil
}

func (l *Logic) MakeUsersFrends(sid, tid string) (string, error) {
	susr, err := l.ds.Search(sid)
	if err != nil {
		return "", err
	}
	suser := susr.(*database.User)
	tusr, err := l.ds.Search(tid)
	if err != nil {
		return "", err
	}
	tuser := tusr.(*database.User)
	for _, val := range suser.Friends {
		// Already friends
		if val == tuser.Id {
			return "", fmt.Errorf("%s and %s already friends", suser.Name, tuser.Name)
		}
	}
	suser.Friends = append(suser.Friends, tuser.Id)
	tuser.Friends = append(tuser.Friends, suser.Id)
	return fmt.Sprintf("%s и %s теперь друзья", suser.Name, tuser.Name), nil
}

func (l *Logic) ShowAllUsers() (string, error) {
	out := "Список пользователей:\n"
	l.ds.Map(func(r database.Record) {
		u := r.(*database.User)
		out += fmt.Sprintln(u)
	})
	return out, nil
}

// end
