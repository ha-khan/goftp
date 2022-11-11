require 'net/ftp'


Net::FTP.open('localhost', {:port=>2023}) do |ftp|
    ftp.login('hkhan', 'password')

    ftp.retrlines('RETR main.txt') do |data|
        puts data
    end

    ftp.quit()
end
