```
   5.4.  SEQUENCING OF COMMANDS AND REPLIES

      The communication between the user and server is intended to be an
      alternating dialogue.  As such, the user issues an FTP command and
      the server responds with a prompt primary reply.  The user should
      wait for this initial primary success or failure response before
      sending further commands.

      Certain commands require a second reply for which the user should
      also wait.  These replies may, for example, report on the progress
      or completion of file transfer or the closing of the data
      connection.  They are secondary replies to file transfer commands.

      One important group of informational replies is the connection
      greetings.  Under normal circumstances, a server will send a 220
      reply, "awaiting input", when the connection is completed.  The
      user should wait for this greeting message before sending any
      commands.  If the server is unable to accept input right away, a
      120 "expected delay" reply should be sent immediately and a 220
      reply when ready.  The user will then know not to hang up if there
      is a delay.

      Spontaneous Replies

         Sometimes "the system" spontaneously has a message to be sent
         to a user (usually all users).  For example, "System going down
         in 15 minutes".  There is no provision in FTP for such
         spontaneous information to be sent from the server to the user.
         It is recommended that such information be queued in the
         server-PI and delivered to the user-PI in the next reply
         (possibly making it a multi-line reply).

      The table below lists alternative success and failure replies for
      each command.  These must be strictly adhered to; a server may
      substitute text in the replies, but the meaning and action implied
      by the code numbers and by the specific command reply sequence
      cannot be altered.

      Command-Reply Sequences

         In this section, the command-reply sequence is presented.  Each
         command is listed with its possible replies; command groups are
         listed together.  Preliminary replies are listed first (with
         their succeeding replies indented and under them), then
         positive and negative completion, and finally intermediary
         replies with the remaining commands from the sequence
         following.  This listing forms the basis for the state
         diagrams, which will be presented separately.

            Connection Establishment
               120
                  220
               220
               421
            Login
               USER
                  230
                  530
                  500, 501, 421
                  331, 332
               PASS
                  230
                  202
                  530
                  500, 501, 503, 421
                  332
               ACCT
                  230
                  202
                  530
                  500, 501, 503, 421
               CWD
                  250
                  500, 501, 502, 421, 530, 550
               CDUP
                  200
                  500, 501, 502, 421, 530, 550
               SMNT
                  202, 250
                  500, 501, 502, 421, 530, 550
            Logout
               REIN
                  120
                     220
                  220
                  421
                  500, 502
               QUIT
                  221
                  500

            Transfer parameters
               PORT
                  200
                  500, 501, 421, 530
               PASV
                  227
                  500, 501, 502, 421, 530
               MODE
                  200
                  500, 501, 504, 421, 530
               TYPE
                  200
                  500, 501, 504, 421, 530
               STRU
                  200
                  500, 501, 504, 421, 530
            File action commands
               ALLO
                  200
                  202
                  500, 501, 504, 421, 530
               REST
                  500, 501, 502, 421, 530
                  350
               STOR
                  125, 150
                     (110)
                     226, 250
                     425, 426, 451, 551, 552
                  532, 450, 452, 553
                  500, 501, 421, 530
               STOU
                  125, 150
                     (110)
                     226, 250
                     425, 426, 451, 551, 552
                  532, 450, 452, 553
                  500, 501, 421, 530
               RETR
                  125, 150
                     (110)
                     226, 250
                     425, 426, 451
                  450, 550
                  500, 501, 421, 530
               LIST
                  125, 150
                     226, 250
                     425, 426, 451
                  450
                  500, 501, 502, 421, 530
               NLST
                  125, 150
                     226, 250
                     425, 426, 451
                  450
                  500, 501, 502, 421, 530
               APPE
                  125, 150
                     (110)
                     226, 250
                     425, 426, 451, 551, 552
                  532, 450, 550, 452, 553
                  500, 501, 502, 421, 530
               RNFR
                  450, 550
                  500, 501, 502, 421, 530
                  350
               RNTO
                  250
                  532, 553
                  500, 501, 502, 503, 421, 530
               DELE
                  250
                  450, 550
                  500, 501, 502, 421, 530
               RMD
                  250
                  500, 501, 502, 421, 530, 550
               MKD
                  257
                  500, 501, 502, 421, 530, 550
               PWD
                  257
                  500, 501, 502, 421, 550
               ABOR
                  225, 226
                  500, 501, 502, 421
            Informational commands
               SYST
                  215
                  500, 501, 502, 421
               STAT
                  211, 212, 213
                  450
                  500, 501, 502, 421, 530
               HELP
                  211, 214
                  500, 501, 502, 421
            Miscellaneous commands
               SITE
                  200
                  202
                  500, 501, 530
               NOOP
                  200
                  500 421
```
