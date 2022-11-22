from ftplib import FTP
import io


if __name__ == '__main__':
    with FTP() as client:
        client.connect('localhost', 2023)
        client.login('hkhan', 'password')
        client.voidcmd("NOOP")
        with io.BytesIO(bytes(b'Go is fun')) as fp:
            client.storlines("STOR main.txt", fp)
        # client.retrlines('RETR main.txt')
        # with open('main.txt', 'wb') as fd:
        #     client.retrbinary('RETR main.txt', fd.write)
        client.set_pasv(False)
        client.retrlines('RETR hello.txt')
        # with io.BytesIO(bytes(b'hello world')) as fp:
        #     client.storlines("STOR main.txt", fp)
