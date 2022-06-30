package model

import (
	"fmt"
	"time"
)

type ErrNotFound struct {
	When time.Time
	What string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("at %v, %s",
		e.When, e.What)
}
