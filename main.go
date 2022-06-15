package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"

	"os"
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (u *User) String() string {
	if u.IsEmpty() {
		return ""

	}
	b, _ := json.Marshal(*u)
	return string(b)
}

func (u *User) Marshal() ([]byte, error) {
	return json.Marshal(*u)
}

func (u *User) Set(s string) error {
	return json.Unmarshal([]byte(s), u)
}

func (u *User) IsEmpty() bool {
	return len(u.Id) == 0
}

type Arguments map[string]string

func parseArgs() Arguments {
	var user User
	var operation, fileName, id string
	flag.StringVar(&operation, "operation", "", `-operation "add"`)
	flag.Var(&user, "item", `-item {id: "1", email: «test@test.com», age: 31}`)
	flag.StringVar(&fileName, "fileName", "", `-fileName "users.json"`)
	flag.StringVar(&id, "id", "", `-id "2"`)
	flag.Parse()

	args := Arguments{}
	args["operation"] = operation
	args["item"] = user.String()
	args["fileName"] = fileName
	args["id"] = id
	return args
}

func fileOpen(fName, op string) (f *os.File, users []User, err error) {
	switch op {
	case "add", "remove":
		f, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, users, fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
	case "findById", "list":
		f, err = os.OpenFile(fName, os.O_RDONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, users, fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
	default:
		return nil, users, fmt.Errorf("Operation %s not allowed!", op)
	}

	body, err := ioutil.ReadAll(f)
	if err != nil && err != io.EOF {
		return nil, users, fmt.Errorf("can't open file [%s] error: %w", fName, err)
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &users)
		if err != nil {
			return nil, users, fmt.Errorf("unmarshal error: %w", err)
		}
	}
	return
}

func findById(users []User, id string) ([]byte, int) {
	for i, user := range users {
		if user.Id == id {
			js, err := user.Marshal()
			if err != nil {
				continue
			}
			return js, i
		}
	}
	return []byte(""), -1
}

func addUser(users []User, user User, w io.Writer) error {
	users = append(users, user)
	b, err := json.Marshal(&users)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func removeUser(users []User, i int, w io.Writer) error {
	users[i] = users[len(users)-1]
	users = users[:len(users)-1]
	b, err := json.Marshal(&users)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func getAllUsers(users []User, w io.Writer) error {
	b, err := json.Marshal(&users)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func Perform(args Arguments, writer io.Writer) error {
	fName := args["fileName"]
	if len(fName) == 0 {
		return errors.New("-fileName flag has to be specified")
	}

	op := args["operation"]
	if len(op) == 0 {
		return errors.New("-operation flag has to be specified")
	}

	switch op {
	case "add":
		item := args["item"]
		if len(item) == 0 {
			return errors.New("-item flag has to be specified")
		}
		user := User{}
		user.Set(item)
		f, users, err := fileOpen(fName, "findById")
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		f.Close()
		_, idx := findById(users, user.Id)
		if idx != -1 {
			fmt.Fprintf(writer, "Item with id %s already exists", user.Id)
			return nil
		}
		f, _, err = fileOpen(fName, op)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		defer f.Close()
		err = addUser(users, user, f)
		if err != nil {
			return fmt.Errorf("add user error: %w", err)
		}
	case "remove":
		id := args["id"]
		if len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}
		f, users, err := fileOpen(fName, "findById")
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		_, idx := findById(users, id)
		f.Close()
		if idx == -1 {
			fmt.Fprintf(writer, "Item with id %s not found", id)
			return nil
		}
		f, _, err = fileOpen(fName, op)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		defer f.Close()
		err = removeUser(users, idx, f)
		if err != nil {
			return fmt.Errorf("remove user error: %w", err)
		}
	case "list":
		f, users, err := fileOpen(fName, op)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		f.Close()
		err = getAllUsers(users, writer)
		if err != nil {
			return fmt.Errorf("list users error: %w", err)
		}
	case "findById":
		id := args["id"]
		if len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}
		f, users, err := fileOpen(fName, op)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", fName, err)
		}
		f.Close()
		user, _ := findById(users, id)
		_, err = writer.Write(user)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", op)
	}
	return nil
}


func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
