package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-autostart"
	"gopkg.in/yaml.v3"
)

const AppName = "AUTODPI"
const AppVersion = "0.1"
const DPIFilePathname string = "./dpi.yaml"
const resolutionCheck = time.Second * 3
const DefaultDPI_YAML = `# 96 font DPI is default on XFCE
# -1 means "no custom font DPI"

# do not add any spaces at left or right of
# the resolution

resolutions:
  - res: 800x600
    dpi: -1
  - res: 1024x768
    dpi: -1
  - res: 1152x864
    dpi: -1
  - res: 1280x720
    dpi: -1
  - res: 1280x800
    dpi: -1
  - res: 1280x1024
    dpi: -1
  - res: 1366x664
    dpi: -1
  - res: 1360x768
    dpi: -1
  - res: 1366x768
    dpi: -1
  - res: 1600x900
    dpi: -1
  - res: 1600x1200
    dpi: -1
  - res: 1680x1050
    dpi: -1
  - res: 1920x1080
    dpi: -1
  - res: 1920x1200
    dpi: -1
  - res: 2048x1536
    dpi: -1
  - res: 3200x1800
    dpi: 148
  - res: 3840x1620
    dpi: 160
  - res: 3840x2160
    dpi: 160
`

var fullDPIFilePathname, _ = filepath.Abs(DPIFilePathname)
var dpiData = make(map[string]any)

func duplicateLog() {
	logFilename := filepath.Base(os.Args[0]) + ".txt"
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
}

func getFullAppName() string {
	return fmt.Sprintf("%v v%v", AppName, AppVersion)
}

func printAppName() {
	log.Println(
		getFullAppName())
	log.Println()
	log.Println("Automatic font DPI changer for the XFCE.")
	log.Println()
}

func printAppInfo() {
	log.Println("This app will automatically change font DPI for your")
	log.Println("current screen resolution, for the Xfce Desktop Environment.")
	log.Println()
	log.Printf("Font DPI will be read from %v file.\n", DPIFilePathname)
	log.Println()
	log.Println("It reqires xrandr, xfconf-query and grep")
	log.Println("commands.")
	log.Println()
	log.Println("This app can be run only on Linux.")
	log.Println()
	log.Println()
}

func printUsages() {
	log.Printf("Usage: %v <option>", os.Args[0])

	log.Println()
	log.Println("Options:")

	log.Println("\t--install")
	log.Println("\t\t\t autorun with the system")
	log.Println()
	log.Println("\t--uninstall")
	log.Println("\t\t\t do not autorun with the system")
	log.Println()
	log.Println("\t--run")
	log.Println("\t\t\t just run")
	log.Println()
	log.Println("\t--status")
	log.Println("\t\t\t check if app (autorun) is installed")
}

func shouldPrintUsages() bool {
	len_args := len(os.Args)

	return len_args != 2 || (len_args > 1 && os.Args[1] == "--help")
}

func getGoAutostartApp() (*autostart.App, error) {
	executable, err := os.Executable()

	if err != nil {
		return nil, err
	}

	fullAppName := getFullAppName()
	app := autostart.App{
		Name:        fullAppName,
		DisplayName: fullAppName,
		Exec:        []string{executable, "--run"},
	}

	return &app, nil
}

func checkInstalled() {
	app, err := getGoAutostartApp()

	if err != nil {
		log.Fatal(err)
	}

	if app.IsEnabled() {
		log.Fatal("App autorun is installed.")
	} else {
		log.Fatal("App autorun is not installed.")
	}
}

func printAppStatus() {
	checkInstalled()
}

func installAutorun() {
	app, err := getGoAutostartApp()

	if err != nil {
		log.Fatal(err)
	}

	if app.IsEnabled() {
		log.Fatal("App already installed.")
	}

	err = app.Enable()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("App installed.")
}

func uninstallAutorun() {
	app, err := getGoAutostartApp()

	if err != nil {
		log.Fatal(err)
	}

	if !app.IsEnabled() {
		log.Fatal("App is not installed.")
	}

	err = app.Disable()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("App uninstalled.")
}

