package cmds

import (
	"fmt"
	"os"

	"github.com/pattonjp/xbox-sync/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	confFileName = ".xbox-sync"
)

var (
	cfgFile       string
	validProfiles = []string{
		"xbox",
		"x3",
	}
	rootCmd = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.Unmarshal(&config)
		},
	}
	config client.Config
)

func bindFlag(name string, def interface{}) {
	viper.BindPFlag(name, rootCmd.PersistentFlags().Lookup(name))
	viper.SetDefault(name, def)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(list())
	rootCmd.AddCommand(addGame())
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/%s)", confFileName))
	rootCmd.PersistentFlags().StringP("server", "s", "", "xbox host server")
	rootCmd.PersistentFlags().Int("port", 21, "xbox server port")
	rootCmd.PersistentFlags().String("user", "", "ftp user")
	rootCmd.PersistentFlags().String("pwd", "", "ftp password")
	rootCmd.PersistentFlags().String("remoteDir", "", "default xbox directory")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "show debug information")
	rootCmd.PersistentFlags().String("localDir", "", "local path to game")

	bindFlag("server", nil)
	bindFlag("user", "xbox")
	bindFlag("pwd", "xbox")
	bindFlag("debug", false)
	bindFlag("port", 21)
	bindFlag("remoteDir", "/g/games")
	bindFlag("localDir", nil)
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(confFileName)
		viper.SafeWriteConfig()
	}
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}

}

func Run() error {
	return rootCmd.Execute()
}

func list() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "list dir",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				config.RemoteGamesDir = args[0]
			}
			cli, err := client.New(config)
			if err != nil {
				return err
			}
			defer cli.Quit()
			entries, err := cli.List("/")
			if err != nil {
				return err
			}
			for _, ent := range entries {
				fmt.Println(ent.Name)
			}
			return nil
		},
	}

	return cmd
}

func addGame() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-game",
		Short: "adds a game",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := client.New(config)
			if err != nil {
				return err
			}
			defer cli.Quit()
			cli.AddGame(config.LocalGamesDir, config.RemoteGamesDir)
			return nil
		},
	}

	return cmd
}
