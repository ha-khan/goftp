```
7.  TYPICAL FTP SCENARIO

   User at host U wanting to transfer files to/from host S:

   In general, the user will communicate to the server via a mediating
   user-FTP process.  The following may be a typical scenario.  The
   user-FTP prompts are shown in parentheses, '---->' represents
   commands from host U to host S, and '<----' represents replies from
   host S to host U.

      LOCAL COMMANDS BY USER              ACTION INVOLVED

      ftp (host) multics<CR>         Connect to host S, port L,
                                     establishing control connections.
                                     <---- 220 Service ready <CRLF>.
      username Doe <CR>              USER Doe<CRLF>---->
                                     <---- 331 User name ok,
                                               need password<CRLF>.
      password mumble <CR>           PASS mumble<CRLF>---->
                                     <---- 230 User logged in<CRLF>.
      retrieve (local type) ASCII<CR>
      (local pathname) test 1 <CR>   User-FTP opens local file in ASCII.
      (for. pathname) test.pl1<CR>   RETR test.pl1<CRLF> ---->
                                     <---- 150 File status okay;
                                           about to open data
                                           connection<CRLF>.
                                     Server makes data connection
                                     to port U.

                                     <---- 226 Closing data connection,
                                         file transfer successful<CRLF>.
      type Image<CR>                 TYPE I<CRLF> ---->
                                     <---- 200 Command OK<CRLF>
      store (local type) image<CR>
      (local pathname) file dump<CR> User-FTP opens local file in Image.
      (for.pathname) >udd>cn>fd<CR>  STOR >udd>cn>fd<CRLF> ---->
                                     <---- 550 Access denied<CRLF>
      terminate                      QUIT <CRLF> ---->
                                     Server closes all
                                     connections
```
