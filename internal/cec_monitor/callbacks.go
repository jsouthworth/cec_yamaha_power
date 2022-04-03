package cec_monitor

// #include <libcec/cecc.h>
import "C"
import (
	"reflect"
	"unsafe"
)

//export callbackLogMessage
func callbackLogMessage(data unsafe.Pointer, message *C.cec_log_message) {
	h := cHandle(data)
	conn := h.Value().(*Conn)
	if conn.fns.logMessage != nil {
		conn.fns.logMessage(C.GoString(message.message))
	}
}

//export callbackKeyPress
func callbackKeyPress(data unsafe.Pointer, press *C.cec_keypress) {
	h := cHandle(data)
	conn := h.Value().(*Conn)
	if conn.fns.keyPress == nil {
		return
	}
	gpress := KeyPress{
		Code:     CECUserControlCode(press.keycode),
		Duration: uint(press.duration),
	}
	conn.fns.keyPress(&gpress)
}

//export callbackCommand
func callbackCommand(data unsafe.Pointer, cmd *C.cec_command) {
	h := cHandle(data)
	conn := h.Value().(*Conn)
	if conn.fns.commandReceived == nil {
		return
	}
	var params []byte
	paramHdr := (*reflect.SliceHeader)(unsafe.Pointer(&params))
	paramHdr.Data = uintptr(unsafe.Pointer(&cmd.parameters.data))
	paramHdr.Len = int(cmd.parameters.size)
	paramHdr.Cap = int(cmd.parameters.size)
	gcmd := Command{
		Initiator:   CECLogicalAddress(cmd.initiator),
		Destination: CECLogicalAddress(cmd.destination),
		Ack:         cmd.ack == 1,
		Eom:         cmd.eom == 1,
		OpcodeSet:   cmd.opcode_set == 1,
		Opcode:      CECOpcode(cmd.opcode),
		Params:      params,
		TXTimeout:   uint32(cmd.transmit_timeout),
	}
	conn.fns.commandReceived(&gcmd)
}
