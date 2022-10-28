require 'net/ftp'

ftp = Net::FTP.new('localhost', {:port=>2023})
ftp.login('hkhan', 'password')
ftp.retrlines('RETR main.txt') do |data| 
    puts data
end
ftp.quit()
ftp.close()
