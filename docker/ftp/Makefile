build:
	docker build -t ftpserver .

run:
	docker run --rm -e FTP_USER=test -e FTP_PASS=test -it --name vsftpd -p 20:20 -p 21:21 -p 21100-21110:21100-21110 ftpserver
