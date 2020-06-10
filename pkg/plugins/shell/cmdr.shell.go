// Copyright © 2020 Hedzr Yeh.

package shell

import (
	"bytes"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/hedzr/cmdr"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// WithShellModule adds `shell` sub-command to cmdr system
func WithShellModule() cmdr.ExecOption {
	return func(w *cmdr.ExecWorker) {
		// daemonImpl = daemonImplX

		w.AddOnBeforeXrefBuilding(func(root *cmdr.RootCommand, args []string) {

			// if modifier != nil {
			// 	root.SubCommands = append(root.SubCommands, modifier(DaemonServerCommand))
			// } else {
			// 	root.SubCommands = append(root.SubCommands, DaemonServerCommand)
			// }
			//
			// // prefix = strings.Join(append(cmdr.RxxtPrefix, "server"), ".")
			// // prefix = "server"
			//
			// attachPreAction(root, preAction)
			// attachPostAction(root, postAction)

			rootOpt := cmdr.NewCmdFrom(&root.Command)
			shell := rootOpt.NewSubCommand().
				Titles("shell", "sh").
				Description("enable internal shell prompt mode", "").
				Action(doShellModeAction)
			shell.NewFlagV(true).Titles("demo-0", "0").ToggleGroup("Sample")
			shell.NewFlagV(false).Titles("demo-1", "1").ToggleGroup("Sample")
			shell.NewFlagV(false).Titles("demo-2", "2").ToggleGroup("Sample")
			shell.NewFlagV(false).Titles("demo-3", "3").ToggleGroup("Sample")
			shell.NewFlagV(false).Titles("demo-4", "4").ToggleGroup("Sample")
		})
	}
}

func doShellModeAction(cmd *cmdr.Command, args []string) (err error) {
	// logrus.Println(cmdr.GetStringR("shell.Sample"))
	switch cmdr.GetStringR("shell.Sample") {
	case "demo-0":
		run0()
	case "demo-1":
		run1()
	case "demo-2":
		run2()
	case "demo-3":
		run3()
	case "demo-4":
		run4()
	}
	return
}

func run0() {
	fmt.Println("cmdr shell prompt here...Type 'exit' or press 'Ctrl-D' to exit.")
	p := prompt.New(executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("cmdr-prompt-mode here"),
	)
	p.Run()
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	} else if in == "quit" || in == "exit" || in == "bye" {
		os.Exit(0)
	}

	fmt.Println("Your input: " + in + " | for running " + cmdr.GetExcutablePath())

	// cmd := exec.Command("/bin/sh", "-c", cmdr.GetExcutablePath()+" "+in)
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	fmt.Printf("Got error: %s\n", err.Error())
	// }
}

func completer(in prompt.Document) []prompt.Suggest {
	cmd, err := cmdr.Match(in.Text)

	// var names string
	// if cmd != nil {
	// 	names = cmd.GetTitleNames()
	// }
	// logrus.Debugf("in: %v, cmd: %q, err = %v", in, names, err)

	if err != nil || cmd == nil {
		return prompt.FilterHasPrefix([]prompt.Suggest{}, in.GetWordBeforeCursor(), true)
	}

	// logrus.Debugf("cmd = %v", cmd)
	return prompt.FilterHasPrefix(buildSuggestsFor(in.Text, cmd), in.GetWordBeforeCursor(), true)
}

func buildSuggestsFor(text string, command *cmdr.Command) (ret []prompt.Suggest) {
	for _, cmd := range command.SubCommands {
		ret = append(ret, prompt.Suggest{Text: cmd.GetTitleName(), Description: cmd.Description})
	}

	for _, flg := range command.Flags {
		ret = append(ret, prompt.Suggest{Text: flg.GetTitleZshFlagName(), Description: flg.Description})
	}

	// var sa []string
	// for _, r := range ret {
	// 	sa = append(sa, r.Text)
	// }
	// logrus.Debugf("%q: cmd.Titles: %q; ret: %v | %v", text, command.GetTitleNames(), len(ret), sa)

	return
}

// ExecuteAndGetResult executes an command line string `s` and capture its outputs.
func ExecuteAndGetResult(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		logrus.Debug("you need to pass the something arguments")
		return ""
	}

	out := &bytes.Buffer{}
	cmd := exec.Command("/bin/sh", "-c", "kubectl "+s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		logrus.Error(err.Error())
		return ""
	}
	r := string(out.Bytes())
	return r
}

func run1() {
	fmt.Println("A bash like sample 1...Type 'Ctrl-D' to exit.")
	p := prompt.New(executor1,
		completer1,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("sql-prompt for multi width characters"),
	)
	p.Run()
}

func executor1(in string) {
	fmt.Println("Your input: " + in)
}

func completer1(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "こんにちは", Description: "'こんにちは' means 'Hello' in Japanese"},
		{Text: "감사합니다", Description: "'안녕하세요' means 'Hello' in Korean."},
		{Text: "您好", Description: "'您好' means 'Hello' in Chinese."},
		{Text: "Добрый день", Description: "'Добрый день' means 'Hello' in Russian."},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func run2() {
	fmt.Println("A bash like sample 2...Type 'Ctrl-D' to exit.")
	in := prompt.Input(">>> ", completer2,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
	fmt.Println("Your input: " + in)
}

func completer2(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
		{Text: "groups", Description: "Combine users with specific rules"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func run3() {
	fmt.Println("A bash like sample: run external command...Type 'Ctrl-D' to exit.")
	p := prompt.New(
		executor3,
		completer3,
	)
	p.Run()
}

func executor3(t string) {
	if t == "bash" {
		cmd := exec.Command("bash")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		cmd := exec.Command(t)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	return
}

func completer3(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "bash"},
	}
}

func run4() {
	p := prompt.New(
		executor4,
		completer4,
		prompt.OptionPrefix(">>> "),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("live-prefix-example"),
	)
	p.Run()
}

// LivePrefixState is a struct for auto-completion
var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func executor4(in string) {
	fmt.Println("Your input: " + in)
	if in == "" {
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = in
		return
	}
	LivePrefixState.LivePrefix = in + "> "
	LivePrefixState.IsEnable = true
}

func completer4(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
		{Text: "groups", Description: "Combine users with specific rules"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}
