package command

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/mylxsw/task-runner/console"
	"github.com/mylxsw/task-runner/log"
)

type Output struct {
	ProcessID string
	Name      string
	Content   string
}

type Command struct {
	Output chan Output
}

// 执行命令，绑定输出
func (self *Command) ExecuteTask(processID string, cmdStr string) error {
	params := strings.Split(cmdStr, " ")
	cmd := exec.Command(params[0], params[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	log.Debug(
		"[%s] command started: %s",
		console.ColorfulText(console.TextRed, processID),
		console.ColorfulText(console.TextGreen, cmdStr),
	)
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		self.bindOutput(processID, cmdStr, &stdout)
	}()
	go func() {
		defer wg.Done()
		self.bindOutput(processID, cmdStr, &stderr)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}

	if cmd.ProcessState.Success() {
		log.Debug(
			"[%s] command [%s] execution success",
			console.ColorfulText(console.TextRed, processID),
			console.ColorfulText(console.TextGreen, cmdStr),
		)
	} else {
		log.Debug(
			"[%s] command [%s] execution failed",
			console.ColorfulText(console.TextRed, processID),
			console.ColorfulText(console.TextGreen, cmdStr),
		)
	}

	return nil
}

// 绑定标准输入、输出到输出channel
func (self *Command) bindOutput(processID string, name string, input *io.ReadCloser) error {
	reader := bufio.NewReader(*input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			if err != io.EOF {
				return fmt.Errorf(
					"[%s] command [%s] execution failed: %s",
					console.ColorfulText(console.TextRed, processID),
					console.ColorfulText(console.TextGreen, name),
					console.ColorfulText(console.TextRed, err.Error()),
				)
			}
			break
		}

		self.Output <- Output{
			ProcessID: processID,
			Name:      name,
			Content:   strings.TrimRight(line, "\n"),
		}
	}

	return nil
}
