package main

import (
	"github.com/jlaffaye/ftp"
)

func FTPConnect(server string) (*ftp.ServerConn, error) {
	client, err := ftp.Dial(server)
	if err != nil {
		return nil, err
	}
	if err := client.Login("anonymous", "anonymous"); err != nil {
		return nil, err
	}
	return client, nil
}
