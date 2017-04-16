package command

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/mylxsw/coyotes/console"
	"github.com/mylxsw/coyotes/log"
	"github.com/mylxsw/coyotes/brokers"
)

type ShellCommand struct {
	output chan brokers.Output
	task   brokers.Task
}

func CreateShellCommand(task brokers.Task, outputChan chan brokers.Output) *ShellCommand {
	return &ShellCommand{
		output: outputChan,
		task:   task,
	}
}

// Execute 执行命令，绑定输出
// 返回值（是否成功，错误）
func (self *ShellCommand) Execute(processID string) (bool, error) {
	params := strings.Split(self.task.TaskName, " ")
	cmd := exec.Command(params[0], params[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return false, err
	}

	log.Debug(
		"[%s] command started: %s",
		console.ColorfulText(console.TextRed, processID),
		console.ColorfulText(console.TextGreen, self.task.TaskName),
	)
	if err := cmd.Start(); err != nil {
		return false, err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		self.bindOutput(processID, &stdout)
	}()
	go func() {
		defer wg.Done()
		self.bindOutput(processID, &stderr)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return false, err
	}

	if cmd.ProcessState.Success() {
		log.Debug(
			"[%s] command [%s] execution success",
			console.ColorfulText(console.TextRed, processID),
			console.ColorfulText(console.TextGreen, self.task.TaskName),
		)
	} else {
		log.Debug(
			"[%s] command [%s] execution failed",
			console.ColorfulText(console.TextRed, processID),
			console.ColorfulText(console.TextGreen, self.task.TaskName),
		)
	}

	return cmd.ProcessState.Success(), nil
}

// bindOutput 绑定标准输入、输出到输出channel
func (self *ShellCommand) bindOutput(processID string, input *io.ReadCloser) error {
	reader := bufio.NewReader(*input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			if err != io.EOF {
				return fmt.Errorf(
					"[%s] command [%s] execution failed: %s",
					console.ColorfulText(console.TextRed, processID),
					console.ColorfulText(console.TextGreen, self.task.TaskName),
					console.ColorfulText(console.TextRed, err.Error()),
				)
			}
			break
		}

		self.output <- brokers.Output{
			ProcessID: processID,
			Task:      self.task,
			Content:   strings.TrimRight(line, "\n"),
		}
	}

	return nil
}
