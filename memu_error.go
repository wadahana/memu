package memu

const (
	RC_Success 				  = 1
	RC_NotImplement 		  = 2
	RC_InvalidArgument 		  = 3
	RC_CommandResultMalformat = 4
	RC_AndroidVerNotSupport   = 5
	RC_EmulatorNotFound       = 6
	RC_NotEmulator            = 7
	RC_EmulatorNotRunning     = 8
	RC_MakeGrabberKeyFail     = 9
	RC_OpenFileMappingFail    = 10
	RC_GrabberNotInitialized  = 11
	RC_CreateImageFail        = 12
	
	RC_SystemError            = 1000
	RC_MemucError             = 1001
	RC_MemuError              = 1002

)

var ErrorSuccess                  *MEmuError = NewError(RC_Success,                "Success")
var ErrorNotImplement             *MEmuError = NewError(RC_NotImplement,           "Not implement")
var ErrorInvalidArgument          *MEmuError = NewError(RC_InvalidArgument,        "Invalid argument")
var ErrorCommandResultMalformat   *MEmuError = NewError(RC_CommandResultMalformat, "Command Result Malformat")
var ErrorAndroidVersionNotSupport *MEmuError = NewError(RC_AndroidVerNotSupport,   "Android version not support")
var ErrorEmulatorNotFound         *MEmuError = NewError(RC_EmulatorNotFound,       "Emulator not found")
var ErrorNotEmulator              *MEmuError = NewError(RC_NotEmulator,            "Not Emulator")
var ErrorEmulatorNotRunning       *MEmuError = NewError(RC_EmulatorNotRunning,     "Emulator Not Running")

var ErrorMakeGrabberKeyFail 	*MEmuError = NewError(RC_MakeGrabberKeyFail,    "Make Grabber key fail.");
var ErrorOpenFileMapFail    	*MEmuError = NewError(RC_OpenFileMappingFail,   "Open File Map fail.");
var ErrorGrabberNotInit     	*MEmuError = NewError(RC_GrabberNotInitialized, "Grabber Not Initialized");
var ErrorCreateImageFail    	*MEmuError = NewError(RC_CreateImageFail,       "Cannot create image.RGBA");

type MEmuError struct {
 	Code      int
    Message   string
}

func (e *MEmuError) Error() string {
    return e.Message
}

func (e *MEmuError) GetCode() int {
	return e.Code
}


func NewError(code int, msg string) *MEmuError {
    return &MEmuError {
        Code:      code,
        Message:   msg,
    }
}
