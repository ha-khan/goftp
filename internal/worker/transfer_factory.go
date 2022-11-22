package worker

import (
	"io"
	"os"
)

// Transfer Parameters, only accepting a subset from spec
type TransferFactory struct {
	// MODE command specifies how the bits of the data are to be transmitted
	// S - Stream
	Mode rune
	//
	//
	// STRUcture and TYPE commands, are used to define the way in which the data are to be represented.
	//
	// F - File (no structure, file is considered to be a sequence of data bytes)
	// R - Record (must be accepted for "text" files (ASCII) )
	Structure rune
	//
	//
	// A - ASCII (primarily for the transfer of text files <CRLF> used to denote end of text line)
	// I - Image (data is sent as contiguous bits, which  are packed into 8-bit transfer bytes)
	Type rune
	//
	//
	//
	//
	PWD string
}

func (t *TransferFactory) Create(fd *os.File) (io.ReadWriteCloser, error) {
	// TODO: use fields Mode, Type, Structure to
	// generate specific to those params ReadWriter
	// and return that to caller
	// need to change interface to io.ReadWriter
	// can return scanner or bufio, or create a custom one for
	// compressed mode, ..etc
	return fd, nil
}

func (t *TransferFactory) SetPWD(pwd string) {
	t.PWD = pwd
}

func (t *TransferFactory) GetPWD() string {
	return t.PWD
}

func (t *TransferFactory) SetType(ty rune) {
	t.Type = ty
}

func (t *TransferFactory) GetType() rune {
	return t.Type
}

func (t *TransferFactory) SetStructure(stru rune) {
	t.Structure = stru
}

func (t *TransferFactory) GetStructure() rune {
	return t.Structure
}

func (t *TransferFactory) SetMode(mo rune) {
	t.Mode = mo
}

func (t *TransferFactory) GetMode() rune {
	return t.Mode
}

func NewDefaultTransferFactory() *TransferFactory {
	return &TransferFactory{
		Mode:      'S', // Stream
		Structure: 'F', // File
		Type:      'A', // ASCII
		PWD:       "/temp",
	}
}
