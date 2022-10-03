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
      is when the user-DTP is sending the data in a transfer mode that
      requires the connection to be closed to indicate EOF.  The server
      MUST close the data connection under the following conditions:

         1. The server has completed sending data in a transfer mode
            that requires a close to indicate EOF.

         2. The server receives an ABORT command from the user.

         3. The port specification is changed by a command from the
            user.

         4. The control connection is closed legally or otherwise.

         5. An irrecoverable error condition occurs.

      Otherwise the close is a server option, the exercise of which the
      server must indicate to the user-process by either a 250 or 226
      reply only.


   3.3.  DATA CONNECTION MANAGEMENT

      Default Data Connection Ports:  All FTP implementations must
      support use of the default data connection ports, and only the
      User-PI may initiate the use of non-default ports.

      Negotiating Non-Default Data Ports:   The User-PI may specify a
      non-default user side data port with the PORT command.  The
      User-PI may request the server side to identify a non-default
      server side data port with the PASV command.  Since a connection
      is defined by the pair of addresses, either of these actions is
      enough to get a different data connection, still it is permitted
      to do both commands to use new ports on both ends of the data
      connection.

      Reuse of the Data Connection:  When using the stream mode of data
      transfer the end of the file must be indicated by closing the
      connection.  This causes a problem if multiple files are to be
      transfered in the session, due to need for TCP to hold the
      connection record for a time out period to guarantee the reliable
      communication.  Thus the connection can not be reopened at once.

         There are two solutions to this problem.  The first is to
         negotiate a non-default port.  The second is to use another
         transfer mode.

         A comment on transfer modes.  The stream transfer mode is
         inherently unreliable, since one can not determine if the
         connection closed prematurely or not.  The other transfer modes
         (Block, Compressed) do not close the connection to indicate the
         end of file.  They have enough FTP encoding that the data
         connection can be parsed to determine the end of the file.
         Thus using these modes one can leave the data connection open
         for multiple file transfers.


```
