package cmd

type Command interface {
	Exec(command string) (stdout, stderr string, err error)
}
