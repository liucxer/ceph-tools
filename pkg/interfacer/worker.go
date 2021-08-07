package interfacer

type Worker interface {
	ExecCmd(cmd string) ([]byte, error)
}
