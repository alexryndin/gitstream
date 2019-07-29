package main

import (
	"fmt"
	"time"
)

type Repo struct {
	Name string
}

type Commit struct {
	Author CommitAuthor
}

type CommitAuthor struct {
	Name string
	Time time.Time
}

func printSlice(sl interface{}) {
	sls := sl.([]interface{})
	for _, v := range sls {
		fmt.Printf("%v\n", v)
	}
}
