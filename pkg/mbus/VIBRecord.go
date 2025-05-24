package mbus

// VIBRecord is a struct that represents a complete M-Bus Value Information Block Record.
// DIF - Data Information Field
// DIFE - Data Information Field Extension
// VIF - Value Information Field
// VIFE - Value Information Field Extension
// Data - Data
type VIBRecord struct {
	DIF  DIFField
	DIFE []DIFEField
	VIF  VIFField
	VIFE []VIFEField
	Data string
}
