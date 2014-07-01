package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Guestbook struct {
	ID       int64
	Name     string
	Message  sql.NullString
	SignedOn time.Time
}

func (g Guestbook) String() string {
	return fmt.Sprintf("%d: name=%s message=%s (valid? %v) signedon=%v", g.ID, g.Name, g.Message.String, g.Message.Valid, g.SignedOn)
}

func main() {
	db, _ := sql.Open("postgres", "sslmode=disable")

	m := GuestbookMapper{db}

	for _, gb := range []*Guestbook{
		{42, "Paul Smith", sql.NullString{"Nice place!", true}, time.Now()},
		{43, "Brian Eno", sql.NullString{"", false}, time.Now()},
		{44, "Leslie Lamport", sql.NullString{"foo", true}, time.Now()},
		{45, "Hamlet", sql.NullString{"", false}, time.Now()},
	} {
		if err := m.Insert(gb); err != nil {
			log.Fatal(err)
		}
	}

	gb, err := m.FindWhere("message is null")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gb)
}
