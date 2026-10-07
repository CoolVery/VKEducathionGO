package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

type SearchResponse struct {
	Limit      *int
	Offset     *int    // Можно учесть после сортировки
	Query      *string // подстрока в 1 из полей
	OrderField *string
	//  1 по возрастанию, 0 как встретилось, -1 по убыванию
	OrderBy *int
}

type User struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type Users struct {
	XMLName xml.Name `xml:"root"`
	List []User `xml:"row"`
}

var AccessOrderField = map[string]string {
	"Id": "Id",
	"Age": "Age",
	"Name": "Name",
}

var AccessOrderBy = map[int]int {
	-1: -1,
	0: 0,
	1: 1,
}

func ReadWithFilterXml(queryResponse string) ([]User, error) {
	userList := make([]User, 0)
	
	db, err := os.Open("dataset.xml")
	defer db.Close()
	if err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(db)
	for {
		tok, err := dec.Token()
		if err != io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "row" {
				var user User
				if err := dec.DecodeElement(&user, &t); err != nil {
					return nil, err
				}
				fullName := user.FirstName + user.LastName
				if strings.Contains(fullName, queryResponse) || strings.Contains(user.About, queryResponse) {
					userList = append(userList, user)
				}
			}
		}
	}
	return userList, nil
}

func ReadAllXml() ([]User, error) {	
	db, err := os.Open("dataset.xml")
	defer db.Close()
	if err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(db)
	var users Users
	if err := dec.Decode(&users); err != nil {
		return nil, err
	}

	return users.List, nil
}


func GetInDb(sr SearchResponse) ([]User, error) {
	var resultUser []User
	var err error
	if sr.Query == nil {
		resultUser, err = ReadAllXml()
	} else {
		resultUser, err = ReadWithFilterXml(*sr.Query)
	}
	if err != nil {
  		return nil, err
	}
	if sr.OrderField != nil {
		if sr.OrderBy != nil && *sr.OrderBy != 0 {
			switch *sr.OrderField {
			case "Id":
				slices.SortFunc(resultUser, func(a, b User) int {
					switch {
					case a.ID > b.ID:
						return 1 * *sr.OrderBy
					case a.ID < b.ID:
						return -1 * *sr.OrderBy
					default:
                		return 0
            		}
				})
			case "Age":
				slices.SortFunc(resultUser, func(a, b User) int {
					switch {
					case a.Age > b.Age:
						return 1 * *sr.OrderBy
					case a.Age < b.Age:
						return -1 * *sr.OrderBy
					default:
                		return 0
            		}
				})
			case "Name":
				slices.SortFunc(resultUser, func(a, b User) int {
					fullNameA := a.FirstName + a.LastName
					fullNameB := b.FirstName + b.LastName
					switch {
					case fullNameA > fullNameB:
						return 1 * *sr.OrderBy
					case fullNameA < fullNameB:
						return -1 * *sr.OrderBy
					default:
                		return 0
            		}
				})
			}
		}
	} else {
		if sr.OrderBy != nil && *sr.OrderBy != 0 { 
			slices.SortFunc(resultUser, func(a, b User) int {
				fullNameA := a.FirstName + a.LastName
				fullNameB := b.FirstName + b.LastName
				switch {
				case fullNameA > fullNameB:
					return 1 * *sr.OrderBy					
				case fullNameA < fullNameB:
					return -1 * *sr.OrderBy
				default:
            		return 0
            	}
			})
		}
	}
	if sr.Offset != nil {
		resultUser = resultUser[*sr.Offset:]		
	}
	if sr.Limit != nil {
		resultUser = resultUser[:*sr.Limit]
	}
	return resultUser, nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Тут писать SearchServer
	q := r.URL.Query()
	currentSearchResponse := SearchResponse{}

	if limit := q.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		currentSearchResponse.Limit = &l
	} 

	if offset := q.Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		currentSearchResponse.Offset = &o 
	} 

	if queryVal := q.Get("query"); queryVal != "" {
		currentSearchResponse.Query = &queryVal
	} 

	if orderField := q.Get("order_field"); orderField != "" {
		if of, ok := AccessOrderField[orderField]; ok {
			currentSearchResponse.OrderField = &of
		} else {
			http.Error(w, "OrderField invalid", http.StatusBadRequest)
			return
		}
	} else {
		val := AccessOrderField["Name"]
		currentSearchResponse.OrderField = &val
	}

	if orderBy := q.Get("order_by"); orderBy != "" {
		ob, err := strconv.Atoi(orderBy)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if _, ok := AccessOrderBy[ob]; !ok {
			http.Error(w, "OrderBy invalid", http.StatusBadRequest)
			return
		} else {
			currentSearchResponse.OrderBy = &ob
		}
	} else {
		val := AccessOrderBy[0]
		currentSearchResponse.OrderBy = &val
	}
	resultUsers, err := GetInDb(currentSearchResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resultUsers); err != nil {
    	fmt.Printf("encode response: %v", err)
	}
}