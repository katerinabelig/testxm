package main

import (
	"flag"
	"fmt"
	"log"
)

func seedUser(store Storage, username, pw, role string) *UserAccount {
	acc, err := NewUserAccount(username, pw, role)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateUser(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account => ", acc.Role)

	return acc
}

func seedUsers(s Storage) {
	seedUser(s, "dinesh", "password", "admin")
	seedUser(s, "gilfoyl", "password", "user")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedUsers(store)
	}

	server := NewAPIServer(":3000", store)
	server.Run()
}
