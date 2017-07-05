package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
)

type Plugin struct {
	Host       string
	Login      string
	Password   string
	Key        string
	Name       string
	Version    string
	Sources    string
	Inclusions string
	Exclusions string
	Language   string
	Profile    string
	Encoding   string
	LcovPath   string
	Debug      bool

	Path        string
	Repo        string
	Branch      string
	BranchOut   string
	Default     string // default master branch
	BranchRegex string // to check against Branch and see if we allow running it
}

func (p *Plugin) Exec() error {

	err, allowed := p.branchAllowed()
	if err != nil {
		return err
	}

	// terminate gracefully the process but do not execute sonar plugin
	if allowed == false {
		os.Exit(0)
	}

	err = p.buildRunnerProperties()
	if err != nil {
		return err
	}

	err = p.execSonarRunner()
	if err != nil {
		logrus.Println(err)
		return err
	}

	p.writePipelineLetter()

	return nil
}

func (p Plugin) branchAllowed() (error, bool) {

	fmt.Printf("Branch %s allowed - executing sonar analysis.\n", p.Branch)
	p.BranchOut = p.Branch; // save Branch name without / (when release)
	return nil, true
}

func (p Plugin) buildRunnerProperties() error {

	p.Key = strings.Replace(p.Key, "/", ":", -1)

	tmpl, err := template.ParseFiles("/opt/sonar/conf/sonar-runner.properties.tmpl")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("/opt/sonar/conf/sonar-runner.properties")
	defer f.Close()
	if err != nil {
		fmt.Println("Error creating file: ", err)
		panic(err)
	}

	if p.Debug {
		err = tmpl.ExecuteTemplate(os.Stdout, "sonar-runner.properties.tmpl", p)
		if err != nil {
			panic(err)
		}
	}

	err = tmpl.ExecuteTemplate(f, "sonar-runner.properties.tmpl", p)
	if err != nil {
		panic(err)
	}

	return nil
}

func (p Plugin) execSonarRunner() error {
	// run archive command
	cmd := exec.Command("java", "-jar", "/opt/sonar/runner.jar", "-Drunner.home=/opt/sonar/")
	printCommand(cmd)
	output, err := cmd.CombinedOutput()
	printOutput(output)

	if err != nil {
		return err
	}

	return nil
}

func (p Plugin) writePipelineLetter() {

	f, err := os.OpenFile(".Pipeline-Letter", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("!!> Error creating / appending to .Pipeline-Letter")
		return
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("*SONAR*: %s/dashboard/index/%s\n", p.Host, strings.Replace(p.Key, "/", ":", -1))); err != nil {
		fmt.Printf("!!> Error writing to .Pipeline-Letter")
	}
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}
