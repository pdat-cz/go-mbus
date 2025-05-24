package mbus

type VIFFieldsRecord struct {
	Unit     string
	Name     string
	Exponent float64
}

var VifFields = map[byte]VIFFieldsRecord{
	0x00: {Unit: "Wh", Name: "Energy", Exponent: 1.0e-3},
	0x01: {Unit: "Wh", Name: "Energy", Exponent: 1.0e-2},
	0x02: {Unit: "Wh", Name: "Energy", Exponent: 1.0e-1},
	0x03: {Unit: "Wh", Name: "Energy", Exponent: 1.0},
	0x04: {Unit: "Wh", Name: "Energy", Exponent: 1.0e1},
	0x05: {Unit: "Wh", Name: "Energy", Exponent: 1.0e2},
	0x06: {Unit: "Wh", Name: "Energy", Exponent: 1.0e3},
	0x07: {Unit: "Wh", Name: "Energy", Exponent: 1.0e4},
	// ENERGY J
	0x08: {Unit: "J", Name: "Energy", Exponent: 1.0e0},
	0x09: {Unit: "J", Name: "Energy", Exponent: 1.0e1},
	0x0A: {Unit: "J", Name: "Energy", Exponent: 1.0e2},
	0x0B: {Unit: "J", Name: "Energy", Exponent: 1.0e3},
	0x0C: {Unit: "J", Name: "Energy", Exponent: 1.0e4},
	0x0D: {Unit: "J", Name: "Energy", Exponent: 1.0e5},
	0x0E: {Unit: "J", Name: "Energy", Exponent: 1.0e6},
	0x0F: {Unit: "J", Name: "Energy", Exponent: 1.0e7},
	// Volume m^3
	0x10: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-6},
	0x11: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-5},
	0x12: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-4},
	0x13: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-3},
	0x14: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-2},
	0x15: {Unit: "m^3", Name: "Volume", Exponent: 1.0e-1},
	0x16: {Unit: "m^3", Name: "Volume", Exponent: 1.0e0},
	0x17: {Unit: "m^3", Name: "Volume", Exponent: 1.0e1},

	// Mass
	0x18: {Unit: "kg", Name: "Mass", Exponent: 1.0e-6},
	0x19: {Unit: "kg", Name: "Mass", Exponent: 1.0e-5},
	0x1A: {Unit: "kg", Name: "Mass", Exponent: 1.0e-4},
	0x1B: {Unit: "kg", Name: "Mass", Exponent: 1.0e-3},
	0x1C: {Unit: "kg", Name: "Mass", Exponent: 1.0e-2},
	0x1D: {Unit: "kg", Name: "Mass", Exponent: 1.0e-1},
	0x1E: {Unit: "kg", Name: "Mass", Exponent: 1.0},
	0x1F: {Unit: "kg", Name: "Mass", Exponent: 1.0e1},

	// On Time
	0x20: {Unit: "s", Name: "On time [seconds]", Exponent: 1.0},
	0x21: {Unit: "s", Name: "On time [minutes]", Exponent: 60.0},
	0x22: {Unit: "s", Name: "On time [hours]", Exponent: 3600.0},
	0x23: {Unit: "s", Name: "On time [days]", Exponent: 86400.0},

	// Operating Time
	0x24: {Unit: "s", Name: "Operating time [seconds]", Exponent: 1.0},
	0x25: {Unit: "s", Name: "Operating time [minutes]", Exponent: 60.0},
	0x26: {Unit: "s", Name: "Operating time [hours]", Exponent: 3600.0},
	0x27: {Unit: "s", Name: "Operating time [days]", Exponent: 86400.0},

	// Power
	0x28: {Unit: "W", Name: "Power", Exponent: 1.0e-3},
	0x29: {Unit: "W", Name: "Power", Exponent: 1.0e-2},
	0x2A: {Unit: "W", Name: "Power", Exponent: 1.0e-1},
	0x2B: {Unit: "W", Name: "Power", Exponent: 1.0},
	0x2C: {Unit: "W", Name: "Power", Exponent: 1.0e1},
	0x2D: {Unit: "W", Name: "Power", Exponent: 1.0e2},
	0x2E: {Unit: "W", Name: "Power", Exponent: 1.0e3},
	0x2F: {Unit: "W", Name: "Power", Exponent: 1.0e4},

	// Power
	0x30: {Unit: "J/h", Name: "Power", Exponent: 1.0e0},
	0x31: {Unit: "J/h", Name: "Power", Exponent: 1.0e1},
	0x32: {Unit: "J/h", Name: "Power", Exponent: 1.0e2},
	0x33: {Unit: "J/h", Name: "Power", Exponent: 1.0e3},
	0x34: {Unit: "J/h", Name: "Power", Exponent: 1.0e4},
	0x35: {Unit: "J/h", Name: "Power", Exponent: 1.0e5},
	0x36: {Unit: "J/h", Name: "Power", Exponent: 1.0e6},
	0x37: {Unit: "J/h", Name: "Power", Exponent: 1.0e7},

	// Volume Flow m3/h
	0x38: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-6},
	0x39: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-5},
	0x3A: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-4},
	0x3B: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-3},
	0x3C: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-2},
	0x3D: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e-1},
	0x3E: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0},
	0x3F: {Unit: "m^3/h", Name: "Volume Flow", Exponent: 1.0e1},

	// Volume Flow ext. m3/min
	0x40: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-7},
	0x41: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-6},
	0x42: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-5},
	0x43: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-4},
	0x44: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-3},
	0x45: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-2},
	0x46: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0e-1},
	0x47: {Unit: "m^3/min", Name: "Volume Flow", Exponent: 1.0},

	// Volume Flow ext. m3/s
	0x48: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-9},
	0x49: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-8},
	0x4A: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-7},
	0x4B: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-6},
	0x4C: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-5},
	0x4D: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-4},
	0x4E: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-3},
	0x4F: {Unit: "m^3/s", Name: "Volume Flow", Exponent: 1.0e-2},

	// Mass flow kg/h
	0x50: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e-3},
	0x51: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e-2},
	0x52: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e-1},
	0x53: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e-0},
	0x54: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e1},
	0x55: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e2},
	0x56: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e3},
	0x57: {Unit: "kg/h", Name: "Mass Flow", Exponent: 1.0e4},

	// Flow Temperature °C
	0x58: {Unit: "°C", Name: "Flow temperature", Exponent: 1.0e-3},
	0x59: {Unit: "°C", Name: "Flow temperature", Exponent: 1.0e-2},
	0x5A: {Unit: "°C", Name: "Flow temperature", Exponent: 1.0e-1},
	0x5B: {Unit: "°C", Name: "Flow temperature", Exponent: 1.0},

	// Return Temperature
	0x5C: {Unit: "°C", Name: "Return temperature", Exponent: 1.0e-3},
	0x5D: {Unit: "°C", Name: "Return temperature", Exponent: 1.0e-2},
	0x5E: {Unit: "°C", Name: "Return temperature", Exponent: 1.0e-1},
	0x5F: {Unit: "°C", Name: "Return temperature", Exponent: 1.0},

	// Temperature Difference K
	0x60: {Unit: "K", Name: "Temperature difference", Exponent: 1.0e-3},
	0x61: {Unit: "K", Name: "Temperature difference", Exponent: 1.0e-2},
	0x62: {Unit: "K", Name: "Temperature difference", Exponent: 1.0e-1},
	0x63: {Unit: "K", Name: "Temperature difference", Exponent: 1.0},

	// External Difference °C
	0x64: {Unit: "°C", Name: "External temperature", Exponent: 1.0e-3},
	0x65: {Unit: "°C", Name: "External temperature", Exponent: 1.0e-2},
	0x66: {Unit: "°C", Name: "External temperature", Exponent: 1.0e-1},
	0x67: {Unit: "°C", Name: "External temperature", Exponent: 1.0},

	// Pressure bar
	0x68: {Unit: "bar", Name: "Pressure", Exponent: 1.0e-3},
	0x69: {Unit: "bar", Name: "Pressure", Exponent: 1.0e-2},
	0x6A: {Unit: "bar", Name: "Pressure", Exponent: 1.0e-1},
	0x6B: {Unit: "bar", Name: "Pressure", Exponent: 1.0},

	// Time point
	0x6C: {Unit: "-", Name: "Time point (date)", Exponent: 1.0},
	0x6D: {Unit: "-", Name: "Time point (date & time)", Exponent: 1.0},

	// Units for H.C.A.
	0x6E: {Unit: "-", Name: "H.C.A.", Exponent: 1.0},

	// Reserved
	0x6F: {Unit: "-", Name: "Reserved", Exponent: 1.0},

	// Average Duration s
	0x70: {Unit: "s", Name: "Averaging Duration", Exponent: 1.0},
	0x71: {Unit: "s", Name: "Averaging Duration", Exponent: 60.0},
	0x72: {Unit: "s", Name: "Averaging Duration", Exponent: 3600.0},
	0x73: {Unit: "s", Name: "Averaging Duration", Exponent: 86400.0},

	// Actuality Duration s
	0x74: {Unit: "s", Name: "Actuality Duration", Exponent: 1.0},
	0x75: {Unit: "s", Name: "Actuality Duration", Exponent: 60.0},
	0x76: {Unit: "s", Name: "Actuality Duration", Exponent: 3600.0},
	0x77: {Unit: "s", Name: "Actuality Duration", Exponent: 86400.0},

	// Fabrication No
	0x78: {Unit: "-", Name: "Fabrication No", Exponent: 1.0},

	// (Enhanced) Identification
	0x79: {Unit: "-", Name: "(Enhanced) Identification", Exponent: 1.0},

	// Bus Address
	0x7A: {Unit: "-", Name: "Bus Address", Exponent: 1.0},

	// Any Vif
	0x7E: {Unit: "-", Name: "Any VIF", Exponent: 1.0},
	0x7F: {Unit: "-", Name: "Manufacturer specific", Exponent: 1.0},
	0xFE: {Unit: "-", Name: "Any VIF", Exponent: 1.0},
	0xFF: {Unit: "-", Name: "Manufacturer specific", Exponent: 1.0},
	0xFB: {Unit: "-", Name: "VIFE", Exponent: 1.0},
}

