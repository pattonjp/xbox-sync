package client

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
)

type Game struct {
	Name      string
	Path      string
	Size      int64
	DirCount  int
	FileCount int
}

func (g *Game) toRemotePath(local, remote string) string {
	base := fmt.Sprintf("%s/%s", remote, g.Name)
	return strings.Replace(local, g.Path, base, -1)
}

func (cli *Client) AddGame(local, remote string) error {
	game, err := selectGame(local)
	if err != nil {
		return err
	}
	dest := filepath.Join(remote, game.Name)
	cli.ChangeDir(remote)
	if err := cli.EnsureDir(dest); err != nil {
		panic(err)
	}
	desc := fmt.Sprintf("uploading %s", game.Name)
	bar := progressbar.Default(int64(game.FileCount), desc)
	filepath.WalkDir(game.Path, func(path string, d fs.DirEntry, err error) error {
		dest := game.toRemotePath(path, remote)
		if d.IsDir() {
			if err := cli.EnsureDir(dest); err != nil {
				panic(err)
			}
		} else {
			if err := cli.EnsureFile(path, dest); err != nil {
				return err
			}
			bar.Add(1)
		}
		return nil
	})
	bar.Close()
	return nil
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func selectGame(gameRoot string) (*Game, error) {
	entries, err := os.ReadDir(gameRoot)
	if err != nil {
		return nil, err
	}
	var games []Game
	for _, ent := range entries {
		if ent.IsDir() {
			game := Game{
				Name: ent.Name(),
				Path: filepath.Join(gameRoot, ent.Name()),
			}

			err := filepath.Walk(game.Path, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					game.DirCount += 1
				} else {
					game.FileCount += 1
					game.Size += info.Size()
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			games = append(games, game)
		}

	}
	fmap := promptui.FuncMap
	fmap["sized"] = prettyByteSize

	prompt := promptui.Select{
		Label: "Select Game Folder",
		Items: games,
		Size:  10,
		Searcher: func(input string, index int) bool {
			game := games[index]
			name := strings.Replace(strings.ToLower(game.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		},
		Templates: &promptui.SelectTemplates{
			FuncMap:  fmap,
			Label:    "{{.}}",
			Active:   "\U0001F3AE {{ .Name | cyan}} ",
			Inactive: "{{.Name}}",
			Selected: "\U0001F3AE {{ .Name | cyan}}",
			Details: `
			---------- Game --------------
			{{ "Name:" | faint}} {{.Name }}
			{{ "Path:" | faint}} {{.Path }}
			{{ "Total Directories" | faint }} {{.DirCount}}
			{{ "Total Files" | faint }} {{.FileCount}}
			{{ "Size" | faint}} {{ .Size | sized }}
			`,
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return &games[i], nil

}
