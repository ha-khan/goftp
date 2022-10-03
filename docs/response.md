```
3.2.  ESTABLISHING DATA CONNECTIONS

      The mechanics of transferring data consists of setting up the data
      connection to the appropriate ports and choosing the parameters
      for transfer.  Both the user and the server-DTPs have a default
      data port.  The user-process default data port is the same as the
      control connection port (i.e., U).  The server-process default
      data port is the port adjacent to the control connection port
      (i.e., L-1).

      The transfer byte size is 8-bit bytes.  This byte size is relevant
      only for the actual transfer of the data; it has no bearing on
      representation of the data within a host's file system.

      The passive data transfer process (this may be a user-DTP or a
      second server-DTP) shall "listen" on the data port prior to
      sending a transfer request command.  The FTP request command
      determines the direction of the data transfer.  The server, upon
      receiving the transfer request, will initiate the data connection
      to the port.  When the connection is established, the data
      transfer begins between DTP's, and the server-PI sends a
      confirming reply to the user-PI.

      Every FTP implementation must support the use of the default data
      ports, and only the USER-PI can initiate a change to non-default
      ports.

      It is possible for the user to specify an alternate data port by
      use of the PORT command.  The user may want a file dumped on a TAC
      line printer or retrieved from a third party host.  In the latter
      case, the user-PI sets up control connections with both
      server-PI's.  One server is then told (by an FTP command) to
      "listen" for a connection which the other will initiate.  The
      user-PI sends one server-PI a PORT command indicating the data
      port of the other.  Finally, both are sent the appropriate
      transfer commands.  The exact sequence of commands and replies
      sent between the user-controller and the servers is defined in the
      Section on FTP Replies.

      In general, it is the server's responsibility to maintain the data
      connection--to initiate it and to close it.  The exception to this


```
