package ftp_client

import (
	"io"

	"github.com/jlaffaye/ftp"
)

type FTPClient struct {
    conn *ftp.ServerConn
}

func NewFTPClient(addr string) (*FTPClient, error) {
    c, err := ftp.Dial(addr)
    if err != nil {
        return nil, err
    }
	if err := c.Login("anonymous", "anonymous"); err != nil {
		return nil, err
	}
    return &FTPClient{conn: c}, nil
}

func (c *FTPClient) RetrieveFile(path string) (io.ReadCloser, error) {
    return c.conn.Retr(path)
}

func (c *FTPClient) Close() error {
    return c.conn.Quit()
}