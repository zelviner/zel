package env

import (
	"io"
	"os"
	"text/template"
	"time"

	"github.com/ZEL-30/zel/config"
	"github.com/ZEL-30/zel/logger"
)

type EnvInfo struct {
	ZelVersion string
	ZelPath    string
	BuildMode  string
	BuildKit   string
	TestMode   string
}

// InitBanner 加载横幅并打印到输出
// 所有错误都被忽略，应用程序不会在错误的情况下打印横幅
func InitBanner(out io.Writer, in io.Reader) {
	if in == nil {
		logger.Log.Fatal("The input is nil")
	}

	banner, err := io.ReadAll(in)
	if err != nil {
		logger.Log.Fatalf("Error while trying to read the banner: %s", err)
	}

	show(out, string(banner))

}

func show(out io.Writer, content string) {
	t, err := template.New("banner").Funcs(template.FuncMap{"Now": Now}).Parse(content)
	if err != nil {
		logger.Log.Fatalf("Cannot parse the banner template: %s", err)
	}

	envInfo := EnvInfo{
		ZelVersion: config.Version,
		BuildMode:  config.Conf.BuildMode,
		TestMode:   config.Conf.TestMode,
	}

	if config.Conf.Kit == nil {
		envInfo.BuildKit = "N/A"
	} else {
		envInfo.BuildKit = config.Conf.Kit.Name
	}

	if cPath := os.Getenv("ZELPATH"); cPath != "" {
		envInfo.ZelPath = cPath
	} else {
		envInfo.ZelPath = config.Conf.ZelPath
	}

	err = t.Execute(out, envInfo)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

// Now 返回指定布局中的当前本地时间
func Now(layout string) string {
	return time.Now().Format(layout)
}