// VifVifeFields VIFE for other VIF then 0XFD, 0xFB
var VifVifeFields = map[byte]VIFFieldsRecord{

	0x0e:      {Unit: "", Name: "Firmware version", Exponent: 1.0},
	0b0100000: {Unit: "per second", Name: "", Exponent: 1.0},
	0b1100000: {Unit: "per second", Name: "", Exponent: 1.0},

	0b0100001: {Unit: "per minute", Name: "", Exponent: 1.0},
	0b1100001: {Unit: "per minute", Name: "", Exponent: 1.0},
}

var VifVifeFbFields = map[byte]VIFFieldsRecord{

	0x00: {Unit: "Wh", Name: "Energy", Exponent: 1.0e5},
	0x01: {Unit: "Wh", Name: "Energy", Exponent: 1.0e6},

	/* E000 100n Energy 10(n-1) GJ 0.1GJ to 1GJ */
	0x08: {Unit: "J", Name: "Energy", Exponent: 1.0e8},
	0x09: {Unit: "J", Name: "Energy", Exponent: 1.0e9},

	/* E001 000n Volume 10(n+2) m3 100m3 to 1000m3 */
	0x10: {Unit: "m^3", Name: "Volume", Exponent: 1.0e2},
	0x11: {Unit: "m^3", Name: "Volume", Exponent: 1.0e3},

	/* E001 100n Mass 10(n+2) t 100t to 1000t */
	0x18: {Unit: "kg", Name: "Mass", Exponent: 1.0e5},
	0x19: {Unit: "kg", Name: "Mass", Exponent: 1.0e6},

	/* E010 0001 Volume 0,1 feet^3 */
	0x21: {Unit: "feet^3", Name: "Volume", Exponent: 1.0e-1},

	/* E010 001n Volume 0,1-1 american gallon */
	0x22: {Unit: "American gallon", Name: "Volume", Exponent: 1.0e-1},
	0x23: {Unit: "American gallon", Name: "Volume", Exponent: 1.0},

	0x24: {Unit: "American gallon/min", Name: "Volume", Exponent: 1.0e-3},
	0x25: {Unit: "American gallon/min", Name: "Volume", Exponent: 1.0},
	0x26: {Unit: "American gallon/od", Name: "Volume", Exponent: 1.0},

	0x28: {Unit: "W", Name: "Power", Exponent: 1.0e5},
	0x29: {Unit: "W", Name: "Power", Exponent: 1.0e6},

	0x30: {Unit: "J", Name: "Power", Exponent: 1.0e8},
	0x31: {Unit: "J", Name: "Power", Exponent: 1.0e9},

	0x58: {Unit: "°F", Name: "Flow temperature", Exponent: 1.0e-3},
	0x59: {Unit: "°F", Name: "Flow temperature", Exponent: 1.0e-2},
	0x5A: {Unit: "°F", Name: "Flow temperature", Exponent: 1.0e-1},
	0x5B: {Unit: "°F", Name: "Flow temperature", Exponent: 1.0},

	0x5C: {Unit: "°F", Name: "Return temperature", Exponent: 1.0e-3},
	0x5D: {Unit: "°F", Name: "Return temperature", Exponent: 1.0e-2},
	0x5E: {Unit: "°F", Name: "Return temperature", Exponent: 1.0e-1},
	0x5F: {Unit: "°F", Name: "Return temperature", Exponent: 1.0},

	0x60: {Unit: "°F", Name: "Temperature difference", Exponent: 1.0e-3},
	0x61: {Unit: "°F", Name: "Temperature difference", Exponent: 1.0e-2},
	0x62: {Unit: "°F", Name: "Temperature difference", Exponent: 1.0e-1},
	0x63: {Unit: "°F", Name: "Temperature difference", Exponent: 1.0},

	0x64: {Unit: "°F", Name: "External temperature", Exponent: 1.0e-3},
	0x65: {Unit: "°F", Name: "External temperature", Exponent: 1.0e-2},
	0x66: {Unit: "°F", Name: "External temperature", Exponent: 1.0e-1},
	0x67: {Unit: "°F", Name: "External temperature", Exponent: 1.0},

	0x70: {Unit: "°F", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-3},
	0x71: {Unit: "°F", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-2},
	0x72: {Unit: "°F", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-1},
	0x73: {Unit: "°F", Name: "Cold / Warm Temperature Limit", Exponent: 1.0},

	0x74: {Unit: "°C", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-3},
	0x75: {Unit: "°C", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-2},
	0x76: {Unit: "°C", Name: "Cold / Warm Temperature Limit", Exponent: 1.0e-1},
	0x77: {Unit: "°C", Name: "Cold / Warm Temperature Limit", Exponent: 1.0},

	0x78: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e-3},
	0x79: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e-2},
	0x7A: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e-1},
	0x7B: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e0},
	0x7C: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e1},
	0x7D: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e2},
	0x7E: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e3},
	0x7F: {Unit: "W", Name: "Cumul count max power", Exponent: 1.0e4},

	0xFF: {Unit: "", Name: "", Exponent: 0.0},
}

