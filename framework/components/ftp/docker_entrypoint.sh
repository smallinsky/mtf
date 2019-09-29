#!/bin/sh -x

addgroup -S $FTP_USER
adduser -D -G $FTP_USER -h /ftp/ -s /bin/false  $FTP_USER

echo "$FTP_USER:$FTP_PASS" | /usr/sbin/chpasswd
chown $FTP_USER:$FTP_USER /ftp/ -R

/usr/sbin/vsftpd /etc/vsftpd/vsftpd.conf
