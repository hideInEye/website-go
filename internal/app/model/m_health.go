package model

import "context"

type IHealth interface {
	Ping(context.Context) error
}
