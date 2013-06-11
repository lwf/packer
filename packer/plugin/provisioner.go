package plugin

import (
	"github.com/mitchellh/packer/packer"
	packrpc "github.com/mitchellh/packer/packer/rpc"
	"log"
	"net/rpc"
	"os/exec"
)

type cmdProvisioner struct {
	p      packer.Provisioner
	client *client
}

func (c *cmdProvisioner) Prepare(configs ...interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Prepare(configs...)
}

func (c *cmdProvisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	c.p.Provision(ui, comm)
}

func (c *cmdProvisioner) checkExit(p interface{}, cb func()) {
	if c.client.Exited() {
		cb()
	} else if p != nil {
		log.Panic(p)
	}
}

// Returns a valid packer.Provisioner where the hook is executed via RPC
// to a plugin that is within a subprocess.
//
// This method will start the given exec.Cmd, which should point to
// the plugin binary to execute. Some configuration will be done to
// the command, such as overriding Stdout and some environmental variables.
//
// This function guarantees the subprocess will end in a timely manner.
func Provisioner(cmd *exec.Cmd) (result packer.Provisioner, err error) {
	config := &ClientConfig{
		Cmd: cmd,
		Managed: true,
	}

	cmdClient := NewClient(config)
	address, err := cmdClient.Start()
	if err != nil {
		return
	}

	defer func() {
		// Make sure the command is properly killed in the case of an error
		if err != nil {
			cmdClient.Kill()
		}
	}()

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return
	}

	result = &cmdProvisioner{
		packrpc.Provisioner(client),
		cmdClient,
	}

	return
}
