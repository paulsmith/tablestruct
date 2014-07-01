package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Person struct {
	ID    int64
	Name  string
	Email string
	Age   int
}

func (p Person) String() string {
	return fmt.Sprintf("%d: name=%s email=%s age=%d", p.ID, p.Name, p.Email, p.Age)
}

type Phone struct {
	ID     int64
	Number string
	Type   int
}

func (p Phone) String() string {
	return fmt.Sprintf("%d: number=%s type=%d", p.ID, p.Number, p.Type)
}

func main() {
	db, _ := sql.Open("postgres", "sslmode=disable")

	m := PersonMapper{db}

	p := &Person{
		ID:    42,
		Name:  "Paul Smith",
		Email: "paulsmith@pobox.com",
		Age:   37,
	}

	if err := m.Insert(p); err != nil {
		log.Fatal(err)
	}

	p2, err := m.Find(42)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)

	p2.Name = "Brian Eno"
	p2.Email = "brian@eno.net"
	p2.Age = 66
	if err := m.Update(p2); err != nil {
		fmt.Println(p2)
	}

	p3, err := m.Find(42)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(p3)

	ps, err := m.FindWhere("age > 30")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ps)
}
