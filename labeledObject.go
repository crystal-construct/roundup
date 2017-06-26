package main

type LabeledObject struct {
	Name   string            `json: "name"`
	Labels map[string]string `json: "labels"`
}
