package main

import "os/user"

func home() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.HomeDir
}
