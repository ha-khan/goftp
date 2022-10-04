from ftplib import FTP
import io


if __name__ == '__main__':
    with FTP() as client:
        client.connect('localhost', 2023)
        client.login('hkhan', 'password')
        fp = io.BytesIO(bytes(b'hello world'))
        client.storlines("STOR main.txt", fp)
        client.retrlines('RETR main.txt')
        with open('main.txt', 'wb') as fd:
            client.retrbinary('RETR main.txt', fd.write)
