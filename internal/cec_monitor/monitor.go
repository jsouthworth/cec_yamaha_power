package cec_monitor

/*
#cgo pkg-config: libcec
#include <libcec/cecc.h>
#include <stdio.h>
#include <stdlib.h>

extern void callbackLogMessage(void* data, const cec_log_message* message);
extern void callbackKeyPress(void* data, const cec_keypress* key);
extern void callbackCommand(void* data, const cec_command* command);

ICECCallbacks g_callbacks;
void initCallbacks(libcec_configuration *conf)
{
	g_callbacks.logMessage = &callbackLogMessage,
	g_callbacks.keyPress = &callbackKeyPress;
	g_callbacks.commandReceived = &callbackCommand;
	g_callbacks.configurationChanged = NULL;
	g_callbacks.alert = NULL;
	g_callbacks.menuStateChanged = NULL;
	g_callbacks.sourceActivated = NULL;
	(*conf).callbacks = &g_callbacks;
}
void setName(libcec_configuration *conf, char* name)
{
	snprintf((*conf).strDeviceName, LIBCEC_OSD_NAME_SIZE, "%s", name);
}
*/
import "C"
import (
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

type strError string

func (e strError) Error() string {
	return string(e)
}

const (
	errCECInitFailed  strError = "failed to initialize cec"
	errUnknownAdapter strError = "unknown adapter"
	errAdapterFailed  strError = "failed to open adapter"
)

type CECLogicalAddress int8

const (
	CECDEVICE_UNKNOWN          CECLogicalAddress = -1
	CECDEVICE_TV               CECLogicalAddress = 0
	CECDEVICE_RECORDINGDEVICE1 CECLogicalAddress = 1
	CECDEVICE_RECORDINGDEVICE2 CECLogicalAddress = 2
	CECDEVICE_TUNER1           CECLogicalAddress = 3
	CECDEVICE_PLAYBACKDEVICE1  CECLogicalAddress = 4
	CECDEVICE_AUDIOSYSTEM      CECLogicalAddress = 5
	CECDEVICE_TUNER2           CECLogicalAddress = 6
	CECDEVICE_TUNER3           CECLogicalAddress = 7
	CECDEVICE_PLAYBACKDEVICE2  CECLogicalAddress = 8
	CECDEVICE_RECORDINGDEVICE3 CECLogicalAddress = 9
	CECDEVICE_TUNER4           CECLogicalAddress = 10
	CECDEVICE_PLAYBACKDEVICE3  CECLogicalAddress = 11
	CECDEVICE_RESERVED1        CECLogicalAddress = 12
	CECDEVICE_RESERVED2        CECLogicalAddress = 13
	CECDEVICE_FREEUSE          CECLogicalAddress = 14
	CECDEVICE_UNREGISTERED     CECLogicalAddress = 15
	CECDEVICE_BROADCAST        CECLogicalAddress = 15
)

type CECOpcode uint16

const (
	CEC_OPCODE_ACTIVE_SOURCE                 CECOpcode = 0x82
	CEC_OPCODE_IMAGE_VIEW_ON                 CECOpcode = 0x04
	CEC_OPCODE_TEXT_VIEW_ON                  CECOpcode = 0x0D
	CEC_OPCODE_INACTIVE_SOURCE               CECOpcode = 0x9D
	CEC_OPCODE_REQUEST_ACTIVE_SOURCE         CECOpcode = 0x85
	CEC_OPCODE_ROUTING_CHANGE                CECOpcode = 0x80
	CEC_OPCODE_ROUTING_INFORMATION           CECOpcode = 0x81
	CEC_OPCODE_SET_STREAM_PATH               CECOpcode = 0x86
	CEC_OPCODE_STANDBY                       CECOpcode = 0x36
	CEC_OPCODE_RECORD_OFF                    CECOpcode = 0x0B
	CEC_OPCODE_RECORD_ON                     CECOpcode = 0x09
	CEC_OPCODE_RECORD_STATUS                 CECOpcode = 0x0A
	CEC_OPCODE_RECORD_TV_SCREEN              CECOpcode = 0x0F
	CEC_OPCODE_CLEAR_ANALOGUE_TIMER          CECOpcode = 0x33
	CEC_OPCODE_CLEAR_DIGITAL_TIMER           CECOpcode = 0x99
	CEC_OPCODE_CLEAR_EXTERNAL_TIMER          CECOpcode = 0xA1
	CEC_OPCODE_SET_ANALOGUE_TIMER            CECOpcode = 0x34
	CEC_OPCODE_SET_DIGITAL_TIMER             CECOpcode = 0x97
	CEC_OPCODE_SET_EXTERNAL_TIMER            CECOpcode = 0xA2
	CEC_OPCODE_SET_TIMER_PROGRAM_TITLE       CECOpcode = 0x67
	CEC_OPCODE_TIMER_CLEARED_STATUS          CECOpcode = 0x43
	CEC_OPCODE_TIMER_STATUS                  CECOpcode = 0x35
	CEC_OPCODE_CEC_VERSION                   CECOpcode = 0x9E
	CEC_OPCODE_GET_CEC_VERSION               CECOpcode = 0x9F
	CEC_OPCODE_GIVE_PHYSICAL_ADDRESS         CECOpcode = 0x83
	CEC_OPCODE_GET_MENU_LANGUAGE             CECOpcode = 0x91
	CEC_OPCODE_REPORT_PHYSICAL_ADDRESS       CECOpcode = 0x84
	CEC_OPCODE_SET_MENU_LANGUAGE             CECOpcode = 0x32
	CEC_OPCODE_DECK_CONTROL                  CECOpcode = 0x42
	CEC_OPCODE_DECK_STATUS                   CECOpcode = 0x1B
	CEC_OPCODE_GIVE_DECK_STATUS              CECOpcode = 0x1A
	CEC_OPCODE_PLAY                          CECOpcode = 0x41
	CEC_OPCODE_GIVE_TUNER_DEVICE_STATUS      CECOpcode = 0x08
	CEC_OPCODE_SELECT_ANALOGUE_SERVICE       CECOpcode = 0x92
	CEC_OPCODE_SELECT_DIGITAL_SERVICE        CECOpcode = 0x93
	CEC_OPCODE_TUNER_DEVICE_STATUS           CECOpcode = 0x07
	CEC_OPCODE_TUNER_STEP_DECREMENT          CECOpcode = 0x06
	CEC_OPCODE_TUNER_STEP_INCREMENT          CECOpcode = 0x05
	CEC_OPCODE_DEVICE_VENDOR_ID              CECOpcode = 0x87
	CEC_OPCODE_GIVE_DEVICE_VENDOR_ID         CECOpcode = 0x8C
	CEC_OPCODE_VENDOR_COMMAND                CECOpcode = 0x89
	CEC_OPCODE_VENDOR_COMMAND_WITH_ID        CECOpcode = 0xA0
	CEC_OPCODE_VENDOR_REMOTE_BUTTON_DOWN     CECOpcode = 0x8A
	CEC_OPCODE_VENDOR_REMOTE_BUTTON_UP       CECOpcode = 0x8B
	CEC_OPCODE_SET_OSD_STRING                CECOpcode = 0x64
	CEC_OPCODE_GIVE_OSD_NAME                 CECOpcode = 0x46
	CEC_OPCODE_SET_OSD_NAME                  CECOpcode = 0x47
	CEC_OPCODE_MENU_REQUEST                  CECOpcode = 0x8D
	CEC_OPCODE_MENU_STATUS                   CECOpcode = 0x8E
	CEC_OPCODE_USER_CONTROL_PRESSED          CECOpcode = 0x44
	CEC_OPCODE_USER_CONTROL_RELEASE          CECOpcode = 0x45
	CEC_OPCODE_GIVE_DEVICE_POWER_STATUS      CECOpcode = 0x8F
	CEC_OPCODE_REPORT_POWER_STATUS           CECOpcode = 0x90
	CEC_OPCODE_FEATURE_ABORT                 CECOpcode = 0x00
	CEC_OPCODE_ABORT                         CECOpcode = 0xFF
	CEC_OPCODE_GIVE_AUDIO_STATUS             CECOpcode = 0x71
	CEC_OPCODE_GIVE_SYSTEM_AUDIO_MODE_STATUS CECOpcode = 0x7D
	CEC_OPCODE_REPORT_AUDIO_STATUS           CECOpcode = 0x7A
	CEC_OPCODE_SET_SYSTEM_AUDIO_MODE         CECOpcode = 0x72
	CEC_OPCODE_SYSTEM_AUDIO_MODE_REQUEST     CECOpcode = 0x70
	CEC_OPCODE_SYSTEM_AUDIO_MODE_STATUS      CECOpcode = 0x7E
	CEC_OPCODE_SET_AUDIO_RATE                CECOpcode = 0x9A

	/* CEC 1.4 */
	CEC_OPCODE_REPORT_SHORT_AUDIO_DESCRIPTORS  CECOpcode = 0xA3
	CEC_OPCODE_REQUEST_SHORT_AUDIO_DESCRIPTORS CECOpcode = 0xA4
	CEC_OPCODE_START_ARC                       CECOpcode = 0xC0
	CEC_OPCODE_REPORT_ARC_STARTED              CECOpcode = 0xC1
	CEC_OPCODE_REPORT_ARC_ENDED                CECOpcode = 0xC2
	CEC_OPCODE_REQUEST_ARC_START               CECOpcode = 0xC3
	CEC_OPCODE_REQUEST_ARC_END                 CECOpcode = 0xC4
	CEC_OPCODE_END_ARC                         CECOpcode = 0xC5
	CEC_OPCODE_CDC                             CECOpcode = 0xF8

	CEC_OPCODE_NONE CECOpcode = 0xFD
)

type CECUserControlCode uint16

const (
	CEC_USER_CONTROL_CODE_SELECT        CECUserControlCode = 0x00
	CEC_USER_CONTROL_CODE_UP            CECUserControlCode = 0x01
	CEC_USER_CONTROL_CODE_DOWN          CECUserControlCode = 0x02
	CEC_USER_CONTROL_CODE_LEFT          CECUserControlCode = 0x03
	CEC_USER_CONTROL_CODE_RIGHT         CECUserControlCode = 0x04
	CEC_USER_CONTROL_CODE_RIGHT_UP      CECUserControlCode = 0x05
	CEC_USER_CONTROL_CODE_RIGHT_DOWN    CECUserControlCode = 0x06
	CEC_USER_CONTROL_CODE_LEFT_UP       CECUserControlCode = 0x07
	CEC_USER_CONTROL_CODE_LEFT_DOWN     CECUserControlCode = 0x08
	CEC_USER_CONTROL_CODE_ROOT_MENU     CECUserControlCode = 0x09
	CEC_USER_CONTROL_CODE_SETUP_MENU    CECUserControlCode = 0x0A
	CEC_USER_CONTROL_CODE_CONTENTS_MENU CECUserControlCode = 0x0B
	CEC_USER_CONTROL_CODE_FAVORITE_MENU CECUserControlCode = 0x0C
	CEC_USER_CONTROL_CODE_EXIT          CECUserControlCode = 0x0D
	// reserved: 0x0E 0x0F
	CEC_USER_CONTROL_CODE_TOP_MENU CECUserControlCode = 0x10
	CEC_USER_CONTROL_CODE_DVD_MENU CECUserControlCode = 0x11
	// reserved: 0x12 ... 0x1C
	CEC_USER_CONTROL_CODE_NUMBER_ENTRY_MODE   CECUserControlCode = 0x1D
	CEC_USER_CONTROL_CODE_NUMBER11            CECUserControlCode = 0x1E
	CEC_USER_CONTROL_CODE_NUMBER12            CECUserControlCode = 0x1F
	CEC_USER_CONTROL_CODE_NUMBER0             CECUserControlCode = 0x20
	CEC_USER_CONTROL_CODE_NUMBER1             CECUserControlCode = 0x21
	CEC_USER_CONTROL_CODE_NUMBER2             CECUserControlCode = 0x22
	CEC_USER_CONTROL_CODE_NUMBER3             CECUserControlCode = 0x23
	CEC_USER_CONTROL_CODE_NUMBER4             CECUserControlCode = 0x24
	CEC_USER_CONTROL_CODE_NUMBER5             CECUserControlCode = 0x25
	CEC_USER_CONTROL_CODE_NUMBER6             CECUserControlCode = 0x26
	CEC_USER_CONTROL_CODE_NUMBER7             CECUserControlCode = 0x27
	CEC_USER_CONTROL_CODE_NUMBER8             CECUserControlCode = 0x28
	CEC_USER_CONTROL_CODE_NUMBER9             CECUserControlCode = 0x29
	CEC_USER_CONTROL_CODE_DOT                 CECUserControlCode = 0x2A
	CEC_USER_CONTROL_CODE_ENTER               CECUserControlCode = 0x2B
	CEC_USER_CONTROL_CODE_CLEAR               CECUserControlCode = 0x2C
	CEC_USER_CONTROL_CODE_NEXT_FAVORITE       CECUserControlCode = 0x2F
	CEC_USER_CONTROL_CODE_CHANNEL_UP          CECUserControlCode = 0x30
	CEC_USER_CONTROL_CODE_CHANNEL_DOWN        CECUserControlCode = 0x31
	CEC_USER_CONTROL_CODE_PREVIOUS_CHANNEL    CECUserControlCode = 0x32
	CEC_USER_CONTROL_CODE_SOUND_SELECT        CECUserControlCode = 0x33
	CEC_USER_CONTROL_CODE_INPUT_SELECT        CECUserControlCode = 0x34
	CEC_USER_CONTROL_CODE_DISPLAY_INFORMATION CECUserControlCode = 0x35
	CEC_USER_CONTROL_CODE_HELP                CECUserControlCode = 0x36
	CEC_USER_CONTROL_CODE_PAGE_UP             CECUserControlCode = 0x37
	CEC_USER_CONTROL_CODE_PAGE_DOWN           CECUserControlCode = 0x38
	// reserved: 0x39 ... 0x3F
	CEC_USER_CONTROL_CODE_POWER        CECUserControlCode = 0x40
	CEC_USER_CONTROL_CODE_VOLUME_UP    CECUserControlCode = 0x41
	CEC_USER_CONTROL_CODE_VOLUME_DOWN  CECUserControlCode = 0x42
	CEC_USER_CONTROL_CODE_MUTE         CECUserControlCode = 0x43
	CEC_USER_CONTROL_CODE_PLAY         CECUserControlCode = 0x44
	CEC_USER_CONTROL_CODE_STOP         CECUserControlCode = 0x45
	CEC_USER_CONTROL_CODE_PAUSE        CECUserControlCode = 0x46
	CEC_USER_CONTROL_CODE_RECORD       CECUserControlCode = 0x47
	CEC_USER_CONTROL_CODE_REWIND       CECUserControlCode = 0x48
	CEC_USER_CONTROL_CODE_FAST_FORWARD CECUserControlCode = 0x49
	CEC_USER_CONTROL_CODE_EJECT        CECUserControlCode = 0x4A
	CEC_USER_CONTROL_CODE_FORWARD      CECUserControlCode = 0x4B
	CEC_USER_CONTROL_CODE_BACKWARD     CECUserControlCode = 0x4C
	CEC_USER_CONTROL_CODE_STOP_RECORD  CECUserControlCode = 0x4D
	CEC_USER_CONTROL_CODE_PAUSE_RECORD CECUserControlCode = 0x4E
	// reserved: 0x4F
	CEC_USER_CONTROL_CODE_ANGLE                     CECUserControlCode = 0x50
	CEC_USER_CONTROL_CODE_SUB_PICTURE               CECUserControlCode = 0x51
	CEC_USER_CONTROL_CODE_VIDEO_ON_DEMAND           CECUserControlCode = 0x52
	CEC_USER_CONTROL_CODE_ELECTRONIC_PROGRAM_GUIDE  CECUserControlCode = 0x53
	CEC_USER_CONTROL_CODE_TIMER_PROGRAMMING         CECUserControlCode = 0x54
	CEC_USER_CONTROL_CODE_INITIAL_CONFIGURATION     CECUserControlCode = 0x55
	CEC_USER_CONTROL_CODE_SELECT_BROADCAST_TYPE     CECUserControlCode = 0x56
	CEC_USER_CONTROL_CODE_SELECT_SOUND_PRESENTATION CECUserControlCode = 0x57
	// reserved: 0x58 ... 0x5F
	CEC_USER_CONTROL_CODE_PLAY_FUNCTION               CECUserControlCode = 0x60
	CEC_USER_CONTROL_CODE_PAUSE_PLAY_FUNCTION         CECUserControlCode = 0x61
	CEC_USER_CONTROL_CODE_RECORD_FUNCTION             CECUserControlCode = 0x62
	CEC_USER_CONTROL_CODE_PAUSE_RECORD_FUNCTION       CECUserControlCode = 0x63
	CEC_USER_CONTROL_CODE_STOP_FUNCTION               CECUserControlCode = 0x64
	CEC_USER_CONTROL_CODE_MUTE_FUNCTION               CECUserControlCode = 0x65
	CEC_USER_CONTROL_CODE_RESTORE_VOLUME_FUNCTION     CECUserControlCode = 0x66
	CEC_USER_CONTROL_CODE_TUNE_FUNCTION               CECUserControlCode = 0x67
	CEC_USER_CONTROL_CODE_SELECT_MEDIA_FUNCTION       CECUserControlCode = 0x68
	CEC_USER_CONTROL_CODE_SELECT_AV_INPUT_FUNCTION    CECUserControlCode = 0x69
	CEC_USER_CONTROL_CODE_SELECT_AUDIO_INPUT_FUNCTION CECUserControlCode = 0x6A
	CEC_USER_CONTROL_CODE_POWER_TOGGLE_FUNCTION       CECUserControlCode = 0x6B
	CEC_USER_CONTROL_CODE_POWER_OFF_FUNCTION          CECUserControlCode = 0x6C
	CEC_USER_CONTROL_CODE_POWER_ON_FUNCTION           CECUserControlCode = 0x6D
	// reserved: 0x6E ... 0x70
	CEC_USER_CONTROL_CODE_F1_BLUE   CECUserControlCode = 0x71
	CEC_USER_CONTROL_CODE_F2_RED    CECUserControlCode = 0x72
	CEC_USER_CONTROL_CODE_F3_GREEN  CECUserControlCode = 0x73
	CEC_USER_CONTROL_CODE_F4_YELLOW CECUserControlCode = 0x74
	CEC_USER_CONTROL_CODE_F5        CECUserControlCode = 0x75
	CEC_USER_CONTROL_CODE_DATA      CECUserControlCode = 0x76
	// reserved: 0x77 ... 0xFF
	CEC_USER_CONTROL_CODE_AN_RETURN        CECUserControlCode = 0x91 // return (Samsung)
	CEC_USER_CONTROL_CODE_AN_CHANNELS_LIST CECUserControlCode = 0x96 // channels list (Samsung)
	CEC_USER_CONTROL_CODE_MAX              CECUserControlCode = 0x96
	CEC_USER_CONTROL_CODE_UNKNOWN          CECUserControlCode = 0xFF
)

type Command struct {
	Initiator   CECLogicalAddress
	Destination CECLogicalAddress
	Ack         bool
	Eom         bool
	OpcodeSet   bool
	Opcode      CECOpcode
	Params      []byte
	TXTimeout   uint32
}

type KeyPress struct {
	Code     CECUserControlCode
	Duration uint
}

type connFns struct {
	logMessage      func(s string)
	keyPress        func(press *KeyPress)
	commandReceived func(cmd *Command)
}

type ConnFn func(fns *connFns)

func OnLogMessage(fn func(s string)) ConnFn {
	return func(fns *connFns) {
		fns.logMessage = fn
	}
}

func OnKeyPress(fn func(press *KeyPress)) ConnFn {
	return func(fns *connFns) {
		fns.keyPress = fn
	}
}

func OnCommandReceived(fn func(cmd *Command)) ConnFn {
	return func(fns *connFns) {
		fns.commandReceived = fn
	}
}

type Conn struct {
	impl C.libcec_connection_t
	fns  connFns
}

func initCEC(data cHandle, deviceName string) (C.libcec_connection_t, error) {
	var conn C.libcec_connection_t
	var conf C.libcec_configuration
	conf.callbackParam = unsafe.Pointer(data)
	conf.clientVersion = C.uint32_t(C.LIBCEC_VERSION_CURRENT)
	conf.bMonitorOnly = 1
	C.initCallbacks(&conf)

	cname := C.CString(deviceName)
	defer C.free(unsafe.Pointer(cname))
	C.setName(&conf, cname)

	conn = C.libcec_initialise(&conf)
	if conn == C.libcec_connection_t(nil) {
		return nil, errCECInitFailed
	}

	return conn, nil
}

func getAdapter(conn C.libcec_connection_t, name string) (string, error) {
	var deviceList [10]C.cec_adapter
	devicesFound :=
		int(C.libcec_find_adapters(conn, &deviceList[0], 10, nil))

	for i := 0; i < devicesFound; i++ {
		device := deviceList[i]
		path := C.GoStringN(&device.path[0], 1024)
		comm := C.GoStringN(&device.comm[0], 1024)

		if strings.Contains(path, name) ||
			strings.Contains(comm, name) {
			return comm, nil
		}
	}

	return "", errUnknownAdapter
}

func openAdapter(conn C.libcec_connection_t, adapter string) error {
	C.libcec_init_video_standalone(conn)

	result := C.libcec_open(
		conn, C.CString(adapter), C.CEC_DEFAULT_CONNECT_TIMEOUT)
	if result < 1 {
		return errAdapterFailed
	}

	return nil
}

func Open(adapter, deviceName string, fns ...ConnFn) (*Conn, error) {
	var cfns connFns
	for _, fn := range fns {
		fn(&cfns)
	}
	out := &Conn{
		fns: cfns,
	}
	handle := cHandleCreate(out)
	conn, err := initCEC(handle, deviceName)
	if err != nil {
		return nil, err
	}
	out.impl = conn
	comm, err := getAdapter(out.impl, adapter)
	if err != nil {
		return nil, err
	}
	err = openAdapter(out.impl, comm)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Conn) Close() {
	C.libcec_destroy(c.impl)
}

var (
	cHandles   = sync.Map{}
	cHandleIdx uintptr
)

const (
	errOutOfCHandles  strError = "opds: ran out of space for C handles"
	errInvalidCHandle strError = "opds: invalid C handle"
)

type cHandle uintptr

func cHandleCreate(v interface{}) cHandle {
	h := atomic.AddUintptr(&cHandleIdx, 1)
	if h == 0 {
		panic(errOutOfCHandles)
	}

	cHandles.Store(h, v)
	return cHandle(h)
}

func (h cHandle) Value() interface{} {
	v, ok := cHandles.Load(uintptr(h))
	if !ok {
		panic(errInvalidCHandle)
	}
	return v
}

func (h cHandle) Delete() {
	_, ok := cHandles.LoadAndDelete(uintptr(h))
	if !ok {
		panic(errInvalidCHandle)
	}
}
