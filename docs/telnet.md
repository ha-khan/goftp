# Overview

Layer 7 protocol over TCP, meant to comm with some remote process and issue commands and such via a terminal interface. 

RFC 959 references telnet being utilized during the control connection phase where FTP clients can issue cmds that setup the state of the connection (PWD, User Name / Password, ... etc)

1. Network Virtual Terminal
	1. Abstract Machine that interprets telnet protocol and correlates them to host so that characters/symbols can be rendered ...etc
	2. Provides base level of valid inputs with expected outputs.
2. Principle of Negotiated Options
	1. Within Telnet Protocol are various "options" that can be used with "DO, DON't, WILL, WON't, ..etc structure.
	2. Allows for creation of "custom" commands in addition to ones provided in NVT (change char set, echo mode, ... etc)
3. Symmetry of Negotiation
	1. Some set of standards (grammar?) that the handshake and protocol are expected to follow.
