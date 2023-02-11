package utils

import (
	"os"
	"time"
)

type Role string

const (
	GUEST string = "GUEST"
	USER         = "USER"
	ADMIN        = "ADMIN"
)

type FileStats struct {
	Name string
	Path string
	Size int64
	Mode os.FileMode
	ModTime time.Time
	IsDir bool
}

type User struct {
	Nickname      string     `validate:"required"`
	Password       string    `validate:"required"`
	Role       Role    `validate:"required"`
}