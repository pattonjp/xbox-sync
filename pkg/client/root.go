package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/jlaffaye/ftp"
)

type Client struct {
	conf Config
	*ftp.ServerConn
}

type Config struct {
	Server         string `mapstructure:"server"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Pwd            string `mapstructure:"pwd"`
	Debug          bool   `mapstructure:"debug"`
	LocalGamesDir  string `mapstructure:"localDir"`
	RemoteGamesDir string `mapstructure:"remoteDir"`
}

func New(conf Config) (*Client, error) {
	cli := &Client{conf: conf}
	return cli, cli.connect()
}

func (cli *Client) EnsureDir(pth string) error {
	dirs := strings.Split(pth, "/")
	base := strings.Join(dirs[0:len(dirs)-1], "/")
	cli.ChangeDir(base)
	err := cli.MakeDir(pth)
	if !strings.Contains(err.Error(), "250") {
		return err
	}
	return nil
}

func (cli *Client) EnsureFile(path, dest string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	cli.ServerConn.Stor(dest, f)
	f.Close()
	return nil
}

func (cli *Client) connect() error {
	var opts []ftp.DialOption
	if cli.conf.Debug {
		opts = append(opts, ftp.DialWithDebugOutput(os.Stdout))
	}
	client, err := ftp.Dial(fmt.Sprintf("%s:%d", cli.conf.Server, cli.conf.Port), opts...)

	if err != nil {
		return err
	}
	if err := client.Login(cli.conf.User, cli.conf.Pwd); err != nil {
		return err
	}
	cli.ServerConn = client
	if cli.conf.RemoteGamesDir == "" {
		cli.conf.RemoteGamesDir = "/"
	}
	if err := cli.ServerConn.ChangeDir(cli.conf.RemoteGamesDir); err != nil {
		return err
	}

	return nil

}
