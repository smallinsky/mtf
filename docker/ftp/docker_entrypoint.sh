#!/bin/sh -x

addgroup -S $FTP_USER
adduser -D -G $FTP_USER -h /ftp/ -s /bin/false  $FTP_USER

echo "$FTP_USER:$FTP_PASS" | /usr/sbin/chpasswd
chown $FTP_USER:$FTP_USER /ftp/ -R

/go/bin/fswatch --dir /ftp --addr $DOCKER_HOST_ADDR:4441 &
/usr/sbin/vsftpd /etc/vsftpd/vsftpd.conf
