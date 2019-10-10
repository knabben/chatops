package command

import (
	"bytes"
	"io"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/kubernetes/scheme"
	"net/url"
	"strings"
)

type ExecOptions struct {
	Command            []string
	Namespace          string
	PodName            string
	ContainerName      string
	Stdin              io.Reader
	CaptureStdout      bool
	CaptureStderr      bool
	PreserveWhitespace bool
}

type PodExec struct {
	ClientSet clientset.Clientset
	Namespace string
	Config *restclient.Config
}

func NewPodExec(config *restclient.Config, clientset kubernetes.Clientset, namespace string) *PodExec {
	return &PodExec{
		ClientSet: clientset,
		Namespace: namespace,
		Config: config,
	}
}

func (p *PodExec) ExecCommandInContainer(podName string, cmd ...string) (string, string, error) {
	return p.ExecCommandInContainerWithFullOutput(podName, cmd...)
}

func (p *PodExec) ExecCommandInContainerWithFullOutput(podName string, cmd ...string) (string, string, error) {
	return p.ExecWithOptions(ExecOptions{
		Command:            cmd,
		Namespace:          p.Namespace,
		PodName:            podName,
		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	})
}

func (p *PodExec) ExecWithOptions(options ExecOptions) (string, string, error) {
	const tty = false

	req := p.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command: options.Command,
		Stdin:   options.Stdin != nil,
		Stdout:  options.CaptureStdout,
		Stderr:  options.CaptureStderr,
		TTY:     tty,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err := execute("POST", req.URL(), p.Config, options.Stdin, &stdout, &stderr, tty)

	if options.PreserveWhitespace {
		return stdout.String(), stderr.String(), err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}
