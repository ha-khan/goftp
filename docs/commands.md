# Overview

```

   5.3.  COMMANDS

      The commands are Telnet character strings transmitted over the
      control connections as described in the Section on FTP Commands.
      The command functions and semantics are described in the Section
      on Access Control Commands, Transfer Parameter Commands, FTP
      Service Commands, and Miscellaneous Commands.  The command syntax
      is specified here.

      The commands begin with a command code followed by an argument
      field.  The command codes are four or fewer alphabetic characters.
      Upper and lower case alphabetic characters are to be treated
      identically.  Thus, any of the following may represent the
      retrieve command:
                  RETR    Retr    retr    ReTr    rETr

      This also applies to any symbols representing parameter values,
      such as A or a for ASCII TYPE.  The command codes and the argument
      fields are separated by one or more spaces.

      The argument field consists of a variable length character string
      ending with the character sequence <CRLF> (Carriage Return, Line
      Feed) for NVT-ASCII representation; for other negotiated languages
      a different end of line character might be used.  It should be
      noted that the server is to take no action until the end of line
      code is received.

      The syntax is specified below in NVT-ASCII.  All characters in the
      argument field are ASCII characters including any ASCII
      represented decimal integers.  Square brackets denote an optional
      argument field.  If the option is not taken, the appropriate
      default is implied.
      5.3.1.  FTP COMMANDS

         The following are the FTP commands:

            USER <SP> <username> <CRLF>
            PASS <SP> <password> <CRLF>
            ACCT <SP> <account-information> <CRLF>
            CWD  <SP> <pathname> <CRLF>
            CDUP <CRLF>
            SMNT <SP> <pathname> <CRLF>
            QUIT <CRLF>
            REIN <CRLF>
            PORT <SP> <host-port> <CRLF>
            PASV <CRLF>
            TYPE <SP> <type-code> <CRLF>
            STRU <SP> <structure-code> <CRLF>
            MODE <SP> <mode-code> <CRLF>
            RETR <SP> <pathname> <CRLF>
            STOR <SP> <pathname> <CRLF>
            STOU <CRLF>
            APPE <SP> <pathname> <CRLF>
            ALLO <SP> <decimal-integer>
                [<SP> R <SP> <decimal-integer>] <CRLF>
            REST <SP> <marker> <CRLF>
            RNFR <SP> <pathname> <CRLF>
            RNTO <SP> <pathname> <CRLF>
            ABOR <CRLF>
            DELE <SP> <pathname> <CRLF>
            RMD  <SP> <pathname> <CRLF>
            MKD  <SP> <pathname> <CRLF>
            PWD  <CRLF>
            LIST [<SP> <pathname>] <CRLF>
            NLST [<SP> <pathname>] <CRLF>
            SITE <SP> <string> <CRLF>
            SYST <CRLF>
            STAT [<SP> <pathname>] <CRLF>
            HELP [<SP> <string>] <CRLF>
            NOOP <CRLF>
```
