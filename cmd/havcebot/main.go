package main

import (
	"context"
	"fmt"
	"time"

	"github.com/havce/havcebot/ctftime"
)

func main() {
	c := ctftime.NewClient()

	now := time.Now()
	finish := time.Now().Add(2 * 24 * 7 * time.Hour)
	events, err := c.FindEvents(context.TODO(), ctftime.EventFilter{
		Start:  &now,
		Finish: &finish,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(events)
}
