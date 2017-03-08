package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
)

// Fastly API doesn't return sorted data
type version struct {
	Number  int
	Version *fastly.Version
}

type wrappedVersions []version

// Satisfy the Sort interface
func (v wrappedVersions) Len() int      { return len(v) }
func (v wrappedVersions) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v wrappedVersions) Less(i, j int) bool {
	return v[i].Number < v[j].Number
}

type vclResponse struct {
	Path    string
	Name    string
	Content string
}

// Globals needed for sharing between functions
var fastlyServiceID string
var latestVersion string
var selectedVersion string

// List of VCL files to process
var vclFiles []string

// Regex used to define user specific filtering
var dirSkipRegex *regexp.Regexp
var dirMatchRegex *regexp.Regexp

// WaitGroup and Channel for storing responses from API
var wg sync.WaitGroup
var ch chan vclResponse

func main() {
	help := flag.Bool("help", false, "show available flags")
	appVersion := flag.Bool("version", false, "show application version")
	debug := flag.Bool("debug", false, "show the error/diff output")
	serviceVersion := flag.String("vcl-version", "", "specify Fastly service 'version' to verify against")
	service := flag.String("service", os.Getenv("FASTLY_SERVICE_ID"), "your service id (fallback: FASTLY_SERVICE_ID)")
	token := flag.String("token", os.Getenv("FASTLY_API_TOKEN"), "your fastly api token (fallback: FASTLY_API_TOKEN)")
	dir := flag.String("dir", os.Getenv("VCL_DIRECTORY"), "vcl directory to compare files against")
	skip := flag.String("skip", "^____", "regex for skipping vcl directories (will also try: VCL_SKIP_DIRECTORY)")
	match := flag.String("match", "", "regex for matching vcl directories (will also try: VCL_MATCH_DIRECTORY)")
	flag.Parse()

	if *help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *appVersion == true {
		fmt.Println("1.0.0")
		os.Exit(1)
	}

	envSkipDir := os.Getenv("VCL_SKIP_DIRECTORY")
	envMatchDir := os.Getenv("VCL_MATCH_DIRECTORY")

	if envSkipDir != "" {
		*skip = envSkipDir
	}

	if envMatchDir != "" {
		*match = envMatchDir
	}

	dirSkipRegex, _ = regexp.Compile(*skip)
	dirMatchRegex, _ = regexp.Compile(*match)

	client, err := fastly.NewClient(*token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fastlyServiceID = *service

	listVersions, err := client.ListVersions(&fastly.ListVersionsInput{
		Service: fastlyServiceID,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wv := wrappedVersions{}
	for _, v := range listVersions {
		i, err := strconv.Atoi(v.Number)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		wv = append(wv, version{i, v})
	}
	sort.Sort(wv)

	latestVersion = strconv.Itoa(wv[len(wv)-1].Number)

	if *serviceVersion != "" {
		selectedVersion = *serviceVersion
	} else {
		selectedVersion = latestVersion
	}

	walkError := filepath.Walk(*dir, aggregate)
	if walkError != nil {
		fmt.Printf("filepath.Walk() returned an error: %v\n", walkError)
	}

	ch = make(chan vclResponse, len(vclFiles))
	for _, vclPath := range vclFiles {
		wg.Add(1)
		go getVCL(vclPath, client)
	}
	wg.Wait()
	close(ch)

	for vclFile := range ch {
		processDiff(vclFile, *debug)
	}
}

func aggregate(path string, f os.FileInfo, err error) error {
	if validPathDefaults(path) && validPathUserDefined(path) && !invalidPathUserDefined(path) {
		vclFiles = append(vclFiles, path)
	}

	return nil
}

func validPathDefaults(path string) bool {
	return !strings.Contains(path, ".git") && strings.Contains(path, ".vcl")
}

func validPathUserDefined(path string) bool {
	return dirMatchRegex.MatchString(path)
}

func invalidPathUserDefined(path string) bool {
	return dirSkipRegex.MatchString(path)
}

func getVCL(path string, client *fastly.Client) {
	defer wg.Done()

	name := extractName(path)

	vclFile, err := client.GetVCL(&fastly.GetVCLInput{
		Service: fastlyServiceID,
		Version: selectedVersion,
		Name:    name,
	})

	if err != nil {
		ch <- vclResponse{
			Path:    path,
			Name:    name,
			Content: fmt.Sprintf("error: %s", err),
		}
	} else {
		ch <- vclResponse{
			Path:    path,
			Name:    name,
			Content: vclFile.Content,
		}
	}
}

func extractName(path string) string {
	_, file := filepath.Split(path)
	return strings.Split(file, ".")[0]
}

func processDiff(vr vclResponse, debug bool) {
	var (
		err    error
		cmdOut []byte
	)
	cmdName := "diff"
	cmdArgs := []string{"-w", "-B", "-I", "[[:space:]]\\+#", "-", vr.Path}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = strings.NewReader(vr.Content)

	if cmdOut, err = cmd.Output(); err != nil {
		color.Red("\nThere was a difference between the version (%s) of '%s' and the version found locally\n\t%s\n", selectedVersion, vr.Name, vr.Path)

		if debug == true {
			fmt.Printf("\n%s\n", string(cmdOut))
		}
	} else {
		color.Green("\nNo difference between the version (%s) of '%s' and the version found locally\n\t%s\n", selectedVersion, vr.Name, vr.Path)
	}
}
