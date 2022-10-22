package worker

// Transfer Parameters, only accepting a subset from spec
type transfer struct {
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
}
