// build +linux
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/sys/unix"

	cec "github.com/jsouthworth/cec_yamaha_power/internal/cec_monitor"
	i2c "github.com/jsouthworth/cec_yamaha_power/internal/i2c_remote_controller"
)

type IRCmd struct {
	proto i2c.IRProtocol
	addr  uint16
	cmd   uint16
}

func (c IRCmd) String() string {
	var proto string
	switch c.proto {
	case i2c.NEC:
		proto = "NEC"
	case i2c.ONKYO:
		proto = "ONKYO"
	}
	return fmt.Sprintf("%s\t%x\t%x", proto, c.addr, c.cmd)
}

const (
	CECDev  = "/dev/ttyACM0"
	I2CBus  = ""
	I2CAddr = 0x08
)

var IRCmds = map[cec.CECOpcode][]IRCmd{
	cec.CEC_OPCODE_ACTIVE_SOURCE: []IRCmd{
		{i2c.NEC, 0x7E, 0x7E},       // power on
		{i2c.ONKYO, 0x857A, 0x7F00}, // select scene 1
	},
	cec.CEC_OPCODE_STANDBY: []IRCmd{
		{i2c.NEC, 0x7E, 0x7F}, // power off
	},
}

var cmds = map[string]func() int{
	"supervisor": supervisor,
	"relay":      relay,
	"init":       initProc,
}

func relay() int {
	fmt.Println("CEC yamaha power relay process")
	files, err := ioutil.ReadDir("/dev")
	if err == nil {
		for _, file := range files {
			fmt.Println(filepath.Join("/dev", file.Name()))
		}
	}
	opcodes := make(chan cec.CECOpcode, 1)
	conn, err := cec.Open(CECDev, "monitor",
		cec.OnCommandReceived(
			func(cmd *cec.Command) {
				opcodes <- cmd.Opcode
			}))
	if err != nil {
		fmt.Fprintln(os.Stderr, "relay cec.Open:", err)
		return 1
	}
	defer conn.Close()
	ir, err := i2c.Open(I2CBus, I2CAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "relay i2c.Open:", err)
		return 1
	}
	defer ir.Close()
	for opcode := range opcodes {
		fmt.Printf("received CEC opcode: %x\n", opcode)
		ircmds := IRCmds[opcode]
		for _, cmd := range ircmds {
			fmt.Printf("Sending cmd: %s\n", cmd)
			ir.Send(cmd.proto, cmd.addr, cmd.cmd)
			time.Sleep(1 * time.Second)
		}
	}
	return 0
}

func supervisor() int {
	fmt.Println("CEC yamaha power supervisor")
	name, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "supervisor:", err)
		return 1
	}
	for {
		cmd := exec.Command(name, "-cmd", "relay")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "supervisor: relay:", err)
		}
		// TODO: figure out how to break out of the loop when
		// failures continually occur.
		time.Sleep(2 * time.Second)
	}
}

func mountProcfs() error {
	return unix.Mount("proc", "/proc", "proc", 0, "")
}

func mountSysfs() error {
	return unix.Mount("sys", "/sys", "sysfs", 0, "")
}

func reboot() {
	unix.Syscall(
		unix.SYS_REBOOT, unix.LINUX_REBOOT_CMD_RESTART, 0, 0)
}

func halt() {
	unix.Syscall(
		unix.SYS_REBOOT, unix.LINUX_REBOOT_CMD_HALT, 0, 0)
}

func initProc() int {
	fmt.Println("CEC yamaha power init process")
	err := mountProcfs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "init-proc: mounting procfs:", err)
		halt()
	}
	err = mountSysfs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "init-proc: mounting sysfs:", err)
		halt()
	}
	rc := supervisor()
	if rc != 0 {
		halt()
	}
	reboot()
	return 0
}

var cmd string

func init() {
	flag.StringVar(&cmd, "cmd", "init", "command to run")
}

func main() {
	flag.Parse()
	fn, ok := cmds[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
	os.Exit(fn())
}
