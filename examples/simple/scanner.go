package main

type Scanner interface {
	Scan(...interface{}) error
}
