package core

import (
	"flag"
	"fmt"
	"opentoken-node/core/internal"
	"opentoken-node/global"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Viper() *viper.Viper {
	config := getConfigPath()

	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err = v.Unmarshal(&global.GLB_CONFIG); err != nil {
			fmt.Println(err)
		}
	})
	if err = v.Unmarshal(&global.GLB_CONFIG); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}

	// root 适配性 根据root位置去找到对应迁移位置,保证root路径有效
	global.GLB_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	return v
}

func getConfigPath() (config string) {
	// `-c` flag parse
	flag.StringVar(&config, "c", "", "choose config file.")
	flag.Parse()
	if config != "" { // 命令行参数不为空 将值赋值于config
		fmt.Printf("您正在使用命令行的 '-c' 参数传递的值, config 的路径为 %s\n", config)
		return
	}
	_, err := os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		config = internal.ConfigLocalFile
		fmt.Printf("配置文件路径不存在, 使用本地配置文件路径: %s\n", config)
	}
	_, err = os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		config = internal.ConfigDefaultFile
		fmt.Printf("配置文件路径不存在, 使用默认配置文件路径: %s\n", config)
	}
	return

}
