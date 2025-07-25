package test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZEL-30/zel/cmake"
	"github.com/ZEL-30/zel/cmd/commands"
	"github.com/ZEL-30/zel/cmd/commands/version"
	"github.com/ZEL-30/zel/config"
	"github.com/ZEL-30/zel/logger"
	"github.com/ZEL-30/zel/logger/colors"
	"github.com/ZEL-30/zel/utils"
)

var CmdTest = &commands.Command{
	UsageLine: "test [appname] [watchall] [-main=*.go] [-downdoc=true]  [-gendoc=true] [-vendor=true] [-e=folderToExclude] [-ex=extraPackageToWatch] [-tags=goBuildTags] [-runmode=BEEGO_RUNMODE]",
	Short:     "Test the application by starting a local development server",
	Long: `
Run command will supervise the filesystem of the application for any changes, and recompile/restart it.
	`,
	PreRun: func(cmd *commands.Command, args []string) {},
	Run:    RunTest,
}

var (
	rebuild   bool // 是否重新构建
	appPath   string
	buildPath string
	testPath  string
	testInfos []string
)

func init() {
	CmdTest.Flag.BoolVar(&rebuild, "r", false, "Clear the build folder in the project and rebuild, default false")
	commands.AvailableCommands = append(commands.AvailableCommands, CmdTest)
}

func RunTest(cmd *commands.Command, args []string) int {

	appPath := utils.GetZelWorkPath()
	buildPath = filepath.Join(appPath, "build")
	testPath = filepath.Join(appPath, "bin", "test")

	if len(args) == 0 {
		showTest()
	} else {
		if len(args) > 2 {
			err := cmd.Flag.Parse(args[1:])
			if err != nil {
				logger.Log.Fatal("Parse args err" + err.Error())
			}
		}
		runTest(args[0])
	}

	return 0
}

func showTest() {

	version.ShowShortVersionBanner()
	fmt.Println()

	// 设置临时环境变量
	dllPath := getDllPath()
	restore, err := utils.SetEnvTemp("PATH", dllPath)
	if err != nil {
		logger.Log.Errorf("Failed to set PATH environment variable: %v", err)
		return
	}
	defer restore() // 确保在函数结束时恢复原始 PATH

	filepath.Walk(testPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if index := strings.Index(path, "-test.exe"); index != -1 {
			cmd := exec.Command(path, "--gtest_list_tests")
			bytes, err := cmd.Output()
			if err != nil {
				logger.Log.Fatal(err.Error())
			}

			testInfos = strings.Split(string(bytes), "\n")
			for _, testInfo := range testInfos {
				switch {

				case strings.HasPrefix(testInfo, "Running"):

				case strings.Index(testInfo, ".") != -1:
					testInfo = colors.RedBold(testInfo[:len(testInfo)-2])
					fmt.Println(`    ├── ` + testInfo)

				case strings.HasPrefix(testInfo, "  "):
					fmt.Println(`    │    └── ` + testInfo[2:])

				}
			}

		}
		return nil
	})

	fmt.Println()

}

func runTest(testName string) {

	var (
		testProgram string
		testExe     string
	)

	if index := strings.Index(testName, "."); index == -1 {

		testProgram = getTestProgramName(testName) + "-test.exe"
		testName += "*"
	} else {
		testProgram = getTestProgramName(testName[:index]) + "-test.exe"
	}

	configArg := cmake.ConfigArg{
		Toolchain:             config.Conf.Toolchain,
		Platform:              config.Conf.Platform,
		BuildType:             config.Conf.BuildType,
		Generator:             config.Conf.Generator,
		NoWarnUnusedCli:       true,
		ExportCompileCommands: true,
		ProjectPath:           appPath,
		BuildPath:             buildPath,
		CXXFlags:              "-D_MD",
	}

	buildArg := cmake.BuildArg{
		BuildPath: buildPath,
	}

	// testName := cases.Title(language.English).String(testName)
	err := cmake.Build(&configArg, &buildArg, rebuild, false)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	// 设置临时环境变量
	dllPath := getDllPath()
	logger.Log.Infof("Setting PATH environment variable to: %s", dllPath)

	restore, err := utils.SetEnvTemp("PATH", dllPath)
	if err != nil {
		logger.Log.Errorf("Failed to set PATH environment variable: %v", err)
		return
	}
	defer restore() // 确保在函数结束时恢复原始 PATH

	testExe = filepath.Join(testPath, testProgram)

	arg := fmt.Sprintf("--gtest_filter=%s", testName)
	c := exec.Command(testExe, arg)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

}

func getTestProgramName(testName string) string {
	var result []byte

	for i, letter := range testName {
		if letter >= 65 && letter <= 90 {
			if i == 0 {
				result = append(result, byte(letter+32))
				continue
			}
			result = append(result, '-')
			result = append(result, byte(letter+32))
		} else {
			result = append(result, byte(letter))
		}
	}
	return string(result)
}

func getDllPath() string {
	var dllPath string
	zelHome := utils.GetZelHomePath()
	switch config.Conf.Platform {
	case "x86":
		dllPath = filepath.Join(zelHome, "installed", "x86-windows")
	case "x64":
		dllPath = filepath.Join(zelHome, "installed", "x64-windows")
	}

	switch config.Conf.BuildType {
	case "Debug":
		dllPath = filepath.Join(dllPath, "debug", "bin")
	case "Release":
		dllPath = filepath.Join(dllPath, "bin")
	}
	return dllPath
}
