package remotecontrol

import (
	"fmt"
	"io"
	"os/exec"
)

func Exec(cmd string) (string, error) {
	cmdRes := exec.Command("/bin/bash", "-c", cmd)
	stdout, _ := cmdRes.StdoutPipe()

	if err := cmdRes.Start(); err != nil {
		return "", fmt.Errorf("execute failed when Start: %s", err.Error())
	}

	out_bytes, _ := io.ReadAll(stdout)
	stdout.Close()

	if err := cmdRes.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		panic(err)
	}

	return string(out_bytes), nil
}
