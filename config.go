package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type config struct {
	DBPath        string `json:"db_file_path"`
	ListenAddress string `json:"listen_address"`
	ListenPort    string `json:"listen_port"`
}

func (c *config) Init(fPath string) {
	byts, err := ioutil.ReadFile(fPath)
	if err != nil {
		fmt.Printf("read file with error:[%v],use default value!\n", err)
		c.DBPath = "database.db"
		c.ListenAddress = "[::]"
		c.ListenPort = "1325"
	}
	json.Unmarshal(byts, c)
	fmt.Println("load configuration done!")
}
