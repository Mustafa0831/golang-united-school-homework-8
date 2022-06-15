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

type Arguments map[string]string

type User struct {
	ID    string `json:"id"`
	Email string `json :"email"`
	Age   int    `json:"age"`
}

func (u *User) String() string {
	if len(u.ID) == 0 {
		return ""
	}

	data, _ := json.Marshal(&u)

	return string(data)
}

func (u *User) Set(s string) error {
	return json.Unmarshal([]byte(s), u)
}

func parseArgs() Arguments {
	var user User
	var operation, fileName, id string

	flag.StringVar(&operation, "operation", "", `-operation "add"`)
	flag.Var(&user, "item", `-item {id: "1", email: «test@test.com», age: 31}`)
	flag.StringVar(&fileName, "fileName", "", `-fileName "users.json"`)
	flag.StringVar(&id, "id", "", `-id "2"`)

	flag.Parse()

	return Arguments{
		"operation": operation,
		"item":      user.String(),
		"fileName":  fileName,
		"id":        id,
	}
}

func Perform(args Arguments, writer io.Writer) error {
	f := args["fileName"]
	if f == "" {
		return errors.New("-fileName flag has to be specified")
	}

	o := args["operation"]
	if o == "" {
		return errors.New("-operation flag has to be specified")
	}

	switch o {
	case "add":
		i := args["item"]

		if i == "" {
			return errors.New("-item flag has to be specified")
		}

		u := User{}
		u.Set(i)

		file, users, err := fileOpen(f, "findById")
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}
		defer file.Close()

		_, index := findById(users, u.ID)
		if index == -1 {
			fmt.Fprintf(writer, "Item with id %s already exists", u.ID)
			return nil
		}

		file, _, err = fileOpen(f, o)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}

		err = addUser(users, u, file)
		if err != nil {
			return fmt.Errorf("add user error: %w", err)
		}

	case "list":
		file, us, err := fileOpen(f, o)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}
		defer file.Close()

		err = getAllUsers(us, writer)
		if err !=nil{
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}
	case "findById":
		i := args["id"]
		if len(i) == 0 {
			return errors.New("-id flag has to be specified")
		}

		file, us, err := fileOpen(f, o)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}
		defer file.Close()

		u, _ := findById(us, i)
		_, err = writer.Write(u)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	case "remove":
		i := args["id"]
		if len(i) == 0 {
			return errors.New("-id flag has to be specified")
		}

		file, us, err := fileOpen(f, "findById")
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}

		_, index := findById(us, i)
		defer file.Close()
		if index == -1 {
			fmt.Fprintf(writer, "Item with id %s not found", i)
			return nil
		}

		file, _, err = fileOpen(f, o)
		if err != nil {
			return fmt.Errorf("can't open file [%s], error: %w", f, err)
		}
		err = removeUser(us, index, file)
		if err != nil {
			return fmt.Errorf("remove user error: %w", err)
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", o)
	}
	return nil
}

func getAllUsers(us []User, wr io.Writer) error {
	data, err := json.Marshal(&us)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = wr.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func removeUser(us []User, index int, file io.Writer) error {
	us[index] = us[len(us)-1]
	us = us[:len(us)-1]
	data, err := json.Marshal(&us)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func findById(users []User, id string) ([]byte, int) {
	for i, user := range users {
		if user.ID == id {
			jsn, err := json.Marshal(&user)
			if err != nil {
				continue
			}
			return jsn, i
		}
	}
	return []byte(""), -1
}

func addUser(us []User, u User, w io.Writer) error {
	us = append(us, u)
	data, err := json.Marshal(&us)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func fileOpen(fileName string, operation string) (f *os.File, users []User, err error) {
	switch operation {
	case "add", "remove":
		if f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			return nil, users, fmt.Errorf("can't open file [%s], error: %w", fileName, err)
		}

	case "findById", "list":
		if f, err = os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
			return nil, users, fmt.Errorf("can't open file [%s], error: %w", fileName, err)
		}
	default:
		return nil, users, fmt.Errorf("Operation %s not allowed!", operation)
	}

	body, err := ioutil.ReadAll(f)
	if err != nil && err != io.EOF {
		return nil, users, fmt.Errorf("can't open file [%s], error: %w", fileName, err)
	}

	if len(body) > 0 {
		if err = json.Unmarshal(body, &users); err != nil {
			return nil, users, fmt.Errorf("unmarshal error: %w", err)
		}
	}
	return
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
