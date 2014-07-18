Clone of Go-Guerrilla SMTPd without Redis or MySQL.


Go-Guerrilla SMTPd
====================

An minimalist SMTP server written in Go, made for receiving large volumes of mail.

### What is Go Guerrilla SMTPd?

It's a small SMTP server written in Go, for the purpose of receiving large volume of email.
Written for GuerrillaMail.com which processes tens of thousands of emails
every hour.

The purpose of this daemon is to grab the email and disconnect as quickly as possible.

A typical user of this software would probably want to customize the saveMail function for
their own systems.

This server does not attempt to filter HTML, check for spam or do any sender 
verification. These steps should be performed by other programs.
The server does NOT send any email including bounces. This should
be performed by a separate program.


### History and purpose

GoGuerrilla is a port of the original 'Guerrilla' SMTP daemon written in PHP using
an event-driven I/O library (libevent)

https://github.com/flashmob/Guerrilla-SMTPd

It's not a direct port, although the purpose and functionality remains identical.


Getting started
===========================

Copy goguerrilla.conf.sample to goguerrilla.conf


Configuration
============================================
The configuration is in strict JSON format. Here is an annotated configuration.
Copy goguerrilla.conf.sample to goguerrilla.conf


	{
	    "GM_ALLOWED_HOSTS":"example.com,sample.com,foo.com,bar.com", // which domains accept mail
	    "GM_MAIL_TABLE":"new_mail", // name of new email table
	    "GM_PRIMARY_MAIL_HOST":"mail.example.com", // given in the SMTP greeting
	    "GSMTP_HOST_NAME":"mail.example.com", // given in the SMTP greeting
	    "GSMTP_LOG_FILE":"/dev/stdout", // not used yet
	    "GSMTP_MAX_SIZE":"131072", // max size of DATA command
	    "GSMTP_PRV_KEY":"/etc/ssl/private/example.com.key", // private key for TLS
	    "GSMTP_PUB_KEY":"/etc/ssl/certs/example.com.crt", // public key for TLS
	    "GSMTP_TIMEOUT":"100", // tcp connection timeout
	    "GSMTP_VERBOSE":"N", // set to Y for debugging
	    "GSTMP_LISTEN_INTERFACE":"5.9.7.183:25",
	    "MYSQL_DB":"gmail_mail", // database name
	    "MYSQL_HOST":"127.0.0.1:3306", // database connect
	    "MYSQL_PASS":"$ecure1t", // database connection pass
	    "MYSQL_USER":"gmail_mail", // database username
	    "GM_MAX_CLIENTS":"500", // max clients that can be handled
		"NGINX_AUTH_ENABLED":"N",// Y or N
		"NGINX_AUTH":"127.0.0.1:8025", // If using Nginx proxy, choose an ip and port to serve Auth requsts for Nginx
	    "SGID":"508",// group id of the user from /etc/passwd
		"GUID":"504" // uid from /etc/passwd
	}

Using Nginx as a proxy
=========================================================
Nginx can be used to proxy SMTP traffic for GoGuerrilla SMTPd

Why proxy SMTP?

 *	Terminate TLS connections: In Nov 2012 when this was written, Golang was
not all there yet when it comes to TLS. The situation is better now but perhaps
not comprehensively so. See [1][2] for current status. 
OpenSSL on the other hand, used in Nginx, has a complete implementation 
of SSL v2/v3 and TLS protocols.

[1] https://code.google.com/p/go/issues/detail?id=5742
[2] https://groups.google.com/forum/#!topic/golang-nuts/LjhVww0TQi4

 *	Could be used for load balancing and authentication in the future.


 a.	Compile nginx with --with-mail --with-mail_ssl_module

 b.	Configuration:

	
		mail {
	        auth_http 127.0.0.1:8025/; # This is the URL to GoGuerrilla's http service which tells Nginx where to proxy the traffic to 								
	        server {
	                listen  15.29.8.163:25;
	                protocol smtp;
	                server_name  ak47.example.com;
	
	                smtp_auth none;
	                timeout 30000;
					smtp_capabilities "SIZE 15728640";
					
					# ssl default off. Leave off if starttls is on
	                #ssl                  on;
	                ssl_certificate      /etc/ssl/certs/ssl-cert-snakeoil.pem;
	                ssl_certificate_key  /etc/ssl/private/ssl-cert-snakeoil.key;
	                ssl_session_timeout  5m;
	                ssl_protocols  SSLv2 SSLv3 TLSv1;
	                ssl_ciphers  HIGH:!aNULL:!MD5;
	                ssl_prefer_server_ciphers   on;
					# TLS off unless client issues STARTTLS command
	                starttls on;
	                proxy on;
	        }
		}
	
			
Assuming that Guerrilla SMTPd has the following configuration settings:

	"GSMTP_MAX_SIZE"		  "15728640",
	"NGINX_AUTH_ENABLED":     "Y",
	"NGINX_AUTH":             "127.0.0.1:8025", 


Starting / Command Line usage
==========================================================

All command line arguments are optional

	-config="goguerrilla.conf": Path to the configuration file
	 -if="": Interface and port to listen on, eg. 127.0.0.1:2525
	 -v="n": Verbose, [y | n]

Starting from the command line (example)

	/usr/bin/nohup /home/mike/goguerrilla -config=/home/mike/goguerrilla.conf 2>&1 &

This will place goguerrilla in the background and continue running

You may also put another process to watch your goguerrilla process and re-start it
if something goes wrong.

License
=======

Copyright (c) 2012 Flashmob, GuerrillaMail.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
