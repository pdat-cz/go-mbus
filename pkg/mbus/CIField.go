package mbus

// CIField is a byte.
// The CI Field is the first byte of the data field
type CIField byte

func NewCIField(b byte) CIField {
	return CIField(b)
}

const (
	CiFieldDataSend          CIField = 0x51
	CiFieldSelectionOfSlaves         = 0x52
	// CiFieldApplicationReset Master can release a reset of application program in the slaves
	CiFieldApplicationReset                    = 0x50
	CiFieldSynchronizeAction                   = 0x54
	CiFieldVariable72                          = 0x72
	CiFieldVariable76                          = 0x76
	CiFieldBaudrate300                         = 0xB8
	CiFieldBaudrate1200                        = 0xBA
	CiFieldBaudrate2400                        = 0xBB
	CiFieldBaudrate4800                        = 0xBC
	CiFieldBaudrate9600                        = 0xBD
	CiFieldBaudrate19200                       = 0xBE
	CiFieldBaudrate38400                       = 0xBF
	CiFieldRequestReadoutCompleteRam           = 0xB1
	CiFieldSendUserDataNotStandardizedRamWrite = 0xB2
	CiFieldInitializeTestCalibrationMode       = 0xB3
	CiFieldEepromRead                          = 0xB4
	CiFieldStartSoftwareTest                   = 0xB6
	CiFieldCodesUsedForHashing0                = 0x90
	CiFieldCodesUsedForHashing1                = 0x91
	CiFieldCodesUsedForHashing2                = 0x92
	CiFieldCodesUsedForHashing3                = 0x93
	CiFieldCodesUsedForHashing4                = 0x94
	CiFieldCodesUsedForHashing5                = 0x95
	CiFieldCodesUsedForHashing6                = 0x96
	CiFieldCodesUsedForHashing7                = 0x97
)

func (cf CIField) String() string {
	switch cf {
	case CiFieldDataSend:
		return "DATA_SEND"
	case CiFieldSelectionOfSlaves:
		return "SELECTION_OF_SLAVES"
	case CiFieldApplicationReset:
		return "APPLICATION_RESET"
	case CiFieldSynchronizeAction:
		return "SYNCHRONIZE_ACTION"
	case CiFieldVariable72:
		return "VARIABLE_DATA_STRUCTURE_72"
	case CiFieldVariable76:
		return "VARIABLE_DATA_STRUCTURE_76"
	case CiFieldBaudrate300:
		return "BAUDRATE_300"
	case CiFieldBaudrate1200:
		return "BAUDRATE_1200"
	case CiFieldBaudrate2400:
		return "BAUDRATE_2400"
	case CiFieldBaudrate4800:
		return "BAUDRATE_4800"
	case CiFieldBaudrate9600:
		return "BAUDRATE_9600"
	case CiFieldBaudrate19200:
		return "BAUDRATE_19200"
	case CiFieldBaudrate38400:
		return "BAUDRATE_38400"
	case CiFieldRequestReadoutCompleteRam:
		return "REQUEST_READOUT_COMPLETE_RAM"
	case CiFieldSendUserDataNotStandardizedRamWrite:
		return "SEND_USER_DATA_NOT_STANDARDIZED_RAM_WRITE"
	case CiFieldInitializeTestCalibrationMode:
		return "INITIALIZE_TEST_CALIBRATION_MODE"
	case CiFieldEepromRead:
		return "EEPROM_READ"
	case CiFieldStartSoftwareTest:
		return "START_SOFTWARE_TEST"
	case CiFieldCodesUsedForHashing0:
		return "CODES_USED_FOR_HASHING_0"
	case CiFieldCodesUsedForHashing1:
		return "CODES_USED_FOR_HASHING_1"
	case CiFieldCodesUsedForHashing2:
		return "CODES_USED_FOR_HASHING_2"
	case CiFieldCodesUsedForHashing3:
		return "CODES_USED_FOR_HASHING_3"
	case CiFieldCodesUsedForHashing4:
		return "CODES_USED_FOR_HASHING_4"
	case CiFieldCodesUsedForHashing5:
		return "CODES_USED_FOR_HASHING_5"
	case CiFieldCodesUsedForHashing6:
		return "CODES_USED_FOR_HASHING_6"
	case CiFieldCodesUsedForHashing7:
		return "CODES_USED_FOR_HASHING_7"
	}
	return "UNDEFINED"
}
