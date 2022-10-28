var Client = require('ftp');
var fs = require('fs');
var process = require('process')
var c = new Client();

let file = 'main.txt'

c.on('ready', function () {
    c.get(file, function (err, stream) {
        if (err) throw err;
        stream.pipe(process.stdout)
        c.logout((err) => {
            if (err) throw err;
            c.end();
        })
    });
});


c.connect({ "host": "localhost", "port": 2023, "user": "hkhan", "password": "password" });
