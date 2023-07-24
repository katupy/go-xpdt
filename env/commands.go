package env

type Commands interface {
	Add() error
	Set() error
	Del() error
}
