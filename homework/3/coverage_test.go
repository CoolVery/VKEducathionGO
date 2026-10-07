package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)



type Usertest struct {
	ID     int
	Name   string
	Age    int
	About  string
	Gender string
}

type TestCase struct {
	ID      string
	Result  []Usertest
	IsError bool
}

func TestCartCheckout(t *testing.T) {
	// cases := []TestCase{
    // 	{
    //     	ID: "1",
    //     	Result: []Usertest{
    //         	{
    //             	ID:     0,
    //             	Name:   "BoydWolf",
    //             	Age:    22,
    //             	About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.",
    //             	Gender: "male",
    //         	},
    //     	},
    //     	IsError: false,
    // 	},
	// }

	// ts := httptest.NewServer(http.HandleFunc())
}
