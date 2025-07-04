package run

import (
	"path/filepath"

	"github.com/ZEL-30/zel/cmake"
	"github.com/ZEL-30/zel/cmd/commands"
	"github.com/ZEL-30/zel/config"
	"github.com/ZEL-30/zel/utils"
)

var CmdRun = &commands.Command{
	UsageLine: "run [appname]",
	Short:     "Run the application",
	Long: `
Run command will supervise the filesystem of the application for any changes, and recompile/restart it.

`,
	PreRun: nil,
	Run:    RunApp,
}

var (
	appName string // 应用程序名称
	rebuild bool   // 是否重建
)

func init() {
	CmdRun.Flag.BoolVar(&rebuild, "r", false, "Clear the build folder in the project and rebuild, default false")
	commands.AvailableCommands = append(commands.AvailableCommands, CmdRun)
}

// RunApp定位要监视的文件，并启动 C++ 应用程序
func RunApp(cmd *commands.Command, args []string) int {

	projectPath := utils.GetZelWorkPath()
	appName = filepath.Base(projectPath)

	buildPath := filepath.Join(projectPath, "build")

	configArg := cmake.ConfigArg{
		Toolchain:             config.Conf.Toolchain,
		Platform:              config.Conf.Platform,
		BuildType:             config.Conf.BuildType,
		Generator:             config.Conf.Generator,
		NoWarnUnusedCli:       true,
		ExportCompileCommands: true,
		ProjectPath:           projectPath,
		BuildPath:             buildPath,
		CXXFlags:              "-D_MD",
	}

	buildArg := cmake.BuildArg{
		BuildPath: buildPath,
	}

	cmake.Run(&configArg, &buildArg, appName, rebuild)

	return 0
}