func checkPlatform() {
	if runtime.GOOS != "linux" {
		log.Fatalln("This app can be used only on Linux.")
	}
}

func createDpiFile() {
	if _, err := os.Stat(fullDPIFilePathname); err == nil {
		return
	}

	log.Printf("%v does not exists, creating default.", fullDPIFilePathname)

	os.WriteFile(fullDPIFilePathname, []byte(DefaultDPI_YAML), 0666)

	log.Printf("%v created.\n", fullDPIFilePathname)
}

func readDpiFile() {
	formatErrMsg := fmt.Sprintf(
		"%v has incorrect format, please delete it to create default.",
		fullDPIFilePathname)

	data, err := os.ReadFile(fullDPIFilePathname)

	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(data, dpiData)

	if err != nil {
		log.Fatalln(err)
	}

	if _, hasResolutions := dpiData["resolutions"]; !hasResolutions {
		log.Fatalf(formatErrMsg)
	}

	for _, dpiData := range dpiData["resolutions"].([]any) {
		dpiDataObj := dpiData.(map[string]any)

		if _, hasDpi := dpiDataObj["dpi"]; !hasDpi {
			log.Fatalln(formatErrMsg)
		}

		if _, hasRes := dpiDataObj["res"]; !hasRes {
			log.Fatalln(formatErrMsg)
		}

		if strings.TrimSpace(dpiDataObj["res"].(string)) == "" {
			log.Fatalln(formatErrMsg)
		}

		if _, isInt := dpiDataObj["dpi"].(int); !isInt {
			log.Fatalln(formatErrMsg)
		}
	}
}

func getPrimaryScreenResolution() string {
	output, err := exec.Command("xrandr").Output()

	if err != nil {
		log.Fatalln(err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)

		// parse line like:
		// Virtual1 connected primary 1366x768+0+0 (normal left inverted right x axis y axis) 0mm x 0mm
		if strings.Contains(line, " connected primary ") {
			res := strings.Split(line, " ")[3]

			return strings.Split(res, "+")[0]
		}
	}

	return ""
}

func findDpiValue(resolution string) *int {
	for _, resData := range dpiData["resolutions"].([]any) {
		resDataMap := resData.(map[string]any)

		if resDataMap["res"] == resolution {
			dpi := resDataMap["dpi"].(int)

			return &dpi
		}
	}
	return nil
}

func loop() {
	lastResolution := ""

	for {
		time.Sleep(resolutionCheck)

		resolution := getPrimaryScreenResolution()

		if resolution == "" || resolution == lastResolution {
			continue
		}

		log.Printf("Found new resolution %v", resolution)

		dpi := findDpiValue(resolution)

		if dpi == nil {
			log.Fatalf(
				"Cannot find font DPI value for %v resolution, please add it in %v file.\n",
				resolution,
				fullDPIFilePathname)
		}

		log.Printf("Setting font DPI %v\n", *dpi)

		cmd := exec.Command(
			"xfconf-query",
			"-c",
			"xsettings",
			"-p",
			"/Xft/DPI",
			"-s",
			strconv.Itoa(*dpi))

		cmd.Start()
		cmd.Wait()

		if cmd.Err != nil {
			log.Fatalln(cmd.Err)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			output, _ := cmd.CombinedOutput()

			log.Fatalf("Cannot set font DPI (exit code %v), xfconf-query:\n%v", cmd.ProcessState.ExitCode(), output)
		}

		lastResolution = resolution
	}
}

func main() {
	duplicateLog()
	printAppName()
	checkPlatform()

	if shouldPrintUsages() {
		printAppInfo()
		printUsages()

		os.Exit(1)
	}

	if os.Args[1] == "--status" {
		printAppStatus()
	} else if os.Args[1] == "--install" {
		installAutorun()
	} else if os.Args[1] == "--uninstall" {
		uninstallAutorun()
	} else if os.Args[1] == "--run" {
		createDpiFile()
		readDpiFile()
		loop()
	} else {
		printAppInfo()
		printUsages()
	}
}
