package main

import "log"

type stressBot struct {
	name string
}

func NewStressBot() *stressBot {
	return &stressBot{"stress"}
}

func (cmd *stressBot) Name() string {
	return cmd.name
}

func (cmd *stressBot) Execute() {
	log.Printf("stress bot")
}
