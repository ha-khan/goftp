from ftplib import FTP

if __name__ == '__main__':
    with FTP() as client:
        client.connect('localhost', 2023)
