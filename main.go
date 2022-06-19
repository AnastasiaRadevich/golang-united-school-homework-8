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
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func unmarshalUsers(content []byte) (users []User, err error) {
	errUnmarshal := json.Unmarshal(content, &users)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return users, nil
}

func unmarshalUser(content []byte) (user User, err error) {
	errUnmarshal := json.Unmarshal(content, &user)
	if errUnmarshal != nil {
		return User{}, errUnmarshal
	}

	return user, nil
}

func marshalUsers(editUsers []User, file *os.File) ([]byte, error) {
	data, errMarsh := json.Marshal(editUsers)
	if errMarsh != nil {
		return nil, errMarsh
	}

	errTrunc := file.Truncate(0)
	if errTrunc != nil {
		return nil, errTrunc
	}

	_, errSeek := file.Seek(0, 0)
	if errSeek != nil {
		return nil, errSeek
	}

	_, errWrite := file.WriteString(string(data))
	if errWrite != nil {
		return nil, errWrite
	}

	return data, nil
}

func (args Arguments) getItemsList() ([]byte, error) {
	file, err := os.OpenFile(args["fileName"], os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func (args Arguments) addItemToList() ([]byte, error) {
	file, errOpen := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()

	fileContent, errFileContent := ioutil.ReadAll(file)
	if errFileContent != nil {
		return nil, errFileContent
	}

	if len(fileContent) == 0 {
		unmarshalUser, errorUnmarshalUser := unmarshalUser([]byte(args["item"]))
		if errorUnmarshalUser != nil {
			return nil, errorUnmarshalUser
		}

		users := make([]User, 1)
		users[0] = unmarshalUser

		data, errMarshal := marshalUsers(users, file)
		if errMarshal != nil {
			return nil, errMarshal
		}
		return data, nil
	}

	unmarshalUsers, errorUnmarshalUsers := unmarshalUsers(fileContent)
	if errorUnmarshalUsers != nil {
		return nil, errorUnmarshalUsers
	}

	unmarshalUser, errorUnmarshalUser := unmarshalUser([]byte(args["item"]))
	if errorUnmarshalUser != nil {
		return nil, errorUnmarshalUser
	}

	for i := 0; i < len(unmarshalUsers); i++ {
		if unmarshalUsers[i].Id == unmarshalUser.Id {
			return []byte(fmt.Sprintf("Item with id %s already exists", unmarshalUser.Id)), nil
		}
	}

	editUsers := append(unmarshalUsers, unmarshalUser)

	data, errMarshal := marshalUsers(editUsers, file)
	if errMarshal != nil {
		return nil, errMarshal
	}

	return data, nil
}

func (args Arguments) findById() ([]byte, error) {
	file, errOpen := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()

	fileContent, errFileContent := ioutil.ReadAll(file)
	if errFileContent != nil {
		return nil, errFileContent
	}

	if len(fileContent) == 0 {
		return nil, nil
	}

	unmarshalUsers, errorUnmarshalUsers := unmarshalUsers(fileContent)
	if errorUnmarshalUsers != nil {
		return nil, errorUnmarshalUsers
	}

	for i := 0; i < len(unmarshalUsers); i++ {
		if unmarshalUsers[i].Id == args["id"] {
			return json.Marshal(unmarshalUsers[i])
		}
	}

	return []byte(""), nil
}

func (args Arguments) removeFromList() ([]byte, error) {
	file, errOpen := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()

	fileContent, errFileContent := ioutil.ReadAll(file)
	if errFileContent != nil {
		return nil, errFileContent
	}

	if len(fileContent) == 0 {
		return nil, nil
	}

	unmarshalUsers, errorUnmarshalUsers := unmarshalUsers(fileContent)
	if errorUnmarshalUsers != nil {
		return nil, errorUnmarshalUsers
	}

	for i := 0; i < len(unmarshalUsers); i++ {
		if unmarshalUsers[i].Id == args["id"] {
			unmarshalUsers = append(unmarshalUsers[:i], unmarshalUsers[i+1:]...)

			data, errMarshal := marshalUsers(unmarshalUsers, file)
			if errMarshal != nil {
				return nil, errMarshal
			}

			return data, nil
		}
	}
	return []byte(fmt.Sprintf("Item with id %s not found", args["id"])), nil
}

func Perform(args Arguments, writer io.Writer) error {
	if args["operation"] == "" {
		return errors.New("-operation flag has to be specified")
	}

	if args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}

	if args["operation"] == "list" {
		list, errList := args.getItemsList()
		if errList != nil {
			return errList
		}

		_, errWrite := writer.Write(list)
		if errWrite != nil {
			return errWrite
		}
		return nil
	}

	if args["operation"] == "add" {
		if args["item"] == "" {
			return errors.New("-item flag has to be specified")
		}

		res, errAdd := args.addItemToList()
		if errAdd != nil {
			return errAdd
		}

		_, errWrite := writer.Write(res)
		if errWrite != nil {
			return errWrite
		}
		return nil
	}

	if args["operation"] == "findById" {
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}

		res, errFind := args.findById()
		if errFind != nil {
			return errFind
		}

		_, errWrite := writer.Write(res)
		if errWrite != nil {
			return errWrite
		}
		return nil
	}

	if args["operation"] == "remove" {
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}

		res, errRemove := args.removeFromList()
		if errRemove != nil {
			return errRemove
		}

		_, errWrite := writer.Write(res)
		if errWrite != nil {
			return errWrite
		}
		return nil
	}

	return fmt.Errorf("Operation %s not allowed!", args["operation"])
}

func parseArgs() Arguments {
	idFlag := flag.String("id", "", "current id")
	itemFlag := flag.String("item", "", "current item")
	operationFlag := flag.String("operation", "", "current operation")
	fileNameFlag := flag.String("fileName", "users.json", "current file name")

	flag.Parse()

	return Arguments{"operation": *operationFlag, "item": *itemFlag, "fileName": *fileNameFlag, "id": *idFlag}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