var VifVifeFdFields = map[byte]VIFFieldsRecord{

	// Local Legal
	0x00: {Unit: "Currency units", Name: "Credit", Exponent: 1.0e-3},
	0x01: {Unit: "Currency units", Name: "Credit", Exponent: 1.0e-2},
	0x02: {Unit: "Currency units", Name: "Credit", Exponent: 1.0e-1},
	0x03: {Unit: "Currency units", Name: "Credit", Exponent: 1.0},

	// Debit Local legal
	0x04: {Unit: "Currency units", Name: "Debit", Exponent: 1.0e-3},
	0x05: {Unit: "Currency units", Name: "Debit", Exponent: 1.0e-2},
	0x06: {Unit: "Currency units", Name: "Debit", Exponent: 1.0e-1},
	0x07: {Unit: "Currency units", Name: "Debit", Exponent: 1.0e-0},

	0x08: {Unit: "", Name: "Access Number (transmission count", Exponent: 1.0},

	0x09: {Unit: "", Name: "Medium", Exponent: 1.0},

	0x0A: {Unit: "", Name: "Manufacturer", Exponent: 1.0},

	0x0B: {Unit: "", Name: "Parameter set identification", Exponent: 1.0},

	0x0C: {Unit: "", Name: "Model / Version", Exponent: 1.0},

	0x0D: {Unit: "", Name: "Hardware version", Exponent: 1.0},

	0x0E: {Unit: "", Name: "Firmware version", Exponent: 1.0},

	0x0F: {Unit: "", Name: "Software version", Exponent: 1.0},

	0x10: {Unit: "", Name: "Customer location", Exponent: 1.0},

	0x11: {Unit: "", Name: "Customer", Exponent: 1.0},

	0x12: {Unit: "", Name: "Access Code User", Exponent: 1.0},
	0x13: {Unit: "", Name: "Access Code Operator", Exponent: 1.0},
	0x14: {Unit: "", Name: "Access Code System Operator", Exponent: 1.0},
	0x15: {Unit: "", Name: "Access Code Developer", Exponent: 1.0},
	0x16: {Unit: "", Name: "Password", Exponent: 1.0},
	0x17: {Unit: "", Name: "Error flags", Exponent: 1.0},
	0x18: {Unit: "", Name: "Error mask", Exponent: 1.0},
	0x19: {Unit: "", Name: "Reserved", Exponent: 1.0},
	0x1A: {Unit: "", Name: "Digital Output", Exponent: 1.0},
	0x1B: {Unit: "", Name: "Digital Input", Exponent: 1.0},
	0x1C: {Unit: "Baud", Name: "Baudrate", Exponent: 1.0},
	0x1D: {Unit: "Bittimes", Name: "Response delay time", Exponent: 1.0},
	0x1F: {Unit: "", Name: "Reserved", Exponent: 1.0},
	0x20: {Unit: "", Name: "First storage # for cyclic storage", Exponent: 1.0},
	0x21: {Unit: "", Name: "Last storage # for cyclic storage", Exponent: 1.0},
	0x22: {Unit: "", Name: "Size of storage block", Exponent: 1.0},

	0x24: {Unit: "s", Name: "Storage interval", Exponent: 1.0},        // seconds
	0x25: {Unit: "s", Name: "Storage interval", Exponent: 60.0},       // minutes
	0x26: {Unit: "s", Name: "Storage interval", Exponent: 3600.0},     // hours
	0x27: {Unit: "s", Name: "Storage interval", Exponent: 86400.0},    // days
	0x28: {Unit: "s", Name: "Storage interval", Exponent: 2629743.83}, // months
	0x29: {Unit: "s", Name: "Storage interval", Exponent: 31556926.0}, // years

	/* E010 11nn Duration since last readout [sec(s)..day(s)] */
	0x2C: {Unit: "s", Name: "Duration since last readout", Exponent: 1.0},     // seconds
	0x2D: {Unit: "s", Name: "Duration since last readout", Exponent: 60.0},    // minutes
	0x2E: {Unit: "s", Name: "Duration since last readout", Exponent: 3600.0},  // hours
	0x2F: {Unit: "s", Name: "Duration since last readout", Exponent: 86400.0}, // days

	/* E011 00nn Duration of tariff (nn=01 ..11: min to days) */
	0x31: {Unit: "s", Name: "Duration since last readout", Exponent: 60.0},    // minutes
	0x32: {Unit: "s", Name: "Duration since last readout", Exponent: 3600.0},  // hours
	0x33: {Unit: "s", Name: "Duration since last readout", Exponent: 86400.0}, // days

	/* E011 01nn Period of tariff [sec(s) to day(s)]  */
	0x34: {Unit: "s", Name: "Period of tariff", Exponent: 1.0},        // seconds
	0x35: {Unit: "s", Name: "Period of tariff", Exponent: 60.0},       // minutes
	0x36: {Unit: "s", Name: "Period of tariff", Exponent: 3600.0},     // hours
	0x37: {Unit: "s", Name: "Period of tariff", Exponent: 86400.0},    // days
	0x38: {Unit: "s", Name: "Period of tariff", Exponent: 2629743.83}, // months
	0x39: {Unit: "s", Name: "Period of tariff", Exponent: 31556926.0}, // years

	/* E011 1010 dimensionless / no VIF */
	0x3A: {Unit: "", Name: "Dimensionless", Exponent: 1.0e0}, // years

	/* E100 nnnn   Volts electrical units */
	0x40: {Unit: "V", Name: "Voltage", Exponent: 1.0e-9},
	0x41: {Unit: "V", Name: "Voltage", Exponent: 1.0e-8},
	0x42: {Unit: "V", Name: "Voltage", Exponent: 1.0e-7},
	0x43: {Unit: "V", Name: "Voltage", Exponent: 1.0e-6},
	0x44: {Unit: "V", Name: "Voltage", Exponent: 1.0e-5},
	0x45: {Unit: "V", Name: "Voltage", Exponent: 1.0e-4},
	0x46: {Unit: "V", Name: "Voltage", Exponent: 1.0e-4},
	0x47: {Unit: "V", Name: "Voltage", Exponent: 1.0e-2},
	0x48: {Unit: "V", Name: "Voltage", Exponent: 1.0e-1},
	0x49: {Unit: "V", Name: "Voltage", Exponent: 1.0e0},
	0x4A: {Unit: "V", Name: "Voltage", Exponent: 1.0e+1},
	0x4B: {Unit: "V", Name: "Voltage", Exponent: 1.0e+2},
	0x4C: {Unit: "V", Name: "Voltage", Exponent: 1.0e+3},
	0x4D: {Unit: "V", Name: "Voltage", Exponent: 1.0e+4},
	0x4E: {Unit: "V", Name: "Voltage", Exponent: 1.0e+5},
	0x4F: {Unit: "V", Name: "Voltage", Exponent: 1.0e+6},

	/* E101 nnnn   A */
	0x50: {Unit: "A", Name: "Current", Exponent: 1.0e-12},
	0x51: {Unit: "A", Name: "Current", Exponent: 1.0e-11},
	0x52: {Unit: "A", Name: "Current", Exponent: 1.0e-10},
	0x53: {Unit: "A", Name: "Current", Exponent: 1.0e-9},
	0x54: {Unit: "A", Name: "Current", Exponent: 1.0e-8},
	0x55: {Unit: "A", Name: "Current", Exponent: 1.0e-7},
	0x56: {Unit: "A", Name: "Current", Exponent: 1.0e-6},
	0x57: {Unit: "A", Name: "Current", Exponent: 1.0e-5},
	0x58: {Unit: "A", Name: "Current", Exponent: 1.0e-4},
	0x59: {Unit: "A", Name: "Current", Exponent: 1.0e-3},
	0x5A: {Unit: "A", Name: "Current", Exponent: 1.0e-2},
	0x5B: {Unit: "A", Name: "Current", Exponent: 1.0e-1},
	0x5C: {Unit: "A", Name: "Current", Exponent: 1.0e-0},
	0x5D: {Unit: "A", Name: "Current", Exponent: 1.0e+1},
	0x5E: {Unit: "A", Name: "Current", Exponent: 1.0e+2},
	0x5F: {Unit: "A", Name: "Current", Exponent: 1.0e+3},

	0x60: {Unit: "", Name: "Reset counter", Exponent: 1.0},
	0x61: {Unit: "", Name: "Cumulation counter", Exponent: 1.0},
	0x62: {Unit: "", Name: "Control signal", Exponent: 1.0},
	0x63: {Unit: "", Name: "Day of wee", Exponent: 1.0},
	0x64: {Unit: "", Name: "Week number", Exponent: 1.0},
	0x65: {Unit: "", Name: "Time point of day changer", Exponent: 1.0},
	0x66: {Unit: "", Name: "State of parameter activation", Exponent: 1.0},
	0x67: {Unit: "", Name: "Special supplier information", Exponent: 1.0},

	0x68: {Unit: "hours", Name: "Duration since last cumulation", Exponent: 1.0},
	0x69: {Unit: "days", Name: "Duration since last cumulation", Exponent: 1.0},
	0x6A: {Unit: "months", Name: "Duration since last cumulation", Exponent: 1.0},
	0x6B: {Unit: "years", Name: "Duration since last cumulation", Exponent: 1.0},

	0x6C: {Unit: "hours", Name: "Operating time battery [hours]", Exponent: 1.0},
	0x6D: {Unit: "days", Name: "Operating time battery [days]", Exponent: 1.0},
	0x6E: {Unit: "months", Name: "Operating time battery [months]", Exponent: 1.0},
	0x6F: {Unit: "years", Name: "Operating time battery [years]", Exponent: 1.0},

	0x70: {Unit: "", Name: "Date and time of battery change", Exponent: 1.0},
}
