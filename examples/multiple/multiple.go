package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Person struct {
	ID        int64
	Name      string
	Email     string
	Age       int
	CreatedOn time.Time
}

func (p Person) String() string {
	return fmt.Sprintf("%d: name=%s email=%s age=%d createdon=%v", p.ID, p.Name, p.Email, p.Age, p.CreatedOn)
}

type PhoneType int

const (
	Home PhoneType = iota
	Work
	Mobile
	Other
)

type Phone struct {
	ID     int64
	Number string
	Type   PhoneType
}

func (p Phone) String() string {
	return fmt.Sprintf("%d: number=%s type=%d", p.ID, p.Number, p.Type)
}

func main() {
	db, _ := sql.Open("postgres", "sslmode=disable")
	personMap := NewPersonMapper(db)
	phoneMap := NewPhoneMapper(db)
	if err := personMap.Insert(&Person{42, "Paul Smith", "paulsmith@pobox.com", 37, time.Now()}); err != nil {
		log.Fatal(err)
	}
	phones := []Phone{
		{1, "800-555-1212", Work},
		{2, "202-123-4567", Home},
		{3, "312-987-6543", Other},
		{4, "301-848-4848", Mobile},
	}
	for _, p := range phones {
		if err := phoneMap.Insert(&p); err != nil {
			log.Fatal(err)
		}
	}
	p, err := phoneMap.FindWhere(fmt.Sprintf("\"type\" = %d", Mobile))
	if err != nil {
		log.Fatal(err)
	}
	if len(p) != 1 {
		log.Fatal("want 1, got %d", len(p))
	}
	actual := phones[len(phones)-1]
	if p[0].Number != actual.Number || p[0].Type != actual.Type {
		log.Fatalf("want %v, got %v", actual, p[0])
	}
	log.Printf("%+v\n", p)
}
