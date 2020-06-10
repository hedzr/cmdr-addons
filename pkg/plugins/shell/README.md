# shell-prompt

To start an embedded shell-like prompt, type and run:

```bash
$ go run ./examples/shell-mode/ sh

```

See also: [`shell-mode`](https://github.com/hedzr/cmdr-examples/tree/master/examples/shell-mode) app.



It looks like:

![image](https://user-images.githubusercontent.com/12786150/71587009-11436500-2b57-11ea-890d-a60989a09248.png)

and

![image](https://user-images.githubusercontent.com/12786150/84097765-09f28280-aa38-11ea-9abc-a1581e24a351.png)


Type `exit`, `quit`, or shortcut  `CTRL-D` to terminate the prompt session.



> `shell-prompt` is powered by https://github.com/c-bata/go-prompt.



## Backstage

Simply add `shell.WithShellModule()` to your `cmdr` entry

```go
func Entry() {
	if err := cmdr.Exec(buildRootCmd(),
		shell.WithShellModule(),
	); err != nil {
		logrus.Fatalf("error: %+v", err)
	}
}
```

